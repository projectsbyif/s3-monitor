package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/golang/glog"
	"github.com/google/trillian"
	"github.com/google/trillian/extension"
	"github.com/google/trillian/merkle/rfc6962"
	"github.com/google/trillian/server"
	"github.com/google/trillian/util/clock"
	"github.com/jamiealquiza/envy"
)

var (
	th = rfc6962.DefaultHasher
	treeId = flag.Int64("treeid", 0, "ID of the Trillian log tree events should be stored in.")
)

type LeafQueuer interface {
	QueueLeaves(context.Context, *trillian.QueueLeavesRequest) (*trillian.QueueLeavesResponse, error)
}

func CreateLeaf(hash []byte, data []byte) *trillian.LogLeaf {
	return &trillian.LogLeaf{
		MerkleLeafHash: hash,
		LeafValue:      data,
	}
}

func CreateHandler(logServer LeafQueuer, logId int64) func(ctx context.Context, s3Event events.S3Event) {
	log := logServer
	return func(ctx context.Context, s3Event events.S3Event) {
		var index int64 = 0
		var leaves []*trillian.LogLeaf

		for _, record := range s3Event.Records {
			index++
			s3 := record.S3
			fmt.Printf("[%s - %s] Bucket = %s, Key = %s \n", record.EventSource, record.EventTime, s3.Bucket.Name, s3.Object.Key)
			data, _ := json.Marshal(s3)
			hash, _ := th.HashLeaf(data)
			leaves = append(leaves, CreateLeaf(hash, data))
		}
		if len(leaves) > 0 {
			req := &trillian.QueueLeavesRequest{LogId: logId, Leaves: leaves}
			res, err := log.QueueLeaves(ctx, req)
			if err != nil {
				glog.Errorf("Unable to write leaf: %v", err)
			}
			glog.Infof("Queueing response: %v", res)
		}
	}
}

func StartTrillian(ctx context.Context, sp server.StorageProvider, treeId int64) (*server.TrillianLogRPCServer, error) {
	registry := extension.Registry{
		AdminStorage: sp.AdminStorage(),
		LogStorage:   sp.LogStorage(),
	}
	timeSource := clock.System
	logServer := *server.NewTrillianLogRPCServer(registry, timeSource)
	glog.Infof("Trillian has started, health: %v", logServer.IsHealthy())
	res, err := logServer.InitLog(ctx, &trillian.InitLogRequest{LogId: treeId})
	if err != nil {
		glog.Warningf("Unable to initlog: %v", err)
		return nil, err
	}
	glog.Infof("Log initialised: %v", res)
	return &logServer, nil
}

func main() {
	envy.Parse("LAMBDA")
	flag.Parse()
	ctx := context.Background()
	sp, err := server.NewStorageProvider("mysql", nil)
	if err != nil {
		glog.Warningf("Unable to create storage provider: %v", err)
		return
	}
	logServer, _ := StartTrillian(ctx, sp, *treeId)
	lambda.Start(CreateHandler(logServer, *treeId))
}
