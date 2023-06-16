package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/crypto"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
)

func QueueItemHash(address string, currency string, amountOrNftTokenId string) ([]byte, bool) {
	hash := solsha3.SoliditySHA3(
		[]string{"address", "address", "uint256"},
		[]interface{}{
			address,
			currency,
			amountOrNftTokenId,
		},
	)
	return hash, true
}

func NftQueueItemHash(address string, currency string, amountOrNftTokenId string, l2_minted bool) ([]byte, bool) {
	hash := solsha3.SoliditySHA3(
		[]string{"address", "address", "uint256", "bool"},
		[]interface{}{
			address,
			currency,
			amountOrNftTokenId,
			l2_minted,
		},
	)
	return hash, true
}

func QueueHash(queue []Transaction, tx_type string) ([]byte, int, bool) {
	var queue_hash []byte

	var valid_queue [][]byte
	for i := 0; i < len(queue); i++ {
		if queue[i].Type == tx_type {
			cb, ok := QueueItemHash(queue[i].To, queue[i].CurrencyOrNftContractAddress, queue[i].AmountOrNftTokenId)
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

func WithdrawalHash(withdrawal []Transaction) ([]byte, []string, []bool, []string, []string, []int, bool) {
	var withdrawal_hash []byte
	var withdrawal_amounts_or_token_id []string
	var withdrawal_l2_minted []bool
	var withdrawal_addresses []string
	var withdrawal_currency_or_nft_contract []string
	var withdrawal_type []int
	withdrawal_str := ""
	for i := 0; i < len(withdrawal); i++ {
		if withdrawal[i].Type == "withdrawal" {

			withdrawal_str += fmt.Sprintf("%02x", 0)
			withdrawal_str += withdrawal[i].To[2:]
			withdrawal_str += withdrawal[i].CurrencyOrNftContractAddress[2:]
			amt, ok := new(big.Int).SetString(withdrawal[i].AmountOrNftTokenId, 10)
			if !ok {
				return withdrawal_hash, withdrawal_amounts_or_token_id, withdrawal_l2_minted, withdrawal_addresses, withdrawal_currency_or_nft_contract, withdrawal_type, ok
			}
			withdrawal_str += fmt.Sprintf("%064x", amt)
			withdrawal_str += fmt.Sprintf("%02x", 0)

			withdrawal_amounts_or_token_id = append(withdrawal_amounts_or_token_id, withdrawal[i].AmountOrNftTokenId)
			withdrawal_addresses = append(withdrawal_addresses, withdrawal[i].To)
			withdrawal_currency_or_nft_contract = append(withdrawal_currency_or_nft_contract, withdrawal[i].CurrencyOrNftContractAddress)
			withdrawal_l2_minted = append(withdrawal_l2_minted, false)
			withdrawal_type = append(withdrawal_type, 0)
		} else if withdrawal[i].Type == "nft_withdrawal" {

			withdrawal_str += fmt.Sprintf("%02x", 1)
			withdrawal_str += withdrawal[i].To[2:]
			withdrawal_str += withdrawal[i].CurrencyOrNftContractAddress[2:]
			amt, ok := new(big.Int).SetString(withdrawal[i].AmountOrNftTokenId, 10)
			if !ok {
				return withdrawal_hash, withdrawal_amounts_or_token_id, withdrawal_l2_minted, withdrawal_addresses, withdrawal_currency_or_nft_contract, withdrawal_type, ok
			}
			withdrawal_str += fmt.Sprintf("%064x", amt)
			if withdrawal[i].L2Minted {
				withdrawal_str += fmt.Sprintf("%02x", 1)
			} else {
				withdrawal_str += fmt.Sprintf("%02x", 0)
			}

			withdrawal_amounts_or_token_id = append(withdrawal_amounts_or_token_id, withdrawal[i].AmountOrNftTokenId)
			withdrawal_addresses = append(withdrawal_addresses, withdrawal[i].To)
			withdrawal_currency_or_nft_contract = append(withdrawal_currency_or_nft_contract, withdrawal[i].CurrencyOrNftContractAddress)
			withdrawal_l2_minted = append(withdrawal_l2_minted, withdrawal[i].L2Minted)
			withdrawal_type = append(withdrawal_type, 1)
		}
	}
	wa, err := hex.DecodeString(withdrawal_str)
	if err != nil {
		return withdrawal_hash, withdrawal_amounts_or_token_id, withdrawal_l2_minted, withdrawal_addresses, withdrawal_currency_or_nft_contract, withdrawal_type, false
	}
	withdrawal_hash = crypto.Keccak256(wa)

	return withdrawal_hash, withdrawal_amounts_or_token_id, withdrawal_l2_minted, withdrawal_addresses, withdrawal_currency_or_nft_contract, withdrawal_type, true
}

func WithdrawalQueueHash(queue []Transaction) ([]byte, int, []string, []string, []string, bool) {
	var queue_hash []byte
	var amounts []string
	var addresses []string
	var tokens []string

	var valid_queue [][]byte
	for i := 0; i < len(queue); i++ {
		if queue[i].Type == "contract_withdrawal" {
			cb, ok := QueueItemHash(queue[i].To, queue[i].CurrencyOrNftContractAddress, queue[i].AmountOrNftTokenId)
			if !ok {
				return queue_hash, 0, addresses, amounts, tokens, ok
			}
			if queue[i].IsInvalid {
				zero := new(big.Int).SetUint64(0)
				valid_queue = append(valid_queue, zero.Bytes())
			} else {
				valid_queue = append(valid_queue, cb)
			}
			addresses = append(addresses, queue[i].To)
			amounts = append(amounts, queue[i].AmountOrNftTokenId)
			tokens = append(tokens, queue[i].CurrencyOrNftContractAddress)
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
	return queue_hash, len(addresses), addresses, amounts, tokens, true
}

func NftWithdrawalQueueHash(queue []Transaction) ([]byte, int, []string, []string, []string, []bool, bool) {
	var queue_hash []byte
	var amounts []string
	var addresses []string
	var tokens []string
	var l2_minted []bool

	var valid_queue [][]byte
	for i := 0; i < len(queue); i++ {
		if queue[i].Type == "nft_contract_withdrawal" {
			cb, ok := NftQueueItemHash(queue[i].To, queue[i].CurrencyOrNftContractAddress, queue[i].AmountOrNftTokenId, queue[i].L2Minted)
			if !ok {
				return queue_hash, 0, addresses, amounts, tokens, l2_minted, ok
			}
			if queue[i].IsInvalid {
				zero := new(big.Int).SetUint64(0)
				valid_queue = append(valid_queue, zero.Bytes())
			} else {
				valid_queue = append(valid_queue, cb)
			}
			addresses = append(addresses, queue[i].To)
			amounts = append(amounts, queue[i].AmountOrNftTokenId)
			tokens = append(tokens, queue[i].CurrencyOrNftContractAddress)
			l2_minted = append(l2_minted, queue[i].L2Minted)
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
	return queue_hash, len(addresses), addresses, amounts, tokens, l2_minted, true
}

func GetOptimizedNonce(used_lister_nonce []uint) []uint {
	optimized_used_lister_nonce := []uint{}
	sort.Slice(used_lister_nonce, func(i, j int) bool { return used_lister_nonce[i] < used_lister_nonce[j] })
	last_optimized_nonce := uint(0)
	for i, nonce := range used_lister_nonce {
		if uint(i+1) == nonce {
			last_optimized_nonce = nonce
			continue
		} else {
			if last_optimized_nonce != 0 {
				optimized_used_lister_nonce = append([]uint{0}, optimized_used_lister_nonce...)
				optimized_used_lister_nonce = append(optimized_used_lister_nonce, last_optimized_nonce)
				last_optimized_nonce = 0
			}
			optimized_used_lister_nonce = append(optimized_used_lister_nonce, nonce)
		}
	}
	if last_optimized_nonce != 0 {
		optimized_used_lister_nonce = append([]uint{0}, optimized_used_lister_nonce...)
		optimized_used_lister_nonce = append(optimized_used_lister_nonce, last_optimized_nonce)
		last_optimized_nonce = 0
	}
	return optimized_used_lister_nonce
}
