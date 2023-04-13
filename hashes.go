package main

import (
	"fmt"
	"math/big"

	solsha3 "github.com/miguelmota/go-solidity-sha3"
)

func DigitalSignatureMessage(from string, to string, currency string, amount string, nonce uint64, block_number int64) string {
	if len(to) != 64 {
		to += "000000000000000000000000"
	}
	if len(from) != 64 {
		from += "000000000000000000000000"
	}
	if len(currency) != 64 {
		currency += "000000000000000000000000"
	}
	amt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%064s", from[2:]) + fmt.Sprintf("%064s", to[2:]) + fmt.Sprintf("%064s", currency[2:]) + fmt.Sprintf("%064x", amt) + fmt.Sprintf("%064x", nonce) + fmt.Sprintf("%064x", block_number)
}

func QueueItemHash(pub_key string, token_id string, amount string) ([]byte, bool) {
	var queue_hash []byte

	cb1, ok := new(big.Int).SetString(pub_key, 16)
	if !ok {
		return queue_hash, ok
	}

	cb2, ok := new(big.Int).SetString(token_id, 16)
	if !ok {
		return queue_hash, ok
	}

	cb3, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return queue_hash, ok
	}

	queue_hash = solsha3.SoliditySHA3(
		[]string{"uint256", "uint256", "uint256"},
		[]interface{}{
			cb1,
			cb2,
			cb3,
		},
	)
	return queue_hash, true
}

func QueueHash(queue []Transaction) ([]byte, int, bool) {
	var queue_hash []byte

	var valid_queue [][]byte
	for i := 0; i < len(queue); i++ {
		if queue[i].Type == "deposit" {
			cb, ok := QueueItemHash(queue[i].To, queue[i].Currency, queue[i].Amount)
			if !ok {
				return queue_hash, 0, ok
			}
			valid_queue = append(valid_queue, cb)
		}
	}
	types := []string{}
	values := []interface{}{}
	for _, item := range valid_queue {
		types = append(types, "uint256")
		values = append(values, new(big.Int).SetBytes(item))
	}
	queue_hash = solsha3.SoliditySHA3(
		types,
		values,
	)
	return queue_hash, len(valid_queue), true
}

func WithdrawalItemHash(amount string, token_id uint, address string) ([]byte, bool) {
	var withdrawal_hash []byte

	cb1, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return withdrawal_hash, ok
	}
	cb2 := new(big.Int).SetUint64(uint64(token_id))
	cb3, ok := new(big.Int).SetString(address, 16)
	if !ok {
		return withdrawal_hash, ok
	}

	withdrawal_hash = solsha3.SoliditySHA3(
		[]string{"uint256", "uint256", "uint256"},
		[]interface{}{
			cb1,
			cb2,
			cb3,
		},
	)

	return withdrawal_hash, true
}

func WithdrawalHash(withdrawal []Transaction) ([]byte, []string, []string, []string, bool) {
	var withdrawal_hash []byte
	var withdrawal_amounts []string
	var withdrawal_addresses []string
	var withdrawal_tokens []string

	var valid_withdrawal [][]byte
	for i := 0; i < len(withdrawal); i++ {
		if withdrawal[i].Type == "withdrawal" {
			cb, ok := new(big.Int).SetString(withdrawal[i].Amount, 10)
			if !ok {
				return withdrawal_hash, withdrawal_amounts, withdrawal_addresses, withdrawal_tokens, ok
			}
			valid_withdrawal = append(valid_withdrawal, cb.Bytes())
			cb, ok = new(big.Int).SetString(withdrawal[i].Currency, 10)
			if !ok {
				return withdrawal_hash, withdrawal_amounts, withdrawal_addresses, withdrawal_tokens, ok
			}
			valid_withdrawal = append(valid_withdrawal, cb.Bytes())
			cb, ok = new(big.Int).SetString(withdrawal[i].To+"000000000000000000000000", 16)
			if !ok {
				return withdrawal_hash, withdrawal_amounts, withdrawal_addresses, withdrawal_tokens, ok
			}
			valid_withdrawal = append(valid_withdrawal, cb.Bytes())
			withdrawal_amounts = append(withdrawal_amounts, withdrawal[i].Amount)
			withdrawal_addresses = append(withdrawal_addresses, withdrawal[i].To)
			withdrawal_tokens = append(withdrawal_tokens, withdrawal[i].Currency)
		}

	}
	types := []string{}
	values := []interface{}{}
	for _, item := range valid_withdrawal {
		types = append(types, "uint256")
		values = append(values, new(big.Int).SetBytes(item))
	}
	withdrawal_hash = solsha3.SoliditySHA3(
		types,
		values,
	)

	return withdrawal_hash, withdrawal_amounts, withdrawal_addresses, withdrawal_tokens, true
}

func WithdrawalQueueItemHash(pub_key string, to string, token_id string, amount string) ([]byte, bool) {
	var queue_hash []byte
	cb1, ok := new(big.Int).SetString(pub_key, 16)
	if !ok {
		return queue_hash, ok
	}
	cb2, ok := new(big.Int).SetString(to+"000000000000000000000000", 16)
	if !ok {
		return queue_hash, ok
	}
	cb3, ok := new(big.Int).SetString(token_id, 16)
	if !ok {
		return queue_hash, ok
	}
	cb4, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return queue_hash, ok
	}
	queue_hash = solsha3.SoliditySHA3(
		[]string{"uint256", "uint256", "uint256", "uint256"},
		[]interface{}{
			cb1,
			cb2,
			cb3,
			cb4,
		},
	)
	return queue_hash, true
}

func WithdrawalQueueHash(queue []Transaction) ([]byte, int, []string, []string, []string, []string, bool) {
	var queue_hash []byte
	var amounts []string
	var addresses []string
	var tokens []string
	var bls_keys []string

	var valid_queue [][]byte
	for i := 0; i < len(queue); i++ {
		if queue[i].Type == "contract_withdrawal" {
			cb, ok := WithdrawalQueueItemHash(queue[i].From, queue[i].To, queue[i].Currency, queue[i].Amount)
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
			tokens = append(tokens, queue[i].Currency)
			bls_keys = append(bls_keys, queue[i].From)

		}
	}
	types := []string{}
	values := []interface{}{}
	for _, item := range valid_queue {
		types = append(types, "uint256")
		values = append(values, new(big.Int).SetBytes(item))
	}
	queue_hash = solsha3.SoliditySHA3(
		types,
		values,
	)

	return queue_hash, len(addresses), addresses, amounts, tokens, bls_keys, true
}
