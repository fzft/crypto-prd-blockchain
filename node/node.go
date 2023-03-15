package node

import (
	"context"
	"fmt"
	"github.com/fzft/crypto-prd-blockchain/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"net"
	"sync"
)

type Node struct {
	version    string
	listenAddr string
	logger     *zap.SugaredLogger

	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version
	proto.UnimplementedNodeServer
}

func New() *Node {
	logger, _ := zap.NewProduction()
	return &Node{
		version: "blocker-0.0.1",
		peers:   make(map[proto.NodeClient]*proto.Version),
		logger:  logger.Sugar(),
	}
}

// Start ...
func (n *Node) Start(listenAddr string, bootstrapNodes ...string) error {
	n.listenAddr = listenAddr
	var (
		opts       []grpc.ServerOption
		grpcServer = grpc.NewServer(opts...)
	)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		panic(err)
	}

	proto.RegisterNodeServer(grpcServer, n)
	n.logger.Infof("Listening on %s", listenAddr)

	// bootstrap the network with a list of already known nodes

	if len(bootstrapNodes) > 0 {
		go n.bootstrapNetwork(bootstrapNodes...)
	}

	return grpcServer.Serve(ln)
}

// Handshake is called when a new peer connects to the node.
func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	peer, _ := peer.FromContext(ctx)
	c, err := makeNodeClient(peer.Addr.String())
	if err != nil {
		return nil, err
	}

	n.addPeer(c, v)
	return n.getVersion(), nil
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)
	fmt.Printf("Received transaction from %s\n", peer.Addr.String())
	return &proto.Ack{}, nil
}

// BootstrapNetwork connects to the bootstrap nodes and syncs the blockchain.
func (n *Node) bootstrapNetwork(bootstrapNodes ...string) error {
	for _, addr := range bootstrapNodes {
		if !n.canConnectWith(addr) {
			continue
		}
		c, v, err := n.dialRemoteNode(addr)
		if err != nil {
			n.logger.Errorf("Error connecting to bootstrap node (%s) - %s", addr, err)
			continue
		}
		n.addPeer(c, v)
	}
	return nil
}

// addPeer adds a peer to the node.
func (n *Node) addPeer(peer proto.NodeClient, version *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	n.peers[peer] = version

	if len(version.PeerList) > 0 {
		go n.bootstrapNetwork(version.PeerList...)
	}

	n.logger.Infof("(%s) Adding peer (%s) - height(%d)", n.listenAddr, version.ListenAddr, version.Height)

}

// deletePeer removes a peer from the node.
func (n *Node) deletePeer(peer proto.NodeClient) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	delete(n.peers, peer)
}

func makeNodeClient(listenAddr string) (proto.NodeClient, error) {
	c, err := grpc.Dial(listenAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return proto.NewNodeClient(c), nil
}

// getVersion is the version of the node.
func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    "blocker-0.1",
		Height:     0,
		ListenAddr: n.listenAddr,
		PeerList:   n.getPeerList(),
	}
}

// canConnectWith returns true if the node can connect with the other node.
func (n *Node) canConnectWith(addr string) bool {
	if n.listenAddr == addr {
		return false
	}

	connectedPeers := n.getPeerList()
	for _, v := range connectedPeers {
		if v == addr {
			return false
		}
	}

	return true
}

// getPeerList returns a list of peers.
func (n *Node) getPeerList() []string {
	n.peerLock.RLock()
	defer n.peerLock.RUnlock()
	var peers []string
	for _, v := range n.peers {
		peers = append(peers, v.ListenAddr)
	}
	return peers
}

// dialRemoteNode connects to a remote node.
func (n *Node) dialRemoteNode(addr string) (proto.NodeClient, *proto.Version, error) {
	c, err := makeNodeClient(addr)
	if err != nil {
		return nil, nil, err
	}
	v, err := c.Handshake(context.Background(), n.getVersion())
	if err != nil {
		return nil, nil, err
	}
	return c, v, nil
}
