// Package blockchain represents the core logic for blockchain operations such as managing blocks,
// transactions, and their interrelationships like merkle trees and proof of work.
package blockchain

import (
	"crypto/sha256"
	"log"
)

// MerkleTree represents a Merkle tree for efficient and secure verification of large data structures.
type MerkleTree struct {
	RootNode *MerkleNode // The root node of the Merkle tree
}

// MerkleNode represents a single node within a Merkle tree.
type MerkleNode struct {
	Left  *MerkleNode // Pointer to the left child node
	Right *MerkleNode // Pointer to the right child node
	Data  []byte      // Data hash stored in the node
}

// NewMerkleNode creates a new Merkle tree node from left and right child nodes and the node's data.
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := MerkleNode{}

	// If it's a leaf node, hash the data. Otherwise, hash the concatenation of child nodes' data.
	if left == nil && right == nil {
		hash := sha256.Sum256(data) // Hashing the data for leaf nodes
		node.Data = hash[:]
	} else {
		prevHashes := append(left.Data, right.Data...) // Concatenating the hashes of child nodes
		hash := sha256.Sum256(prevHashes)              // Hashing the concatenated hashes
		node.Data = hash[:]
	}

	node.Left = left
	node.Right = right

	return &node
}

// NewMerkleTree creates a new Merkle tree using a slice of data.
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	// Creating a leaf node for each data element
	for _, dat := range data {
		node := NewMerkleNode(nil, nil, dat)
		nodes = append(nodes, *node)
	}

	// Panic if there are no nodes, as a Merkle tree cannot be created
	if len(nodes) == 0 {
		log.Panic("No Merkle nodes present")
	}

	// Constructing the tree layer by layer
	for len(nodes) > 1 {
		// Duplicate the last node if the number of nodes is odd
		if len(nodes)%2 != 0 {
			nodes = append(nodes, nodes[len(nodes)-1])
		}

		var level []MerkleNode
		// Creating new nodes from pairs of existing nodes
		for i := 0; i < len(nodes); i += 2 {
			node := NewMerkleNode(&nodes[i], &nodes[i+1], nil)
			level = append(level, *node)
		}

		nodes = level // Move up one level in the tree
	}

	tree := MerkleTree{&nodes[0]} // The root of the tree is the only node at the top level

	return &tree
}
