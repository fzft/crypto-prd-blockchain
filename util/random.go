package util

import (
	randc "crypto/rand"
	"github.com/fzft/crypto-prd-blockchain/proto"
	"io"
	"math/rand"
	"time"
)

// RandomHash returns a random hash.
func RandomHash() []byte {
	hash := make([]byte, 32)
	if _, err := io.ReadFull(randc.Reader, hash); err != nil {
		panic(err)
	}
	return hash
}

// RandomBlock returns a random block.
func RandomBlock() *proto.Block {
	header := &proto.Header{
		Version:   1,
		Height:    int32(rand.Intn(1000)),
		PrevHash:  RandomHash(),
		RootHash:  RandomHash(),
		Timestamp: time.Now().UnixNano(),
	}

	return &proto.Block{
		Header: header,
	}
}
