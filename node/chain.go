package node

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/fzft/crypto-prd-blockchain/crypto"
	"github.com/fzft/crypto-prd-blockchain/proto"
	"github.com/fzft/crypto-prd-blockchain/types"
)

const godSeed = "b3853c01222f908d08a87d0dd8ce7b0d1324d9967b5da9342ee185d5c1ee295e"

type HeaderList struct {
	headers []*proto.Header
}

func NewHeaderList() *HeaderList {
	return &HeaderList{}
}

// Get returns the header at the given index
func (l *HeaderList) Get(index int) *proto.Header {
	if index > l.Height() {
		panic("index too high")
	}

	return l.headers[index]
}

func (l *HeaderList) AddHeader(header *proto.Header) {
	l.headers = append(l.headers, header)
}

// Len returns the length of the list
func (l *HeaderList) Len() int {
	return len(l.headers)
}

// Height returns the height of the list
func (l *HeaderList) Height() int {
	return l.Len() - 1
}

type UTXO struct {
	Hash     string
	OutIndex int
	Amount   int64
	Spent    bool
}

type Chain struct {
	blockStore BlockStore
	txStore    TxStore
	headers    *HeaderList
	utxoStore  UTXOSore
}

func NewChain(bs BlockStore, ts TxStore) *Chain {
	chain := &Chain{blockStore: bs, headers: NewHeaderList(), txStore: ts, utxoStore: NewMemoryUTXOStore()}
	chain.addBlock(createGenesisBlock())
	return chain
}

// Height returns the height of the chain
func (c *Chain) Height() int {
	return c.headers.Height()
}

// AddBlock adds a block to the chain
func (c *Chain) AddBlock(block *proto.Block) error {
	if err := c.ValidateBlock(block); err != nil {
		return err
	}
	return c.addBlock(block)
}

// GetBlockByHeight returns a block by its height
func (c *Chain) GetBlockByHeight(height int) (*proto.Block, error) {
	if c.Height() < height {
		return nil, fmt.Errorf("given height (%d) too high - height (%d)", height, c.Height())
	}

	header := c.headers.Get(height)
	hash := hex.EncodeToString(types.HashHeader(header))
	return c.blockStore.Get(hash)
}

// GetBlockByHash returns a block by its hash
func (c *Chain) GetBlockByHash(hash []byte) (*proto.Block, error) {
	hashHex := hex.EncodeToString(hash)
	return c.blockStore.Get(hashHex)
}

// ValidateBlock validates a block
func (c *Chain) ValidateBlock(block *proto.Block) error {
	// validate the signature of the block
	if !types.VerifyBlock(block) {
		return fmt.Errorf("block signature is invalid")
	}

	// validate if the preHash is the actual hash of the current block
	currentBlock, err := c.GetBlockByHeight(c.Height())
	if err != nil {
		return err
	}
	hash := types.HashBlock(currentBlock)
	if !bytes.Equal(hash, block.Header.PrevHash) {
		return fmt.Errorf("block hash does not match previous block hash")
	}

	for _, tx := range block.Transactions {
		if err = c.validateTransaction(tx); err != nil {
			return err
		}
	}

	return nil
}

// validateTransaction validates a transaction
func (c *Chain) validateTransaction(tx *proto.Transaction) error {
	if !types.VerifyTransaction(tx) {
		return fmt.Errorf("invalid tx signature")
	}

	//validate if the inputs are valid
	nInputs := len(tx.Inputs)
	for i := 0; i < nInputs; i++ {
		preTxHash := hex.EncodeToString(tx.Inputs[i].PrevTxHash)
		key := fmt.Sprintf("%s_%d", preTxHash, i)
		utxo, err := c.utxoStore.Get(key)
		if err != nil {
			return err
		}
		if utxo.Spent {
			return fmt.Errorf("output %d of tx %s is already spent", i, preTxHash)
		}

		//TODO(fzft): validate if the amount is correct

	}

	return nil
}

// addBlock adds a block to the chain
func (c *Chain) addBlock(block *proto.Block) error {
	c.headers.AddHeader(block.Header)

	for _, tx := range block.Transactions {
		if err := c.txStore.Put(tx); err != nil {
			return err
		}

		hash := hex.EncodeToString(types.HashTransaction(tx))

		// address_txHash
		for it, output := range tx.Outputs {
			utxo := &UTXO{
				Hash:     hash,
				OutIndex: it,
				Amount:   output.Amount,
				Spent:    false,
			}

			if err := c.utxoStore.Put(utxo); err != nil {
				return err
			}
		}

	}

	return c.blockStore.Put(block)
}

// createGenesisBlock creates a genesis block
func createGenesisBlock() *proto.Block {
	// hard coded genesis block
	prvKey := crypto.GeneratePrivateKeyFromSeedStr(godSeed)

	block := &proto.Block{
		Header: &proto.Header{
			Version: 1,
		},
	}

	tx := &proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{},
		Outputs: []*proto.TxOutput{
			{
				Amount:  1000,
				Address: prvKey.PublicKey().Address().Bytes(),
			},
		},
	}

	block.Transactions = append(block.Transactions, tx)
	types.SignBlock(prvKey, block)
	return block
}
