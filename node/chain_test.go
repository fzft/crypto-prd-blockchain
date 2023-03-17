package node

import (
	"encoding/hex"
	"fmt"
	"github.com/fzft/crypto-prd-blockchain/crypto"
	"github.com/fzft/crypto-prd-blockchain/proto"
	"github.com/fzft/crypto-prd-blockchain/types"
	"github.com/fzft/crypto-prd-blockchain/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func randomBlock(t *testing.T, chain *Chain) *proto.Block {
	prvKey := crypto.GeneratePrivateKey()
	b := util.RandomBlock()
	preBlock, err := chain.GetBlockByHeight(chain.Height())
	require.Nil(t, err)
	b.Header.PrevHash = types.HashBlock(preBlock)
	types.SignBlock(prvKey, b)
	return b
}

func TestChainHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTxStore())
	for i := 0; i < 10; i++ {
		b := randomBlock(t, chain)
		require.Nil(t, chain.AddBlock(b))
		assert.Equal(t, i+1, chain.Height())
	}
}

func TestChainAddBlock(t *testing.T) {
	bs := NewMemoryBlockStore()
	chain := NewChain(bs, NewMemoryTxStore())

	for i := 0; i < 10; i++ {
		block := randomBlock(t, chain)
		blockHash := types.HashBlock(block)
		require.Nil(t, chain.AddBlock(block))

		fetchedBlock, err := chain.GetBlockByHash(blockHash)
		require.Nil(t, err)
		assert.Equal(t, block, fetchedBlock)

		fetchedBlockByHeight, err := chain.GetBlockByHeight(i + 1)
		require.Nil(t, err)
		assert.Equal(t, block, fetchedBlockByHeight)
	}
}

func TestNewChain(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTxStore())
	assert.Equal(t, 0, chain.Height())

	_, err := chain.GetBlockByHeight(0)
	require.Nil(t, err)
}

func TestAddBlockWithTx(t *testing.T) {
	var (
		chain     = NewChain(NewMemoryBlockStore(), NewMemoryTxStore())
		b         = randomBlock(t, chain)
		prvKey    = crypto.GeneratePrivateKeyFromSeedStr(godSeed)
		recipient = crypto.GeneratePrivateKey().PublicKey().Address().Bytes()
	)

	// hard code for genesis block
	prevTx, err := chain.txStore.Get("28706e21cb96f758ee855577b6111cb8788d9b4c9eead5e19c09baa05ee16d90")

	inputs := []*proto.TxInput{
		{
			PublicKey:    prvKey.PublicKey().Bytes(),
			PrevOutIndex: 0,
			PrevTxHash:   types.HashTransaction(prevTx),
		},
	}
	outputs := []*proto.TxOutput{
		{
			Amount:  100,
			Address: recipient,
		},
		{
			Amount:  900,
			Address: prvKey.PublicKey().Address().Bytes(),
		},
	}

	tx := &proto.Transaction{
		Version: 1,
		Inputs:  inputs,
		Outputs: outputs,
	}

	sig := types.SignTransaction(prvKey, tx)
	tx.Inputs[0].Signature = sig.Bytes()

	b.Transactions = append(b.Transactions, tx)
	require.Nil(t, chain.AddBlock(b))

	// check if all the outputs are unspent y querying the utxo store
	txHash := hex.EncodeToString(types.HashTransaction(tx))
	fetchedTx, err := chain.txStore.Get(txHash)
	require.Nil(t, err)
	assert.Equal(t, tx, fetchedTx)

	for i := 0; i < len(tx.Outputs); i++ {
		key := fmt.Sprintf("%s_%d", txHash, i)
		utxo, err := chain.utxoStore.Get(key)
		require.Nil(t, err)
		assert.Equal(t, txHash, utxo.Hash)
		assert.True(t, !utxo.Spent)
	}
}

func TestAddBlockWithTxInsufficientFunds(t *testing.T) {
	var (
		chain     = NewChain(NewMemoryBlockStore(), NewMemoryTxStore())
		b         = randomBlock(t, chain)
		prvKey    = crypto.GeneratePrivateKeyFromSeedStr(godSeed)
		recipient = crypto.GeneratePrivateKey().PublicKey().Address().Bytes()
	)

	// hard code for genesis block
	prevTx, _ := chain.txStore.Get("28706e21cb96f758ee855577b6111cb8788d9b4c9eead5e19c09baa05ee16d90")

	inputs := []*proto.TxInput{
		{
			PublicKey:    prvKey.PublicKey().Bytes(),
			PrevOutIndex: 0,
			PrevTxHash:   types.HashTransaction(prevTx),
		},
	}
	outputs := []*proto.TxOutput{
		{
			Amount:  10001,
			Address: recipient,
		},
		{
			Amount:  1,
			Address: prvKey.PublicKey().Address().Bytes(),
		},
	}

	tx := &proto.Transaction{
		Version: 1,
		Inputs:  inputs,
		Outputs: outputs,
	}

	sig := types.SignTransaction(prvKey, tx)
	tx.Inputs[0].Signature = sig.Bytes()

	b.Transactions = append(b.Transactions, tx)
	require.Nil(t, chain.AddBlock(b))
}
