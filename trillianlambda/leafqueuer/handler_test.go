package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/golang/mock/gomock"
	"github.com/google/trillian"
	"github.com/google/trillian/extension"
	"github.com/google/trillian/server"
	"github.com/google/trillian/storage"
	"github.com/google/trillian/util/clock"
	"github.com/projectsbyif/verifiable-cloudtrail/trillianlambda/leafqueuer/testonly"

	stestonly "github.com/google/trillian/storage/testonly"
)

var (
	tree *trillian.Tree
)

func loadS3Event(t *testing.T, fileName string) events.S3Event {
	var inputEvent events.S3Event
	s, _ := ioutil.ReadFile(fileName)
	if err := json.Unmarshal([]byte(s), &inputEvent); err != nil {
		t.Errorf("could not unmarshal event. details: %v", err)
	}
	return inputEvent
}

func numberOfLeaves(ls storage.LogStorage) int64 {
	tx, _ := ls.Snapshot(context.TODO())
	defer tx.Close()
	trees, _ := tx.GetUnsequencedCounts(context.TODO())
	return trees[tree.TreeId]
}

func assertLeavesAdded(t *testing.T, ls storage.LogStorage, expectedNumberOfNewLeaves int64, f func()) {
	countBefore := numberOfLeaves(ls)
	f()
	countAfter := numberOfLeaves(ls)
	if countAfter-countBefore != expectedNumberOfNewLeaves {
		t.Errorf("Expected %v leaf in tree, found %v", expectedNumberOfNewLeaves, countAfter-countBefore)
	}
}

func TestHandlerTrillianIntegration(t *testing.T) {
	// Setup code
	ctx := context.Background()
	sp, _ := server.NewStorageProvider("memory", nil)
	sp.AdminStorage().ReadWriteTransaction(ctx, func(ctx context.Context, tx storage.AdminTX) error {
		tree, _ = tx.CreateTree(ctx, stestonly.LogTree)
		tx.Commit()
		return nil
	})
	registry := extension.Registry{
		AdminStorage: sp.AdminStorage(),
		LogStorage:   sp.LogStorage(),
	}
	timeSource := clock.System
	logServer := *server.NewTrillianLogRPCServer(registry, timeSource)
	logServer.InitLog(ctx, &trillian.InitLogRequest{LogId: tree.TreeId})

	t.Run("one event", func(t *testing.T) {
		assertLeavesAdded(t, sp.LogStorage(), 1, func() {
			inputEvent := loadS3Event(t, "testdata/oneEvent.json")
			ProcessEvents(ctx, inputEvent, &logServer, tree.TreeId)
		})
	})

	t.Run("no events", func(t *testing.T) {
		assertLeavesAdded(t, sp.LogStorage(), 0, func() {
			inputEvent := loadS3Event(t, "testdata/noEvent.json")
			ProcessEvents(ctx, inputEvent, &logServer, tree.TreeId)
		})
	})

	t.Run("multiple events", func(t *testing.T) {
		assertLeavesAdded(t, sp.LogStorage(), 5, func() {
			inputEvent := loadS3Event(t, "testdata/fiveEvents.json")
			ProcessEvents(ctx, inputEvent, &logServer, tree.TreeId)
		})
	})
}

func TestHandler_OnlyQueuesLeavesWhenEventsExist(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockObj := testonly.NewMockLeafQueuer(mockCtrl)
	inputEvent := loadS3Event(t, "testdata/noEvent.json")
	mockObj.EXPECT().QueueLeaves(gomock.Any(), gomock.Any()).Times(0)
	ProcessEvents(context.TODO(), inputEvent, mockObj, 0)
}
