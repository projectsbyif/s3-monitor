package main

import (
	"context"
	"github.com/google/trillian"
	"github.com/google/trillian/server"
	"github.com/google/trillian/storage"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"testing"

	stestonly "github.com/google/trillian/storage/testonly"
)

func TestHandleRequest(t *testing.T) {
	s := `
{
  "Records": [
    {
      "eventVersion": "2.0",
      "eventSource": "aws:s3",
      "awsRegion": "us-east-1",
      "eventTime": "1970-01-01T00:00:00.123Z",
      "eventName": "ObjectCreated:Put",
      "userIdentity": {
        "principalId": "EXAMPLE"
      },
      "requestParameters": {
        "sourceIPAddress": "127.0.0.1"
      },
      "responseElements": {
        "x-amz-request-id": "C3D13FE58DE4C810",
        "x-amz-id-2": "FMyUVURIY8/IgAtTv8xRjskZQpcIZ9KG4V5Wp6S7S/JRWeUWerMUE5JgHvANOjpD"
      },
      "s3": {
        "s3SchemaVersion": "1.0",
        "configurationId": "testConfigRule",
        "bucket": {
          "name": "sourcebucket",
          "ownerIdentity": {
            "principalId": "EXAMPLE"
          },
          "arn": "arn:aws:s3:::mybucket"
        },
        "object": {
          "key": "HappyFace.jpg",
          "size": 1024,
          "urlDecodedKey": "HappyFace.jpg",
          "versionId": "version",
          "eTag": "d41d8cd98f00b204e9800998ecf8427e",
          "sequencer": "Happy Sequencer"
        }
      }
    }
  ]
}
`
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
	treeId := tree.TreeId
	logServer, _ := StartTrillian(ctx, sp, treeId)
	handleRequest := CreateHandler(logServer, treeId)
	handleRequest(ctx, inputEvent)
	tx, _ := sp.LogStorage().Snapshot(ctx)
	defer tx.Close()
	trees, _ := tx.GetUnsequencedCounts(ctx)
	if trees[treeId] != 1 {
		t.Errorf("Expected 1 leaf in tree, found %v", trees[treeId])
	}
}
