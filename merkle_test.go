package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"

	solsha3 "github.com/miguelmota/go-solidity-sha3"
)

func TestNewMerkleTree(t *testing.T) {
	for j := 0; j < 128; j++ {
		val_hash := [][]byte{}
		for i := 0; i < 64; i++ {
			hash := solsha3.SoliditySHA3(
				[]string{"address", "uint256", "uint256"},
				[]interface{}{
					"0x0000000000000000000000000000000000000000",
					strconv.FormatUint(uint64(i), 10),
					strconv.FormatUint(uint64(i+1), 10),
				},
			)
			val_hash = append(val_hash, hash[:])
		}
		tree := NewMerkleTree(val_hash)
		for i := 0; i < 8; i++ {
			v := tree.Verify(i)
			if !v {
				fmt.Println("Invalid Verify")
				t.Errorf("Expected %t, got %t", true, v)
			}
			proof, _ := tree.Proof(i)
			v = tree.VerifyProof(proof, i)
			if !v {
				fmt.Println("Invalid VerifyProof")
				t.Errorf("Expected %t, got %t", true, v)
			}
		}
		tree_sync := NewMerkleTreeSync(val_hash)

		desired_root := "65c8f695dd28fe7b133794a2c70aff09f12125d4dfdd4ea6b80af49c9e6079c9"
		if hex.EncodeToString(tree.Root) != desired_root {
			t.Errorf("Expected %s, got %s", desired_root, hex.EncodeToString(tree.Root))
			return
		}
		if hex.EncodeToString(tree_sync.Root) != desired_root {
			t.Errorf("Expected %s, got %s", desired_root, hex.EncodeToString(tree_sync.Root))
			return
		}
	}
}

func TestNewMerkleTreeUpdate(t *testing.T) {
	val_hash := [][]byte{}
	val_expected_updated_hash := [][]byte{}
	for i := 0; i < 64; i++ {
		hash := solsha3.SoliditySHA3(
			[]string{"address", "uint256", "uint256"},
			[]interface{}{
				"0x0000000000000000000000000000000000000000",
				strconv.FormatUint(uint64(i), 10),
				strconv.FormatUint(uint64(i+1), 10),
			},
		)
		if i >= 10 && i < 50 {
			expected_updated_hash := solsha3.SoliditySHA3(
				[]string{"address", "uint256", "uint256"},
				[]interface{}{
					"0x0000000000000000000000000000000000000000",
					strconv.FormatUint(uint64(i+10), 10),
					strconv.FormatUint(uint64(i+1+10), 10),
				},
			)
			val_expected_updated_hash = append(val_expected_updated_hash, expected_updated_hash[:])
		} else {
			val_expected_updated_hash = append(val_expected_updated_hash, hash[:])
		}
		val_hash = append(val_hash, hash[:])
	}
	expected_update_tree := NewMerkleTree(val_expected_updated_hash)

	for j := 0; j < 128; j++ {
		tree := NewMerkleTree(val_hash)
		for i := 10; i < 50; i++ {
			hash := solsha3.SoliditySHA3(
				[]string{"address", "uint256", "uint256"},
				[]interface{}{
					"0x0000000000000000000000000000000000000000",
					strconv.FormatUint(uint64(i+10), 10),
					strconv.FormatUint(uint64(i+1+10), 10),
				},
			)
			tree.UpdateLeaf(i, hex.EncodeToString(hash[:]))
		}
		if !bytes.Equal(tree.Root, expected_update_tree.Root) {
			t.Errorf("Expected %s, got %s", hex.EncodeToString(expected_update_tree.Root), hex.EncodeToString(tree.Root))
			return
		}
	}

}
