package main

import (
	"math/big"
)

func DigitalSignatureMessageHash(from string, to string, currency string, amount string, nonce string) ([]byte, bool) {
	hFunc := NewMiMC()
	var hashed_message []byte
	cb, ok := new(big.Int).SetString(from, 16)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb, ok = new(big.Int).SetString(to, 16)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb, ok = new(big.Int).SetString(currency, 10)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb, ok = new(big.Int).SetString(amount, 10)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb, ok = new(big.Int).SetString(nonce, 10)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hashed_message = hFunc.Sum(nil)
	hFunc.Reset()
	return hashed_message, true
}

func LeafHash(pub_key string, balance_root string) ([]byte, bool) {
	hFunc := NewMiMC()
	var hashed_message []byte
	cb, ok := new(big.Int).SetString(pub_key, 16)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb, ok = new(big.Int).SetString(balance_root, 16)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hashed_message = hFunc.Sum(nil)
	hFunc.Reset()
	return hashed_message, true
}

func G1Hash(g1_keys [2]string) ([]byte, bool) {
	hFunc := NewMiMC()
	var hashed_message []byte
	cb, ok := new(big.Int).SetString(g1_keys[0], 16)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb, ok = new(big.Int).SetString(g1_keys[1], 16)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hashed_message = hFunc.Sum(nil)
	hFunc.Reset()
	return hashed_message, true
}
