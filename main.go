package main

import (
	"context"
	"github.com/fzft/crypto-prd-blockchain/crypto"
	"github.com/fzft/crypto-prd-blockchain/node"
	"github.com/fzft/crypto-prd-blockchain/proto"
	"github.com/fzft/crypto-prd-blockchain/util"
	"google.golang.org/grpc"
	"time"
)

func main() {
	makeNode(":3000", true)
	makeNode(":3001", false, ":3000")

	time.Sleep(2 * time.Second)
	makeNode(":3002", false, ":3001")

	for {
		time.Sleep(800 * time.Millisecond)
		makeTransaction()
	}

	select {}

}

func makeNode(listenAddr string, isValidator bool, bootstrapNodes ...string) *node.Node {
	cfg := node.ServerConfig{
		Version:    "Blocker-1.0",
		ListenAddr: listenAddr,
	}

	if isValidator {
		cfg.PrivateKey = crypto.GeneratePrivateKey()
	}
	n := node.New(cfg)
	go n.Start(listenAddr, bootstrapNodes...)
	return n
}

func makeTransaction() {
	client, err := grpc.Dial(":3000", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer client.Close()
	c := proto.NewNodeClient(client)
	prvKey := crypto.GeneratePrivateKey()
	tx := &proto.Transaction{
		Version: 1,
		Inputs: []*proto.TxInput{
			{
				PrevTxHash:   util.RandomHash(),
				PrevOutIndex: 0,
				PublicKey:    prvKey.PublicKey().Bytes(),
			},
		},

		Outputs: []*proto.TxOutput{
			{
				Amount:  999,
				Address: prvKey.PublicKey().Address().Bytes(),
			},
		},
	}

	_, err = c.HandleTransaction(context.TODO(), tx)
	if err != nil {
		panic(err)
	}
}
