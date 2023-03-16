package types

import (
	"crypto/sha256"
	"github.com/fzft/crypto-prd-blockchain/crypto"
	"github.com/fzft/crypto-prd-blockchain/proto"
	pb "github.com/golang/protobuf/proto"
)

// VerifyBlock verifies the block.
func VerifyBlock(block *proto.Block) bool {
	if len(block.Signature) != crypto.SignatureLen {
		return false
	}

	if len(block.PublicKey) != crypto.PubKeyLen {
		return false
	}

	sig := crypto.SignatureFromBytes(block.Signature)
	pubKey := crypto.PublicKeyFromBytes(block.PublicKey)
	return sig.Verify(HashBlock(block), pubKey)
}

// SignBlock signs the block.
func SignBlock(pk *crypto.PrivateKey, block *proto.Block) *crypto.Signature {
	hash := HashBlock(block)
	sig := pk.Sign(hash)
	block.PublicKey = pk.PublicKey().Bytes()
	block.Signature = sig.Bytes()
	return sig
}

// HashBlock returns the hash of the block.
func HashBlock(block *proto.Block) []byte {
	return HashHeader(block.Header)
}

func HashHeader(header *proto.Header) []byte {
	b, err := pb.Marshal(header)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]
}
