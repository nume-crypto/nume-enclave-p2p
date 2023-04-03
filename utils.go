package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"time"
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

func GetLeafHash(pub_key string, root string) ([]byte, bool) {
	var hash []byte
	hFunc := NewMiMC()
	cb, ok := new(big.Int).SetString(pub_key, 16)
	if !ok {
		return hash, ok
	}
	root_bi, ok := new(big.Int).SetString(root, 16)
	if !ok {
		return hash, ok
	}

	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	hFunc.Write(root_bi.Bytes())
	hash = hFunc.Sum(nil)

	hFunc.Reset()
	return hash, true
}

func NestedMapsEqual(m1, m2 map[string]map[uint]string) bool {
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

func MapsEqual(m1, m2 map[uint]string) bool {
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

func GetBalancesRoot(balances map[uint]string, max_num_balances int) (string, bool) {
	hFunc := NewMiMC()
	var balances_data = make([][]byte, max_num_balances)
	for i := 0; i < max_num_balances; i++ {
		if val, ok := balances[uint(i)]; ok {
			cb, ok := new(big.Int).SetString(val, 16)
			if !ok {
				return "", ok
			}
			hFunc.Write(cb.Bytes())
			balances_data[i] = hFunc.Sum(nil)
			hFunc.Reset()
		} else {
			cb, ok := new(big.Int).SetString("0", 16)
			if !ok {
				return "", ok
			}
			hFunc.Write(cb.Bytes())
			balances_data[i] = hFunc.Sum(nil)
			hFunc.Reset()
		}
	}
	balances_tree := NewMerkleTree(balances_data, hFunc)
	return hex.EncodeToString(balances_tree.Root), true
}
