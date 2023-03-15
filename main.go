package main

import (
	"context"
	"github.com/fzft/crypto-prd-blockchain/node"
	"github.com/fzft/crypto-prd-blockchain/proto"
	"google.golang.org/grpc"
	"time"
)

func main() {
	makeNode(":3000")
	makeNode(":3001", ":3000")

	time.Sleep(2 * time.Second)
	makeNode(":3002", ":3001")

	//go func() {
	//	for {
	//		time.Sleep(2 * time.Second)
	//		makeTransaction()
	//	}
	//}()
	//
	//log.Fatal(node.Start(":3000"))

	select {}

}

func makeNode(listenAddr string, bootstrapNodes ...string) *node.Node {
	n := node.New()
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
	//tx := &proto.Transaction{
	//	Version: 1,
	//}

	version := &proto.Version{
		Version:    "0.0.1",
		Height:     200,
		ListenAddr: ":4000",
	}

	_, err = c.Handshake(context.Background(), version)
	if err != nil {
		panic(err)
	}
}
