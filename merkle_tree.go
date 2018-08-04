package main

import (
	"crypto/sha256"
	"math"
)

// MerkleNode is the structure of Merkle node
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

// CreateNode creates a new Merkle node
func CreateNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := MerkleNode{}

	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		node.Data = hash[:]
	} else {
		prevHash := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHash)
		node.Data = hash[:]
	}

	node.Left = left
	node.Right = right

	return &node
}

// MerkleTree is the structure of Merkle tree
type MerkleTree struct {
	Root *MerkleNode
}

// CreateTree creates a new Merkle tree
func CreateTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, datum := range data {
		node := CreateNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}

	height := int(math.Log2(float64(len(nodes))))

	for i := 0; i < height; i++ {
		var newLevel []MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			node := CreateNode(&nodes[j], &nodes[j+1], nil)
			newLevel = append(newLevel, *node)
		}

		nodes = newLevel
	}

	tree := MerkleTree{&nodes[0]}

	return &tree
}
