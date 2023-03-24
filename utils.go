package main

import (
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
		return false
	}
	for k, v1 := range m1 {
		if v2, ok := m2[k]; !ok || !MapsEqual(v1, v2) {
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
