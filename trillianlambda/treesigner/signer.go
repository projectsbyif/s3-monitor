package main

import (
	"bytes"
	"context"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3/s3manager"
	"github.com/golang/glog"
	"github.com/google/trillian"
	"github.com/google/trillian/extension"
	"github.com/google/trillian/quota"
	"github.com/google/trillian/server"
	"github.com/google/trillian/trees"
	"github.com/google/trillian/util/clock"
	"github.com/jamiealquiza/envy"

	"github.com/google/trillian/crypto/keys/der"
	_ "github.com/google/trillian/crypto/keys/der/proto"
	_ "github.com/google/trillian/crypto/keys/pem/proto"
	_ "github.com/google/trillian/crypto/keys/pkcs11/proto"
	_ "github.com/google/trillian/crypto/keyspb"
)

var (
	treeId     = flag.Int64("treeid", 0, "ID of the Trillian log tree events should be stored in.")
	bucketName = flag.String("bucket_name", "", "S3 Bucket containing signed log roots.")
)

type LogRootVerificationData struct {
	RootHash         string
	LogRootSignature string
	PublicKey        crypto.PublicKey
}

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

	if err != nil {
		glog.Errorf("Unable to snapshot for tree: %v", err)
		return trillian.SignedLogRoot{}, err
	}

	slr, err := tx.LatestSignedLogRoot(ctx)
	tx.Commit()

	if err != nil {
		glog.Errorf("Unable to get latest signed log root: %v", err)
		return trillian.SignedLogRoot{}, err
	}

	return slr, nil
}

func publishToS3(ctx context.Context, registry extension.Registry, treeId int64, signedLogRoot trillian.SignedLogRoot) {
	rootHash := base64.StdEncoding.EncodeToString(signedLogRoot.GetRootHash())
	logRootSignature := base64.StdEncoding.EncodeToString(signedLogRoot.GetLogRootSignature())
	glog.Infof("Publishing to S3 for root hash: %v", rootHash)

	tree, err := trees.GetTree(ctx, registry.AdminStorage, treeId, trees.GetOpts{Operation: trees.Admin})
	glog.Infof("Got tree ID: %v", tree.GetTreeId())
	if err != nil {
		glog.Errorf("Failed to get Tree, %v", err)
	}
	publicKey, _ := der.FromPublicProto(tree.GetPublicKey())

	l := LogRootVerificationData{rootHash, logRootSignature, publicKey}
	body, err := json.Marshal(l)
	if err != nil {
		glog.Errorf("Failed to marshal json, %v", err)
	}

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		glog.Errorf("Failed to LoadDefaultAWSConfig, %v", err)
	}
	uploader := s3manager.NewUploader(cfg)

	t := time.Now()
	timekey := t.Format("2006/01/02")

	glog.Info("Uploading to S3 len: %v", len(body))
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(*bucketName),
		Key:    aws.String(fmt.Sprintf("%v/%v.json", timekey, treeId)),
		Body:   bytes.NewReader(body),
	})

	if err != nil {
		glog.Errorf("failed to upload file, %v", err)
	}

	glog.Infof("file uploaded to, %s", aws.StringValue(&result.Location))
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

	signedLogRoot, _ := TreeSigner(ctx, sequencerManager, *treeId, &info)
	publishToS3(ctx, registry, *treeId, signedLogRoot)
}

func main() {
	envy.Parse("LAMBDA")
	flag.Parse()
	lambda.Start(lambdaHandler)
}
