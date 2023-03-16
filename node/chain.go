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

type Chain struct {
	blockStore BlockStore
	txStore    TxStore
	headers    *HeaderList
}

func NewChain(bs BlockStore, ts TxStore) *Chain {
	chain := &Chain{blockStore: bs, headers: NewHeaderList(), txStore: ts}
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
	return nil
}

// addBlock adds a block to the chain
func (c *Chain) addBlock(block *proto.Block) error {
	c.headers.AddHeader(block.Header)

	for _, tx := range block.Transactions {
		if err := c.txStore.Put(tx); err != nil {
			return err
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
				Address: prvKey.PublicKey().Bytes(),
			},
		},
	}

	block.Transactions = append(block.Transactions, tx)
	types.SignBlock(prvKey, block)
	return block
}
