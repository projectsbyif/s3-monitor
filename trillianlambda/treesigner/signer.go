package main

import (
	"context"
	"flag"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/golang/glog"
	"github.com/google/trillian"
	"github.com/google/trillian/extension"
	"github.com/google/trillian/quota"
	"github.com/google/trillian/server"
	"github.com/google/trillian/trees"
	"github.com/google/trillian/util/clock"
	"github.com/jamiealquiza/envy"

	_ "github.com/google/trillian/crypto/keys/der/proto"
	_ "github.com/google/trillian/crypto/keys/pem/proto"
	_ "github.com/google/trillian/crypto/keys/pkcs11/proto"
	_ "github.com/google/trillian/crypto/keyspb"
)

var (
	treeId = flag.Int64("treeid", 0, "ID of the Trillian log tree events should be stored in.")
)

func TreeSigner(ctx context.Context, sequencerManager server.LogOperation, treeId int64, info *server.LogOperationInfo) (trillian.SignedLogRoot, error) {
	glog.Infof("Running a pass on tree: %v", treeId)
	res, err := sequencerManager.ExecutePass(ctx, treeId, info)
	if err != nil {
		glog.Warningf("Unable to execute pass: %v", err)
		return trillian.SignedLogRoot{}, err
	}
	glog.Infof("Pass complete: %v", res)

	signedLogRoot, err := getSignedLogRoot(ctx, treeId, info.Registry)
	if err != nil {
		glog.Errorf("Unable to get signed log root: %v", err)
		return trillian.SignedLogRoot{}, err
	}

	return signedLogRoot, nil
}

func getSignedLogRoot(ctx context.Context, treeId int64, registry extension.Registry) (trillian.SignedLogRoot, error) {
	tree, err := trees.GetTree(ctx, registry.AdminStorage, treeId, trees.GetOpts{Operation: trees.Admin})

	if err != nil {
		glog.Errorf("Unable to get tree: %v", err)
		return trillian.SignedLogRoot{}, err
	}

	tx, err := registry.LogStorage.SnapshotForTree(ctx, tree)
	defer tx.Close()

	if err != nil {
		glog.Errorf("Unable to snapshot for tree: %v", err)
		return trillian.SignedLogRoot{}, err
	}

	slr, err := tx.LatestSignedLogRoot(ctx)

	if err != nil {
		glog.Errorf("Unable to get latest signed log root: %v", err)
		return trillian.SignedLogRoot{}, err
	}

	return slr, nil
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
