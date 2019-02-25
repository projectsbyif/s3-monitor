package main

import (
	"flag"
	"context"
	"fmt"
	"os"
	"strconv"
	"github.com/golang/glog"
	"github.com/google/trillian"
	"github.com/google/trillian/server"
	"github.com/google/trillian/extension"
	"github.com/google/trillian/util/clock"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var logServer server.TrillianLogRPCServer

func HandleRequest(ctx context.Context, s3Event events.S3Event) {
	for _, record := range s3Event.Records {
		s3 := record.S3
		fmt.Printf("[%s - %s] Bucket = %s, Key = %s \n", record.EventSource, record.EventTime, s3.Bucket.Name, s3.Object.Key)
		// server.QueueLeaf
	}
}

func StartTrillian(ctx context.Context, sp server.StorageProvider, treeId int64) {
	registry := extension.Registry{
		AdminStorage:  sp.AdminStorage(),
		LogStorage:    sp.LogStorage(),
	}
	timeSource := clock.System
	logServer = *server.NewTrillianLogRPCServer(registry, timeSource)
	glog.Infof("Trillian has started, health: %v", logServer.IsHealthy())
	res, err := logServer.InitLog(ctx, &trillian.InitLogRequest{LogId: treeId})
	if err != nil {
		glog.Warningf("Unable to initlog: %v", err)
		return
	}
	glog.Infof("Log initialised: %v", res)
}

func main() {
	flag.Parse()
	ctx := context.Background()
	sp, err := server.NewStorageProvider("mysql", nil)
	if err != nil {
		glog.Warningf("Unable to create storage provider: %v", err)
		return
	}
	treeId, err := strconv.ParseInt(os.Getenv("TREE_ID"), 0, 64)
	if err != nil {
		glog.Warningf("Invalid tree ID: %v", err)
		return
	}
	StartTrillian(ctx, sp, treeId)
	lambda.Start(HandleRequest)
}
