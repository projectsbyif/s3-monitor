package main

import (
	"context"
	"flag"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/trillian/server"
	"github.com/jamiealquiza/envy"
)

func TreeSigner(ctx context.Context, sequencerManager *server.SequencerManager, treeId int64, info *server.LogOperationInfo) {
	sequencerManager.ExecutePass(ctx, treeId, info)
}

func lambdaHandler(ctx context.Context) {
	// LogStorage

	// registry

	// guardWindow

	// SequencerManager

	// info

	// TreeSigner(ctx, SequencerManager, treeId, info)
}

func main() {
	envy.Parse("LAMBDA")
	flag.Parse()
	lambda.Start(lambdaHandler)
}
