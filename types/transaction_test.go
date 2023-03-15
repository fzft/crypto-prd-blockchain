package types

import (
	"github.com/fzft/crypto-prd-blockchain/crypto"
	"github.com/fzft/crypto-prd-blockchain/proto"
	"github.com/fzft/crypto-prd-blockchain/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHashTransaction(t *testing.T) {
	fromPrvKey := crypto.GeneratePrivateKey()
	fromPubKey := fromPrvKey.PublicKey()
	fromAddress := fromPubKey.Address().Bytes()
	toPrvKey := crypto.GeneratePrivateKey()
	toAddress := toPrvKey.PublicKey().Address().Bytes()

	input := &proto.TxInput{
		PrevTxHash:   util.RandomHash(),
		PrevOutIndex: 0,
		PublicKey:    fromPubKey.Bytes(),
	}

	output1 := &proto.TxOutput{
		Amount:  5,
		Address: toAddress,
	}

	output2 := &proto.TxOutput{
		Amount:  95,
		Address: fromAddress,
	}

	tx := &proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{input},
		Outputs: []*proto.TxOutput{output1, output2},
	}

	sig := SignTransaction(fromPrvKey, tx)
	input.Signature = sig.Bytes()

	assert.True(t, VerifyTransaction(tx))
}
