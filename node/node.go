package node

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/fzft/crypto-prd-blockchain/crypto"
	"github.com/fzft/crypto-prd-blockchain/proto"
	"github.com/fzft/crypto-prd-blockchain/types"
	"github.com/fzft/crypto-prd-blockchain/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"net"
	"sync"
	"time"
)

const blockInterval = 5 * time.Second

type MemPool struct {
	txx *util.KeyValueStore[string, *proto.Transaction]
}

func NewMemPool() *MemPool {
	return &MemPool{
		txx: util.NewKeyValueStore[string, *proto.Transaction](),
	}
}

// Has checks if the transaction is in the mempool.
func (m *MemPool) Has(tx *proto.Transaction) bool {
	hash := hex.EncodeToString(types.HashTransaction(tx))
	return m.txx.Has(hash)
}

// Add adds a transaction to the mempool.
func (m *MemPool) Add(tx *proto.Transaction) bool {
	hash := hex.EncodeToString(types.HashTransaction(tx))
	return m.txx.HasOrAdd(hash, tx)
}

// Len returns the number of transactions in the mempool.
func (m *MemPool) Len() int {
	return m.txx.Len()
}

// Clear Pop all transactions from the mempool.
func (m *MemPool) Clear() []*proto.Transaction {
	return m.txx.PopAllVal()
}

type ServerConfig struct {
	Version    string
	ListenAddr string
	PrivateKey *crypto.PrivateKey
}

type Node struct {
	ServerConfig
	logger *zap.SugaredLogger

	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version

	mempool *MemPool

	proto.UnimplementedNodeServer
}

func New(cfg ServerConfig) *Node {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.TimeKey = ""
	logger, _ := loggerConfig.Build()
	return &Node{
		peers:        make(map[proto.NodeClient]*proto.Version),
		logger:       logger.Sugar(),
		mempool:      NewMemPool(),
		ServerConfig: cfg,
	}
}

// Start ...
func (n *Node) Start(listenAddr string, bootstrapNodes ...string) error {
	n.ListenAddr = listenAddr
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

	if n.PrivateKey != nil {
		go n.validatorLoop()
	}

	return grpcServer.Serve(ln)
}

// Handshake is called when a new peer connects to the node.
func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	c, err := makeNodeClient(v.ListenAddr)
	if err != nil {
		return nil, err
	}

	n.addPeer(c, v)
	return n.getVersion(), nil
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {

	peer, _ := peer.FromContext(ctx)
	hash := hex.EncodeToString(types.HashTransaction(tx))

	if n.mempool.Add(tx) {
		n.logger.Debugw("received transaction", "from", peer.Addr.String(), "hash", hash, "we", n.ListenAddr)
		go func() {
			if err := n.broadcast(tx); err != nil {
				n.logger.Errorf("Error broadcasting transaction - %s", err)
			}
		}()
	}

	return &proto.Ack{}, nil
}

// validatorLoop
func (n *Node) validatorLoop() {
	n.logger.Infow("Starting validator loop", "pubkey", n.PrivateKey.PublicKey().Address(), "blockTime", blockInterval)
	ticker := time.NewTicker(blockInterval)
	for {
		<-ticker.C
		txx := n.mempool.Clear()
		n.logger.Debugw("time to create new block", "lenTx", len(txx))
	}
}

// broadcast
func (n *Node) broadcast(msg any) error {
	for peer := range n.peers {
		switch v := msg.(type) {
		case *proto.Transaction:
			if _, err := peer.HandleTransaction(context.Background(), v); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown message type %T", v)
		}

	}
	return nil
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

	n.logger.Infof("(%s) Adding peer (%s) - height(%d)", n.ListenAddr, version.ListenAddr, version.Height)

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
		ListenAddr: n.ListenAddr,
		PeerList:   n.getPeerList(),
	}
}

// canConnectWith returns true if the node can connect with the other node.
func (n *Node) canConnectWith(addr string) bool {
	if n.ListenAddr == addr {
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
