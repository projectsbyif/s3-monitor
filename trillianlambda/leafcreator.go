package trillianlambda

import "github.com/google/trillian"

func CreateLeaf(hash []byte, data []byte) *trillian.LogLeaf {
	return &trillian.LogLeaf{
		MerkleLeafHash:   hash,
		LeafValue:        data,
		LeafIdentityHash: hash,
	}
}
