package main

import (
	"fmt"
	"math/big"
)

func DigitalSignatureMessage(from string, to string, currency uint, amount string, nonce uint64, block_number int64) string {
	if len(to) != 64 {
		to += "000000000000000000000000"
	}
	amt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return ""
	}
	return from + to + fmt.Sprintf("%064x", currency) + fmt.Sprintf("%064x", amt) + fmt.Sprintf("%064x", nonce) + fmt.Sprintf("%064x", block_number)
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

func QueueItemHash(pub_key string, token_id uint, amount string) ([]byte, bool) {
	var queue_hash []byte
	hFunc := NewMiMC()
	cb, ok := new(big.Int).SetString(pub_key, 16)
	if !ok {
		return queue_hash, ok
	}

	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)

	cb = new(big.Int).SetUint64(uint64(token_id))
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)

	cb, ok = new(big.Int).SetString(amount, 10)
	if !ok {
		return queue_hash, ok
	}
	hFunc.Write(cb.Bytes())
	queue_hash = hFunc.Sum(nil)

	hFunc.Reset()
	return queue_hash, true
}

func QueueHash(queue []Transaction) ([]byte, int, bool) {
	var queue_hash []byte
	hFunc := NewMiMC()
	var valid_queue [][]byte
	for i := 0; i < len(queue); i++ {
		if queue[i].Type == "deposit" {
			cb, ok := QueueItemHash(queue[i].To, queue[i].CurrencyTokenOrder, queue[i].Amount)
			if !ok {
				return queue_hash, 0, ok
			}
			valid_queue = append(valid_queue, cb)
		}
	}
	for _, item := range valid_queue {
		hFunc.Write(item)
		queue_hash = hFunc.Sum(nil)
	}
	hFunc.Reset()
	return queue_hash, len(valid_queue), true
}

func WithdrawalItemHash(amount string, token_id uint, address string) ([]byte, bool) {
	var withdrawal_hash []byte
	hFunc := NewMiMC()
	cb, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return withdrawal_hash, ok
	}
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb = new(big.Int).SetUint64(uint64(token_id))
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb, ok = new(big.Int).SetString(address, 16)
	if !ok {
		return withdrawal_hash, ok
	}

	hFunc.Write(cb.Bytes())
	withdrawal_hash = hFunc.Sum(nil)

	hFunc.Reset()
	return withdrawal_hash, true
}

func WithdrawalHash(withdrawal []Transaction) ([]byte, []string, []string, []uint, bool) {
	var withdrawal_hash []byte
	var withdrawal_amounts []string
	var withdrawal_addresses []string
	var withdrawal_tokens []uint

	hFunc := NewMiMC()
	var valid_withdrawal [][]byte
	for i := 0; i < len(withdrawal); i++ {
		if withdrawal[i].Type == "withdrawal" {
			cb, ok := new(big.Int).SetString(withdrawal[i].Amount, 10)
			if !ok {
				return withdrawal_hash, withdrawal_amounts, withdrawal_addresses, withdrawal_tokens, ok
			}
			valid_withdrawal = append(valid_withdrawal, cb.Bytes())
			cb = new(big.Int).SetUint64(uint64(withdrawal[i].CurrencyTokenOrder))
			valid_withdrawal = append(valid_withdrawal, cb.Bytes())
			cb, ok = new(big.Int).SetString(withdrawal[i].To+"000000000000000000000000", 16)
			if !ok {
				return withdrawal_hash, withdrawal_amounts, withdrawal_addresses, withdrawal_tokens, ok
			}
			valid_withdrawal = append(valid_withdrawal, cb.Bytes())
			withdrawal_amounts = append(withdrawal_amounts, withdrawal[i].Amount)
			withdrawal_addresses = append(withdrawal_addresses, withdrawal[i].To)
			withdrawal_tokens = append(withdrawal_tokens, withdrawal[i].CurrencyTokenOrder)
		}

	}
	for _, item := range valid_withdrawal {
		hFunc.Write(item)
		withdrawal_hash = hFunc.Sum(nil)
	}
	hFunc.Reset()
	return withdrawal_hash, withdrawal_amounts, withdrawal_addresses, withdrawal_tokens, true
}

func WithdrawalQueueItemHash(pub_key string, to string, token_id uint, amount string) ([]byte, bool) {
	var queue_hash []byte
	hFunc := NewMiMC()
	cb, ok := new(big.Int).SetString(pub_key, 16)
	if !ok {
		return queue_hash, ok
	}

	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb, ok = new(big.Int).SetString(to+"000000000000000000000000", 16)
	if !ok {
		return queue_hash, ok
	}

	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)

	cb = new(big.Int).SetUint64(uint64(token_id))
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)

	cb, ok = new(big.Int).SetString(amount, 10)
	if !ok {
		return queue_hash, ok
	}
	hFunc.Write(cb.Bytes())
	queue_hash = hFunc.Sum(nil)

	hFunc.Reset()
	return queue_hash, true
}

func WithdrawalQueueHash(queue []Transaction) ([]byte, int, []string, []string, []uint, []string, bool) {
	var queue_hash []byte
	var amounts []string
	var addresses []string
	var tokens []uint
	var bls_keys []string

	hFunc := NewMiMC()
	var valid_queue [][]byte
	for i := 0; i < len(queue); i++ {
		if queue[i].Type == "contract_withdrawal" {
			cb, ok := WithdrawalQueueItemHash(queue[i].From, queue[i].To, queue[i].CurrencyTokenOrder, queue[i].Amount)
			if !ok {
				return queue_hash, 0, addresses, amounts, tokens, bls_keys, ok
			}
			if queue[i].IsInvalid {
				zero := new(big.Int).SetUint64(0)
				valid_queue = append(valid_queue, zero.Bytes())
			} else {
				valid_queue = append(valid_queue, cb)
			}
			addresses = append(addresses, queue[i].To)
			amounts = append(amounts, queue[i].Amount)
			tokens = append(tokens, queue[i].CurrencyTokenOrder)
			bls_keys = append(bls_keys, queue[i].From)

		}
	}
	for _, item := range valid_queue {
		hFunc.Write(item)
		queue_hash = hFunc.Sum(nil)
	}
	hFunc.Reset()
	return queue_hash, len(addresses), addresses, amounts, tokens, bls_keys, true
}
