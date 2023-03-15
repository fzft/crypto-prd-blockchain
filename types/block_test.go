package types

import (
	"github.com/fzft/crypto-prd-blockchain/crypto"
	"github.com/fzft/crypto-prd-blockchain/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHashBlock(t *testing.T) {
	block := util.RandomBlock()
	hash := HashBlock(block)
	assert.Equal(t, 32, len(hash))
}

func TestSignBlock(t *testing.T) {
	prvKey := crypto.GeneratePrivateKey()
	pubKey := prvKey.PublicKey()
	block := util.RandomBlock()
	sig := SignBlock(prvKey, block)

	assert.Equal(t, 64, len(sig.Bytes()))
	assert.True(t, sig.Verify(HashBlock(block), pubKey))
}