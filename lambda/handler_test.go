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

func TestHandleRequest(t *testing.T) {
	s, _ := ioutil.ReadFile("testdata/event.json")
	ctx := context.Background()
	var inputEvent events.S3Event
	if err := json.Unmarshal([]byte(s), &inputEvent); err != nil {
		t.Errorf("could not unmarshal event. details: %v", err)
	}
	sp, _ := server.NewStorageProvider("memory", nil)
	var tree *trillian.Tree
	sp.AdminStorage().ReadWriteTransaction(ctx, func(ctx context.Context, tx storage.AdminTX) error {
		tree, _ = tx.CreateTree(ctx, stestonly.LogTree)
		tx.Commit()
		return nil
	})
	treeID := tree.TreeId
	logServer, _ := StartTrillian(ctx, sp, treeID)
	handleRequest := CreateHandler(logServer, treeID)
	handleRequest(ctx, inputEvent)
	tx, _ := sp.LogStorage().Snapshot(ctx)
	defer tx.Close()
	trees, _ := tx.GetUnsequencedCounts(ctx)
	if trees[treeID] != 1 {
		t.Errorf("Expected 1 leaf in tree, found %v", trees[treeID])
	}
}
