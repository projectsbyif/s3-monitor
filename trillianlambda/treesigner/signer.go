package main

import (
	"context"
	"flag"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/golang/glog"
	"github.com/google/trillian/extension"
	"github.com/google/trillian/quota"
	"github.com/google/trillian/server"
	"github.com/google/trillian/util/clock"
	"github.com/jamiealquiza/envy"

	_ "github.com/google/trillian/crypto/keyspb"
	_ "github.com/google/trillian/crypto/keys/der/proto"
	_ "github.com/google/trillian/crypto/keys/pem/proto"
	_ "github.com/google/trillian/crypto/keys/pkcs11/proto"
)

var (
	treeId = flag.Int64("treeid", 0, "ID of the Trillian log tree events should be stored in.")
)

func TreeSigner(ctx context.Context, sequencerManager *server.SequencerManager, treeId int64, info *server.LogOperationInfo) {
	glog.Infof("Running a pass on tree: %v", treeId)
	res, err := sequencerManager.ExecutePass(ctx, treeId, info)
	if err != nil {
		glog.Warningf("Unable to execute pass: %v", err)
		return
	}
	glog.Infof("Pass complete: %v", res)
}

func lambdaHandler(ctx context.Context) {
	sp, err := server.NewStorageProvider("mysql", nil)
	if err != nil {
		glog.Warningf("Unable to create storage provider: %v", err)
		return
	}
	registry := extension.Registry{
		AdminStorage: sp.AdminStorage(),
		LogStorage:   sp.LogStorage(),
		QuotaManager: quota.Noop(),
	}
	gw := time.Second
	sequencerManager := server.NewSequencerManager(registry, gw)

	info := server.LogOperationInfo{
		Registry:   registry,
		TimeSource: clock.System,
		BatchSize:  100,
	}

	TreeSigner(ctx, sequencerManager, *treeId, &info)
}

func main() {
	envy.Parse("LAMBDA")
	flag.Parse()
	lambda.Start(lambdaHandler)
}
