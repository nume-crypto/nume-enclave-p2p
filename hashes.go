package main

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
)

func QueueItemHash(address string, currency string, amount string) ([]byte, bool) {
	hash := solsha3.SoliditySHA3(
		[]string{"address", "address", "uint256"},
		[]interface{}{
			address,
			currency,
			amount,
		},
	)
	return hash, true
}

func QueueHash(queue []NftTransaction) ([]byte, int, bool) {
	var queue_hash []byte

	var valid_queue [][]byte
	for i := 0; i < len(queue); i++ {
		if queue[i].Type == "deposit" {
			tk := new(big.Int).SetUint64(uint64(queue[i].NftTokenId))
			cb, ok := QueueItemHash(queue[i].To, queue[i].NftContractAddress, tk.String())
			if !ok {
				return queue_hash, 0, ok
			}
			valid_queue = append(valid_queue, cb)
		}
	}
	types := []string{}
	values := []interface{}{}
	for _, item := range valid_queue {
		fmt.Println(hex.EncodeToString(item))
		types = append(types, "uint256")
		values = append(values, new(big.Int).SetBytes(item))
	}
	queue_hash = solsha3.SoliditySHA3(
		types,
		values,
	)
	return queue_hash, len(valid_queue), true
}

func WithdrawalHash(withdrawal []NftTransaction) ([]byte, []uint, []bool, []string, []string, bool) {
	var withdrawal_hash []byte
	var withdrawal_amounts []uint
	var withdrawal_l2_minted []bool
	var withdrawal_addresses []string
	var withdrawal_tokens []string
	withdrawal_str := ""
	for i := 0; i < len(withdrawal); i++ {
		if withdrawal[i].Type == "nft_withdrawal" {
			withdrawal_l2_minted = append(withdrawal_l2_minted, true)
			withdrawal_str += withdrawal[i].To[2:]
			withdrawal_str += withdrawal[i].NftContractAddress[2:]
			withdrawal_str += fmt.Sprintf("%064x", withdrawal[i].NftTokenId)
			withdrawal_str += fmt.Sprintf("%02x", 1)
			withdrawal_amounts = append(withdrawal_amounts, withdrawal[i].NftTokenId)
			withdrawal_addresses = append(withdrawal_addresses, withdrawal[i].To)
			withdrawal_tokens = append(withdrawal_tokens, withdrawal[i].NftContractAddress)
		}

	}
	fmt.Println("withdrawal_str", withdrawal_str)
	wa, err := hex.DecodeString(withdrawal_str)
	if err != nil {
		fmt.Println("err", err)
	}
	withdrawal_hash = crypto.Keccak256(wa)

	return withdrawal_hash, withdrawal_amounts, withdrawal_l2_minted, withdrawal_addresses, withdrawal_tokens, true
}

func WithdrawalQueueHash(queue []NftTransaction) ([]byte, int, []string, []uint, []string, bool) {
	var queue_hash []byte
	var amounts []uint
	var addresses []string
	var tokens []string

	var valid_queue [][]byte
	for i := 0; i < len(queue); i++ {
		if queue[i].Type == "contract_withdrawal" {
			tk := new(big.Int).SetUint64(uint64(queue[i].NftTokenId))
			cb, ok := QueueItemHash(queue[i].To, queue[i].NftContractAddress, tk.String())
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
			amounts = append(amounts, queue[i].NftTokenId)
			tokens = append(tokens, queue[i].NftContractAddress)

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
