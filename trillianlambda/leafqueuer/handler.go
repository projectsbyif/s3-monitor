package main

import (
	"context"
	"flag"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/golang/glog"
	"github.com/google/trillian"
	"github.com/google/trillian/extension"
	"github.com/google/trillian/merkle/rfc6962"
	"github.com/google/trillian/server"
	"github.com/google/trillian/util/clock"
	"github.com/jamiealquiza/envy"

	"github.com/projectsbyif/verifiable-cloudtrail/trillianlambda"
)

var (
	th     = rfc6962.DefaultHasher
	treeId = flag.Int64("treeid", 0, "ID of the Trillian log tree events should be stored in.")
)

type LeafQueuer interface {
	QueueLeaves(context.Context, *trillian.QueueLeavesRequest) (*trillian.QueueLeavesResponse, error)
}

func ProcessEvents(ctx context.Context, s3Event events.S3Event, logServer LeafQueuer, treeId int64) {
	var index int64 = 0
	var leaves []*trillian.LogLeaf

	for _, record := range s3Event.Records {
		index++
		glog.Infof("[%s - %s] Bucket = %s, Key = %s \n", record.EventSource, record.EventTime, record.S3.Bucket.Name, record.S3.Object.Key)
		leaves = append(leaves, trillianlambda.CreateLeafFromS3Event(record))
	}
	if len(leaves) > 0 {
		req := &trillian.QueueLeavesRequest{LogId: treeId, Leaves: leaves}
		res, err := logServer.QueueLeaves(ctx, req)
		if err != nil {
			glog.Errorf("Unable to write leaf: %v", err)
		}
		glog.Infof("Queueing response: %v", res)
	}
}

func lambdaHandler(ctx context.Context, s3Event events.S3Event) {
	sp, err := server.NewStorageProvider("mysql", nil)
	if err != nil {
		glog.Warningf("Unable to create storage provider: %v", err)
		return
	}
	registry := extension.Registry{
		AdminStorage: sp.AdminStorage(),
		LogStorage:   sp.LogStorage(),
	}
	timeSource := clock.System
	logServer := *server.NewTrillianLogRPCServer(registry, timeSource)
	ProcessEvents(ctx, s3Event, &logServer, *treeId)
}

func main() {
	envy.Parse("LAMBDA")
	flag.Parse()
	lambda.Start(lambdaHandler)
}
