package util

import (
	"crypto/sha256"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MyData struct {
	Value string
}

func (d MyData) Hash() []byte {
	hasher := sha256.New()
	hasher.Write([]byte(d.Value))
	return hasher.Sum(nil)
}

func TestNewMerkleNode(t *testing.T) {
	data := []MyData{
		{Value: "apple"},
		{Value: "banana"},
		{Value: "cherry"},
		{Value: "date"},
	}

	NewMerkleTree(data)
	//expectedMerkleRoot := "553c18f24d704b662bf49cede71758a1e7cb1dbee8948185bdc895843db0f5ca"
	//merkleRoot := mt.MerkleRoot()

	//assert.Equal(t, expectedMerkleRoot, hex.EncodeToString(merkleRoot))

}

func TestVerifyTree(t *testing.T) {
	data := []MyData{
		{Value: "apple"},
		{Value: "banana"},
		{Value: "cherry"},
		{Value: "date"},
	}

	mt := NewMerkleTree(data)

	assert.True(t, mt.VerifyTree())
}

func TestEmptyMerkleTree(t *testing.T) {
	data := []MyData{
		{Value: "apple"},
	}
	mt := NewMerkleTree(data)
	assert.True(t, mt.VerifyTree())
}
