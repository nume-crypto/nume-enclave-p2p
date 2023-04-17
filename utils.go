package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"time"

	solsha3 "github.com/miguelmota/go-solidity-sha3"
)

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func GetLeafHash(address string, root string) []byte {
	hash := solsha3.SoliditySHA3(
		[]string{"address", "bytes32"},
		[]interface{}{
			address,
			root,
		},
	)
	return hash
}

func NestedMapsEqual(m1, m2 map[string]map[string]string) bool {
	if len(m1) != len(m2) {
		fmt.Println("len(m1)", len(m1), "len(m2)", len(m2))
		return false
	}
	for k, v1 := range m1 {
		if v2, ok := m2[k]; !ok || !MapsEqual(v1, v2) {
			fmt.Println("v1", v1, "v2", v2)
			return false
		}
	}
	return reflect.DeepEqual(m1, m2)
}

func MapsEqual(m1, m2 map[string]string) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v1 := range m1 {
		if v2, ok := m2[k]; !ok || v1 != v2 {
			return false
		}
	}
	return reflect.DeepEqual(m1, m2)
}

func GetBalancesRoot(balances map[string]string, user_balance_order []string, max_num_balances int) (string, bool) {

	balances_tree := &MerkleTree{}
	var balances_data = make([][]byte, max_num_balances)
	for i := 0; i < max_num_balances; i++ {
		if i < len(balances) {
			cb2, ok := new(big.Int).SetString(balances[user_balance_order[i]], 10)
			if !ok {
				return "", ok
			}
			hash := solsha3.SoliditySHA3(
				[]string{"address", "uint256"},
				[]interface{}{
					user_balance_order[i],
					cb2,
				},
			)
			balances_data[i] = hash
		} else {
			hash := solsha3.SoliditySHA3(
				[]string{"address", "uint256"},
				[]interface{}{
					"0x0000000000000000000000000000000000000000",
					"0",
				},
			)
			balances_data[i] = hash
		}
	}
	balances_tree = NewMerkleTree(balances_data)
	return hex.EncodeToString(balances_tree.Root), true
}
