package util

import (
	"bytes"
	"crypto/sha256"
)

type Hashable interface {
	Hash() []byte
}

type MerkleTree[T Hashable] struct {
	Root *MerkleNode
	Data []T
}

type MerkleNode struct {
	Left, Right *MerkleNode
	Hash        []byte
}

func NewMerkleTree[T Hashable](data []T) *MerkleTree[T] {
	var nodes []*MerkleNode

	for _, datum := range data {
		nodes = append(nodes, NewMerkleNode(datum.Hash(), nil, nil))
	}

	for len(nodes) > 1 {
		level := make([]*MerkleNode, 0)

		for i := 0; i < len(nodes); i += 2 {
			if i+1 < len(nodes) {
				level = append(level, NewMerkleNode([]byte{}, nodes[i], nodes[i+1]))
			} else {
				level = append(level, nodes[i])
			}
		}

		nodes = level
	}

	return &MerkleTree[T]{Root: nodes[0], Data: data}
}

func NewMerkleNode(hash []byte, left, right *MerkleNode) *MerkleNode {
	node := MerkleNode{
		Left:  left,
		Right: right,
		Hash:  hash,
	}

	if left != nil && right != nil {
		hasher := sha256.New()
		hasher.Write(append(left.Hash, right.Hash...))
		node.Hash = hasher.Sum(nil)
	}

	return &node
}

func (tree *MerkleTree[T]) MerkleRoot() []byte {
	return tree.Root.Hash
}

func (tree *MerkleTree[T]) VerifyTree() bool {
	return tree.verifyNode(tree.Root)
}

func (tree *MerkleTree[T]) verifyNode(node *MerkleNode) bool {
	if node.Left == nil && node.Right == nil {
		return true
	}

	if node.Left != nil && node.Right != nil {
		hasher := sha256.New()
		hasher.Write(append(node.Left.Hash, node.Right.Hash...))
		expectedHash := hasher.Sum(nil)

		return bytes.Equal(node.Hash, expectedHash) && tree.verifyNode(node.Left) && tree.verifyNode(node.Right)
	}

	return false
}
