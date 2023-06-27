package main

import (
	"bytes"
	"encoding/hex"
	"math"
	"math/big"
	"sync"

	solsha3 "github.com/miguelmota/go-solidity-sha3"
)

type MerkleTree struct {
	Root   []byte
	Nodes  [][]MerkleNode
	height int
}

type MerkleNode struct {
	left  *MerkleNode
	right *MerkleNode
	Data  []byte
}

func newMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := MerkleNode{}

	if left == nil && right == nil {
		node.Data = data[:]
	} else {
		hash := solsha3.SoliditySHA3(
			[]string{"uint256", "uint256"},
			[]interface{}{
				new(big.Int).SetBytes(left.Data),
				new(big.Int).SetBytes(right.Data),
			},
		)
		node.Data = hash[:]
	}

	node.left = left
	node.right = right

	return &node
}

func NewMerkleTree(data [][]byte) *MerkleTree {
	var tree MerkleTree

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
			node := newMerkleNode(nil, nil, d)
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
							node := newMerkleNode(node1, node2, nil)
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
							node := newMerkleNode(node1, node2, nil)
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
	// fmt.Println("Merkle Tree: ")
	// for i := 0; i < merkle_tree_height; i++ {
	// 	for j := 0; j < len(merkle_tree_org[i]); j++ {
	// 		fmt.Println(hex.EncodeToString(merkle_tree_org[i][j].Data))
	// 	}
	// }
	// fmt.Println("Merkle Tree Height: ", merkle_tree_height)

	tree.Nodes = merkle_tree_org
	tree.Root = merkle_tree_org[merkle_tree_height-1][0].Data[:]
	tree.height = merkle_tree_height
	return &tree
}

func NewMerkleTreeSync(data [][]byte) *MerkleTree {
	var tree MerkleTree

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
	for i, d := range data {
		node := newMerkleNode(nil, nil, d)
		nodes[i] = *node
		merkle_tree_org[0][i] = *node
	}
	leaves_in_level = len(data)
	for i := 1; i < merkle_tree_height; i++ {
		for j := 0; j < leaves_in_level-1; j += 2 {
			if i == 0 {
				node1 := &nodes[j]
				node2 := &nodes[j]
				if j+1 < leaves_in_level {
					node2 = &nodes[j+1]
				}
				node := newMerkleNode(node1, node2, nil)
				merkle_tree_org[i][j/2] = *node
			} else {
				node1 := &merkle_tree_org[i-1][j]
				node2 := &merkle_tree_org[i-1][j]
				if j+1 < leaves_in_level {
					node2 = &merkle_tree_org[i-1][j+1]
				}
				node := newMerkleNode(node1, node2, nil)
				merkle_tree_org[i][j/2] = *node
			}
		}
		leaves_in_level /= 2
	}
	tree.Nodes = merkle_tree_org
	tree.Root = merkle_tree_org[merkle_tree_height-1][0].Data[:]
	tree.height = merkle_tree_height
	return &tree
}

func (tree MerkleTree) Proof(index int) ([][]byte, []int64) {
	var proof [][]byte
	var helper []int64
	position := float64(index)
	if index > -1 {
		for i := 0; i < tree.height-1; i++ {
			var neighbour MerkleNode
			if int64(position)%2 == 0 {
				neighbour = tree.Nodes[i][int64(position+1)]
				position = math.Floor(position / 2)
				proof = append(proof, neighbour.Data)
				helper = append(helper, 1)
			} else {
				neighbour = tree.Nodes[i][int64(position-1)]
				position = math.Floor((position - 1) / 2)
				proof = append(proof, neighbour.Data)
				helper = append(helper, 0)
			}
		}
	}
	return proof, helper
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
				hash = solsha3.SoliditySHA3(
					[]string{"uint256", "uint256"},
					[]interface{}{
						new(big.Int).SetBytes(hash),
						new(big.Int).SetBytes(neighbour),
					},
				)
			} else {
				neighbour = tree.Nodes[i][int64(position-1)].Data[:]
				position = math.Floor((position - 1) / 2)
				hash = solsha3.SoliditySHA3(
					[]string{"uint256", "uint256"},
					[]interface{}{
						new(big.Int).SetBytes(neighbour),
						new(big.Int).SetBytes(hash),
					},
				)
			}
		}
	}
	return bytes.Equal(hash, tree.Root)
}

func (tree MerkleTree) UpdateLeaf(index int, new_leaf string) (string, error) {
	hash, err := hex.DecodeString(new_leaf)
	if err != nil {
		return "", err
	}
	position := float64(index)
	tree.Nodes[0][index].Data = hash
	if index > -1 {
		for i := 0; i < tree.height-1; i++ {
			var neighbour []byte
			if int64(position)%2 == 0 {
				neighbour = tree.Nodes[i][int64(position+1)].Data[:]
				position = math.Floor(position / 2)
				hash = solsha3.SoliditySHA3(
					[]string{"uint256", "uint256"},
					[]interface{}{
						new(big.Int).SetBytes(hash),
						new(big.Int).SetBytes(neighbour),
					},
				)
				tree.Nodes[i+1][int64(position)].Data = hash
			} else {
				neighbour = tree.Nodes[i][int64(position-1)].Data[:]
				position = math.Floor((position - 1) / 2)
				hash = solsha3.SoliditySHA3(
					[]string{"uint256", "uint256"},
					[]interface{}{
						new(big.Int).SetBytes(neighbour),
						new(big.Int).SetBytes(hash),
					},
				)
				tree.Nodes[i+1][int64(position)].Data = hash
			}
		}
	}
	copy(tree.Root, hash)
	return hex.EncodeToString(hash), nil
}

func (tree MerkleTree) VerifyProof(proof [][]byte, index int) bool {
	hash := tree.Nodes[0][index].Data
	position := float64(index)
	for i := 0; i < tree.height-1; i++ {
		var neighbour []byte
		if int64(position)%2 == 0 {
			neighbour = proof[i]
			position = math.Floor(position / 2)
			hash = solsha3.SoliditySHA3(
				[]string{"uint256", "uint256"},
				[]interface{}{
					new(big.Int).SetBytes(hash),
					new(big.Int).SetBytes(neighbour),
				},
			)
		} else {
			neighbour = proof[i]
			position = math.Floor((position - 1) / 2)
			hash = solsha3.SoliditySHA3(
				[]string{"uint256", "uint256"},
				[]interface{}{
					new(big.Int).SetBytes(neighbour),
					new(big.Int).SetBytes(hash),
				},
			)
		}
	}

	return bytes.Equal(hash, tree.Root)
}
