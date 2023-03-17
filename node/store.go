package node

import (
	"encoding/hex"
	"fmt"
	"github.com/fzft/crypto-prd-blockchain/proto"
	"github.com/fzft/crypto-prd-blockchain/types"
	"github.com/fzft/crypto-prd-blockchain/util"
)

type TxStore interface {
	Put(*proto.Transaction) error
	Get(string) (*proto.Transaction, error)
}

type BlockStore interface {
	Put(buffer *proto.Block) error
	Get(string) (*proto.Block, error)
}

type UTXOSore interface {
	Put(utxo *UTXO) error
	Get(string) (*UTXO, error)
}

type MemoryBlockStore struct {
	blocks *util.KeyValueStore[string, *proto.Block]
}

func NewMemoryBlockStore() *MemoryBlockStore {
	return &MemoryBlockStore{
		blocks: util.NewKeyValueStore[string, *proto.Block](),
	}
}

func (m *MemoryBlockStore) Put(block *proto.Block) error {
	hash := hex.EncodeToString(types.HashBlock(block))
	m.blocks.Put(hash, block)
	return nil
}

func (m *MemoryBlockStore) Get(hash string) (*proto.Block, error) {
	block, ok := m.blocks.Get(hash)
	if !ok {
		return nil, fmt.Errorf("block with hash %s not found", hash)
	}
	return block, nil
}

type MemoryTxStore struct {
	txx *util.KeyValueStore[string, *proto.Transaction]
}

func NewMemoryTxStore() *MemoryTxStore {
	return &MemoryTxStore{
		txx: util.NewKeyValueStore[string, *proto.Transaction](),
	}
}

func (m *MemoryTxStore) Put(tx *proto.Transaction) error {
	hash := hex.EncodeToString(types.HashTransaction(tx))
	m.txx.Put(hash, tx)
	return nil
}

func (m *MemoryTxStore) Get(hash string) (*proto.Transaction, error) {
	tx, ok := m.txx.Get(hash)
	if !ok {
		return nil, fmt.Errorf("transaction with hash %s not found", hash)
	}
	return tx, nil
}

type MemoryUTXOStore struct {
	utxos *util.KeyValueStore[string, *UTXO]
}

func NewMemoryUTXOStore() *MemoryUTXOStore {
	return &MemoryUTXOStore{
		utxos: util.NewKeyValueStore[string, *UTXO](),
	}
}

func (m *MemoryUTXOStore) Put(utxo *UTXO) error {
	key := fmt.Sprintf("%s_%d", utxo.Hash, utxo.OutIndex)

	m.utxos.Put(key, utxo)
	return nil
}

func (m *MemoryUTXOStore) Get(hash string) (*UTXO, error) {
	utxo, ok := m.utxos.Get(hash)
	if !ok {
		return nil, fmt.Errorf("utxo with hash %s not found", hash)
	}
	return utxo, nil
}
