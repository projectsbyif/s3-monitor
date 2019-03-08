package main

import (
	"context"
	"crypto"
	"crypto/rand"
	"testing"
	"time"

	"github.com/google/trillian"
	"github.com/google/trillian/extension"
	"github.com/google/trillian/merkle/rfc6962"
	"github.com/google/trillian/quota"
	"github.com/google/trillian/server"
	"github.com/google/trillian/storage"
	"github.com/google/trillian/util/clock"
	"github.com/projectsbyif/s3-monitor/trillianlambda"

	tcrypto "github.com/google/trillian/crypto"
	"github.com/google/trillian/crypto/keys/der"
	stestonly "github.com/google/trillian/storage/testonly"
)

var (
	th   = rfc6962.DefaultHasher
	tree *trillian.Tree
)

func numberOfLeaves(ls storage.LogStorage) int {
	tx, _ := ls.SnapshotForTree(context.TODO(), tree)
	defer tx.Commit()
	leaves, _ := tx.GetLeavesByRange(context.TODO(), 0, 10)
	return len(leaves)
}

func assertLeavesAdded(t *testing.T, ls storage.LogStorage, expectedNumberOfNewLeaves int, f func()) {
	countBefore := numberOfLeaves(ls)
	f()
	countAfter := numberOfLeaves(ls)
	if countAfter-countBefore != expectedNumberOfNewLeaves {
		t.Errorf("Expected %v leaf in tree, found %v", expectedNumberOfNewLeaves, countAfter-countBefore)
	}
}

func leafGenerator() *trillian.LogLeaf {
	data := make([]byte, 10)
	rand.Read(data)
	hash, _ := th.HashLeaf(data)

	return trillianlambda.CreateLeaf(hash, data)
}

func isValidSignedLogRoot(t *testing.T, signedLogRoot trillian.SignedLogRoot) {
	publicKey, _ := der.FromPublicProto(tree.GetPublicKey())

	_, err := tcrypto.VerifySignedLogRoot(publicKey, crypto.SHA256, &signedLogRoot)
	if err != nil {
		t.Errorf("Not a valid signed log root: %v", err)
	}
}

func TestTreeSignerTrillianIntegration(t *testing.T) {
	// Setup code
	ctx := context.Background()
	sp, _ := server.NewStorageProvider("memory", nil)
	sp.AdminStorage().ReadWriteTransaction(ctx, func(ctx context.Context, tx storage.AdminTX) error {
		tree, _ = tx.CreateTree(ctx, stestonly.LogTree)
		tx.Commit()
		return nil
	})

	registry := extension.Registry{
		AdminStorage: sp.AdminStorage(),
		LogStorage:   sp.LogStorage(),
		QuotaManager: quota.Noop(),
	}
	gw := time.Second
	sequencerManager := server.NewSequencerManager(registry, gw)

	logServer := *server.NewTrillianLogRPCServer(registry, clock.System)
	logServer.InitLog(ctx, &trillian.InitLogRequest{LogId: tree.TreeId})

	info := server.LogOperationInfo{
		Registry:   registry,
		TimeSource: clock.System,
		BatchSize:  100,
	}

	queueTimestamp := time.Now()

	t.Run("no events", func(t *testing.T) {
		assertLeavesAdded(t, sp.LogStorage(), 0, func() {
			sp.LogStorage().QueueLeaves(ctx, tree, []*trillian.LogLeaf{}, queueTimestamp)
			signedLogRoot, _ := TreeSigner(ctx, sequencerManager, tree.TreeId, &info)
			isValidSignedLogRoot(t, signedLogRoot)
		})
	})

	t.Run("one event", func(t *testing.T) {
		assertLeavesAdded(t, sp.LogStorage(), 1, func() {
			sp.LogStorage().QueueLeaves(ctx, tree, []*trillian.LogLeaf{leafGenerator()}, queueTimestamp)
			signedLogRoot, _ := TreeSigner(ctx, sequencerManager, tree.TreeId, &info)
			isValidSignedLogRoot(t, signedLogRoot)
		})
	})

	t.Run("five events", func(t *testing.T) {
		assertLeavesAdded(t, sp.LogStorage(), 5, func() {
			sp.LogStorage().QueueLeaves(ctx, tree, []*trillian.LogLeaf{leafGenerator(), leafGenerator(), leafGenerator(), leafGenerator(), leafGenerator()}, queueTimestamp)
			signedLogRoot, _ := TreeSigner(ctx, sequencerManager, tree.TreeId, &info)
			isValidSignedLogRoot(t, signedLogRoot)
		})
	})
}
