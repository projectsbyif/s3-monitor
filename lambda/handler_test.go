package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/trillian"
	"github.com/google/trillian/server"
	"github.com/google/trillian/storage"

	stestonly "github.com/google/trillian/storage/testonly"
)

var (
	treeID int64
)

func loadS3Event(t *testing.T, fileName string) events.S3Event {
	var inputEvent events.S3Event
	s, _ := ioutil.ReadFile(fileName)
	if err := json.Unmarshal([]byte(s), &inputEvent); err != nil {
		t.Errorf("could not unmarshal event. details: %v", err)
	}
	return inputEvent
}

func numberOfLeaves(tx storage.LogMetadata) int64 {
	trees, _ := tx.GetUnsequencedCounts(context.TODO())
	return trees[treeID]
}

func assertLeavesAdded(t *testing.T, tx storage.LogMetadata, expectedNumberOfNewLeaves int64, f func()) {
	countBefore := numberOfLeaves(tx)
	f()
	countAfter := numberOfLeaves(tx)
	if countAfter-countBefore != expectedNumberOfNewLeaves {
		t.Errorf("Expected %v leaf in tree, found %v", expectedNumberOfNewLeaves, countAfter-countBefore)
	}
}

func TestHandleRequest(t *testing.T) {
	// Setup code
	ctx := context.Background()
	sp, _ := server.NewStorageProvider("memory", nil)
	var tree *trillian.Tree
	sp.AdminStorage().ReadWriteTransaction(ctx, func(ctx context.Context, tx storage.AdminTX) error {
		tree, _ = tx.CreateTree(ctx, stestonly.LogTree)
		tx.Commit()
		return nil
	})
	treeID = tree.TreeId
	logServer, _ := StartTrillian(ctx, sp, treeID)
	handleRequest := CreateHandler(logServer, treeID)
	tx, _ := sp.LogStorage().Snapshot(ctx)
	defer tx.Close()

	t.Run("one event", func(t *testing.T) {
		assertLeavesAdded(t, tx, 1, func() {
			inputEvent := loadS3Event(t, "testdata/oneEvent.json")
			handleRequest(ctx, inputEvent)
		})
	})

	t.Run("no events", func(t *testing.T) {
		assertLeavesAdded(t, tx, 0, func() {
			inputEvent := loadS3Event(t, "testdata/noEvent.json")
			handleRequest(ctx, inputEvent)
		})
	})

	t.Run("multiple events", func(t *testing.T) {
		assertLeavesAdded(t, tx, 5, func() {
			inputEvent := loadS3Event(t, "testdata/fiveEvents.json")
			handleRequest(ctx, inputEvent)
		})
	})
}
