package types

import (
	"crypto/sha256"
	"github.com/fzft/crypto-prd-blockchain/crypto"
	"github.com/fzft/crypto-prd-blockchain/proto"
	pb "github.com/golang/protobuf/proto"
)

// SignBlock signs the block.
func SignBlock(pk *crypto.PrivateKey, block *proto.Block) *crypto.Signature {
	return pk.Sign(HashBlock(block))
}

// HashBlock returns the hash of the block.
func HashBlock(block *proto.Block) []byte {
	b, err := pb.Marshal(block)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]
}
