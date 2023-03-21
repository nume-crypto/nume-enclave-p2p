//go:build !test
// +build !test

package main

import (
	"bytes"
	"hash"
	"math"
	"sync"
)

type MerkleTree struct {
	Root   []byte
	Nodes  [][]MerkleNode
	hFunc  hash.Hash
	height int
}

type MerkleNode struct {
	left  *MerkleNode
	right *MerkleNode
	Data  []byte
}

func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := MerkleNode{}

	if left == nil && right == nil {
		node.Data = data[:]
	} else {
		hFunc := NewMiMC()
		hFunc.Write(left.Data)
		hFunc.Sum(nil)
		hFunc.Write(right.Data)
		hash := hFunc.Sum(nil)
		node.Data = hash[:]
	}

	node.left = left
	node.right = right

	return &node
}

func NewMerkleTree(data [][]byte, h hash.Hash) *MerkleTree {
	var tree MerkleTree
	tree.hFunc = h

	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	merkle_tree_height := int(math.Log2(float64(len(data)))) + 1
	var merkle_tree_org = make([][]MerkleNode, merkle_tree_height)
	leaves_in_level := len(data)
	for i := range merkle_tree_org {
		merkle_tree_org[i] = make([]MerkleNode, leaves_in_level)
		leaves_in_level /= 2
	}

	var nodes []MerkleNode = make([]MerkleNode, len(data))
	var wg sync.WaitGroup
	for i, d := range data {
		wg.Add(1)
		go func(i int, d []byte) {
			node := NewMerkleNode(nil, nil, d)
			nodes[i] = *node
			merkle_tree_org[0][i] = *node
			wg.Done()
		}(i, d)
	}
	leaves_in_level = len(data)
	for i := 1; i < merkle_tree_height; i++ {
		for j := 0; j < leaves_in_level-1; j += 2 {
			wg.Add(1)
			go func(i int, j int, n int) {
				if i == 0 {
					for {
						if nodes[j].Data != nil && nodes[j+1].Data != nil {
							node1 := &nodes[j]
							node2 := &nodes[j]
							if j+1 < n {
								node2 = &nodes[j+1]
							}
							node := NewMerkleNode(node1, node2, nil)
							merkle_tree_org[i][j/2] = *node
							break
						}
					}
				} else {
					for {
						if merkle_tree_org[i-1][j].Data != nil && merkle_tree_org[i-1][j+1].Data != nil {
							node1 := &merkle_tree_org[i-1][j]
							node2 := &merkle_tree_org[i-1][j]
							if j+1 < n {
								node2 = &merkle_tree_org[i-1][j+1]
							}
							node := NewMerkleNode(node1, node2, nil)
							merkle_tree_org[i][j/2] = *node
							break
						}
					}

				}
				wg.Done()
			}(i, j, leaves_in_level)
		}
		leaves_in_level /= 2
	}
	wg.Wait()
	tree.Nodes = merkle_tree_org
	tree.Root = merkle_tree_org[merkle_tree_height-1][0].Data[:]
	tree.height = merkle_tree_height
	return &tree
}

func (tree MerkleTree) Proof(index int) [][]byte {
	var proof [][]byte
	position := float64(index)
	if index > -1 {
		for i := 0; i < tree.height-1; i++ {
			var neighbour MerkleNode
			if int64(position)%2 == 0 {
				neighbour = tree.Nodes[i][int64(position+1)]
				position = math.Floor(position / 2)
				proof = append(proof, neighbour.Data)
			} else {
				neighbour = tree.Nodes[i][int64(position-1)]
				position = math.Floor((position - 1) / 2)
				proof = append(proof, neighbour.Data)
			}
		}
	}
	return proof
}

func (tree MerkleTree) Verify(index int) bool {
	hash := tree.Nodes[0][index].Data
	position := float64(index)
	if index > -1 {
		for i := 0; i < tree.height-1; i++ {
			var neighbour []byte
			if int64(position)%2 == 0 {
				neighbour = tree.Nodes[i][int64(position+1)].Data[:]
				position = math.Floor(position / 2)
				hFunc := NewMiMC()
				prevHashes := append(hash, neighbour...)
				hFunc.Write(prevHashes)
				hash = hFunc.Sum(nil)
				hFunc.Reset()
			} else {
				neighbour = tree.Nodes[i][int64(position-1)].Data[:]
				position = math.Floor((position - 1) / 2)
				hFunc := NewMiMC()
				prevHashes := append(neighbour, hash...)
				hFunc.Write(prevHashes)
				hash = hFunc.Sum(nil)
				hFunc.Reset()
			}
		}
	}
	return bytes.Equal(hash, tree.Root)
}
