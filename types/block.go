package types

import (
	"bytes"
	"crypto/sha256"
	"github.com/fzft/crypto-prd-blockchain/crypto"
	"github.com/fzft/crypto-prd-blockchain/proto"
	"github.com/fzft/crypto-prd-blockchain/util"
	pb "github.com/golang/protobuf/proto"
)

const godSeed = "b3853c01222f908d08a87d0dd8ce7b0d1324d9967b5da9342ee185d5c1ee295e"

type TxHash struct {
	hash []byte
}

func NewTxHash(hash []byte) TxHash {
	return TxHash{
		hash: hash,
	}
}

func (t TxHash) Hash() []byte {
	return t.hash
}

// VerifyBlock verifies the block.
func VerifyBlock(block *proto.Block) bool {
	if len(block.Transactions) > 0 {
		if !VerifyRootHash(block) {
			return false
		}
	}

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

// GenerateCoinbaseTx generates a coinbase transaction.
func GenerateCoinbaseTx() *proto.Transaction {
	// hard code the god seed
	prvKey := crypto.GeneratePrivateKeyFromSeedStr(godSeed)
	return &proto.Transaction{
		Version: 1,
		Inputs: []*proto.TxInput{
			{
				PrevTxHash:   []byte{},
				PrevOutIndex: 0xffffffff,
				Signature:    []byte{},
				PublicKey:    []byte{},
			},
		},
		Outputs: []*proto.TxOutput{
			{
				Amount:  100,
				Address: prvKey.PublicKey().Address().Bytes(),
			},
		},
	}
}

// SignBlock signs the block.
func SignBlock(pk *crypto.PrivateKey, block *proto.Block) *crypto.Signature {
	hash := HashBlock(block)
	sig := pk.Sign(hash)
	block.PublicKey = pk.PublicKey().Bytes()
	block.Signature = sig.Bytes()

	if len(block.Transactions) > 0 {
		tree := GetMerkleTree(block)
		block.Header.RootHash = tree.MerkleRoot()
	}

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

// GetMerkleTree calculates the root hash of the merkle tree.
func GetMerkleTree(b *proto.Block) *util.MerkleTree[TxHash] {
	if len(b.Transactions) == 0 {
		return nil
	}

	list := make([]TxHash, len(b.Transactions))
	for i, tx := range b.Transactions {
		list[i] = NewTxHash(HashTransaction(tx))
	}

	t := util.NewMerkleTree(list)
	return t
}

// VerifyRootHash verifies the root hash of the merkle tree.
func VerifyRootHash(b *proto.Block) bool {
	tree := GetMerkleTree(b)
	valid := tree.VerifyTree()

	if len(b.Header.RootHash) == 0 {
		b.Header.RootHash = tree.MerkleRoot()
	}

	if !valid {
		return false
	}

	return bytes.Equal(tree.MerkleRoot(), b.Header.RootHash)
}
