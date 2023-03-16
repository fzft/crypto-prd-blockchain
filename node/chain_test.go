package node

import (
	"encoding/hex"
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
		recipient = crypto.GeneratePrivateKey().PublicKey().Bytes()
	)

	inputs := []*proto.TxInput{
		{
			PublicKey: prvKey.PublicKey().Bytes(),
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
	b.Transactions = append(b.Transactions, tx)
	require.Nil(t, chain.AddBlock(b))

	txHash := hex.EncodeToString(types.HashTransaction(tx))
	fetchedTx, err := chain.txStore.Get(txHash)
	require.Nil(t, err)
	assert.Equal(t, tx, fetchedTx)
}
