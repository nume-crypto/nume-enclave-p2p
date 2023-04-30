package main

import (
	"fmt"
	"strconv"
)

func CheckNonce(last_nonce, current_nonce uint64) bool {
	return last_nonce < current_nonce
}

type TransferTransaction struct {
	ID                uint
	EncryptionOutputs []string `json:"encryptionOutputs"`
}

type TransactionPublicData struct {
	Transactions []TransferTransaction `json:"transactions"`
}

func TransitionState(state_balances map[string]map[string]bool, transactions []NftTransaction) (map[string]map[string]bool, uint, map[string]bool, map[string]uint64, error) {
	users_updated_map := make(map[string]bool)
	settlement_type := uint(0)
	user_nonce_tracker := make(map[string]uint64)
	has_deposit := false
	has_withdraw := false
	has_contract_withdrawal := false
	for i, transaction := range transactions {
		if transaction.Type != "nft_deposit" && transaction.Type != "nft_mint" {
			verified, err := VerifyData(transaction)
			if !verified || err != nil {
				return state_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("digital signature verification failed for transaction number %v %s %s", i+1, transaction.From, err)
			}
		}
		if !CheckNonce(user_nonce_tracker[transaction.From], uint64(transaction.Nonce)) && transaction.Type != "nft_deposit" && transaction.Type != "nft_mint" {
			return state_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("nonce check failed for transaction number %v", i+1)
		}
		user_nonce_tracker[transaction.From] = uint64(transaction.Nonce)
		switch transaction.Type {
		case "nft_deposit":
			users_updated_map[transaction.To] = true
			has_deposit = true
			if _, ok := state_balances[transaction.To]; ok {
				state_balances[transaction.To][transaction.NftContractAddress+"-"+strconv.Itoa(int(transaction.NftTokenId))] = true
			} else {
				state_balances[transaction.To] = make(map[string]bool)
				state_balances[transaction.To][transaction.NftContractAddress+"-"+strconv.Itoa(int(transaction.NftTokenId))] = true
			}
		case "nft_mint":
			users_updated_map[transaction.To] = true
			if _, ok := state_balances[transaction.To]; ok {
				state_balances[transaction.To][transaction.NftContractAddress+"-"+strconv.Itoa(int(transaction.NftTokenId))] = true
			} else {
				state_balances[transaction.To] = make(map[string]bool)
				state_balances[transaction.To][transaction.NftContractAddress+"-"+strconv.Itoa(int(transaction.NftTokenId))] = true
			}
		case "nft_transfer":
			users_updated_map[transaction.From] = true
			users_updated_map[transaction.To] = true
			if _, ok := state_balances[transaction.To]; ok {
				state_balances[transaction.To][transaction.NftContractAddress+"-"+strconv.Itoa(int(transaction.NftTokenId))] = true
			} else {
				state_balances[transaction.To] = make(map[string]bool)
				state_balances[transaction.To][transaction.NftContractAddress+"-"+strconv.Itoa(int(transaction.NftTokenId))] = true
			}
			if _, ok := state_balances[transaction.From]; ok {
				delete(state_balances[transaction.From], transaction.NftContractAddress+"-"+strconv.Itoa(int(transaction.NftTokenId)))
			}
			if len(state_balances[transaction.From]) == 0 {
				delete(state_balances, transaction.From)
			}
		case "nft_withdrawal":
			users_updated_map[transaction.From] = true
			has_withdraw = true
			if _, ok := state_balances[transaction.From]; ok {
				delete(state_balances[transaction.From], transaction.NftContractAddress+"-"+strconv.Itoa(int(transaction.NftTokenId)))
			}
			if len(state_balances[transaction.From]) == 0 {
				delete(state_balances, transaction.From)
			}
		}
	}

	settlement_type = 0
	if has_deposit {
		settlement_type = 1
	}
	if has_withdraw {
		settlement_type = 2
	}
	if has_contract_withdrawal {
		settlement_type = 3
	}
	if has_deposit && has_withdraw {
		settlement_type = 4
	}
	if has_deposit && has_contract_withdrawal {
		settlement_type = 5
	}
	if has_withdraw && has_contract_withdrawal {
		settlement_type = 6
	}
	if has_deposit && has_withdraw && has_contract_withdrawal {
		settlement_type = 7
	}

	return state_balances, settlement_type, users_updated_map, user_nonce_tracker, nil
}
