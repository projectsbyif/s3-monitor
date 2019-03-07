package trillianlambda

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/trillian"
	"github.com/google/trillian/merkle/rfc6962"
)

var (
	th = rfc6962.DefaultHasher
)

type S3LeafEvent struct {
	BucketArn   string `json:"bucket_arn"`
	Key         string `json:"key"`
	Size        int64  `json:"size"`
	ETag        string `json:"eTag"`
	EventSource string `json:"eventSource"`
}

func CreateLeafFromS3Event(event events.S3EventRecord) *trillian.LogLeaf {
	leafEvent := S3LeafEvent{
		BucketArn:   event.S3.Bucket.Arn,
		Key:         event.S3.Object.Key,
		Size:        event.S3.Object.Size,
		ETag:        event.S3.Object.ETag,
		EventSource: event.EventSource,
	}
	data, _ := json.Marshal(leafEvent)
	hash, _ := th.HashLeaf(data)
	return CreateLeaf(hash, data)

}

func CreateLeaf(hash []byte, data []byte) *trillian.LogLeaf {
	return &trillian.LogLeaf{
		MerkleLeafHash:   hash,
		LeafValue:        data,
		LeafIdentityHash: hash,
	}
}
