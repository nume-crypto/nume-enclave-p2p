package main

import (
	"fmt"
	"math/big"
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

func GenerateTransactionPublicData(transactions []Transaction, address_pubkey_map map[string]string, block_number float64) TransactionPublicData {
	var transferTransactions []TransferTransaction
	for _, transaction := range transactions {
		switch transaction.Type {
		case "transfer":
			if address_pubkey_map[transaction.From] != "" && address_pubkey_map[transaction.To] != "" {
				encryptedTxSender, err := EncryptTransactionWithPubKey(&transaction, block_number, address_pubkey_map[transaction.From])
				if err != nil {
					fmt.Println("unable to encrypt message for ", transaction.From)
					continue
				}
				encryptedTxReceiver, err := EncryptTransactionWithPubKey(&transaction, block_number, address_pubkey_map[transaction.To])
				if err != nil {
					fmt.Println("unable to encrypt message for ", transaction.To)
					continue
				}
				transferTransaction := TransferTransaction{
					ID:                transaction.Id,
					EncryptionOutputs: []string{encryptedTxSender, encryptedTxReceiver},
				}
				transferTransactions = append(transferTransactions, transferTransaction)
			}
		}
	}

	transactionPublicData := TransactionPublicData{
		Transactions: transferTransactions,
	}
	return transactionPublicData
}

func TransitionState(state_balances map[string]map[string]string, transactions []Transaction, currencies []string) (map[string]map[string]string, uint, map[string]bool, map[string]uint64, error) {
	users_updated_map := make(map[string]bool)
	settlement_type := uint(0)
	user_nonce_tracker := make(map[string]uint64)
	has_deposit := false
	has_withdraw := false
	has_contract_withdrawal := false
	has_nft_deposit := false
	has_nft_withdraw := false
	for i, transaction := range transactions {
		if transaction.IsInvalid {
			if transaction.Type == "contract_withdrawal" {
				has_contract_withdrawal = true
			}
			continue
		}
		if transaction.Type != "nft_deposit" && transaction.Type != "nft_mint" && transaction.Type != "deposit" && transaction.Type != "contract_withdrawal" {
			verified, err := VerifyData(transaction, currencies)
			if !verified || err != nil {
				return state_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("digital signature verification failed for transaction number %v %s %s", i+1, transaction.From, err)
			}
		}
		if !CheckNonce(user_nonce_tracker[transaction.From], uint64(transaction.Nonce)) && transaction.Type != "nft_deposit" && transaction.Type != "nft_mint" && transaction.Type != "deposit" && transaction.Type != "contract_withdrawal" {
			return state_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("nonce check failed for transaction number %v", i+1)
		}
		user_nonce_tracker[transaction.From] = uint64(transaction.Nonce)
		switch transaction.Type {
		case "nft_deposit":
			users_updated_map[transaction.To] = true
			has_deposit = true
			if _, ok := state_balances[transaction.To]; ok {
				state_balances[transaction.To][transaction.CurrencyOrNftContractAddress+"-"+transaction.AmountOrNftTokenId] = "yes"
			} else {
				state_balances[transaction.To] = make(map[string]string)
				state_balances[transaction.To][transaction.CurrencyOrNftContractAddress+"-"+transaction.AmountOrNftTokenId] = "yes"
			}
		case "nft_mint":
			users_updated_map[transaction.To] = true
			if _, ok := state_balances[transaction.To]; ok {
				state_balances[transaction.To][transaction.CurrencyOrNftContractAddress+"-"+transaction.AmountOrNftTokenId] = "yes"
			} else {
				state_balances[transaction.To] = make(map[string]string)
				state_balances[transaction.To][transaction.CurrencyOrNftContractAddress+"-"+transaction.AmountOrNftTokenId] = "yes"
			}
		case "nft_transfer":
			users_updated_map[transaction.From] = true
			users_updated_map[transaction.To] = true
			if _, ok := state_balances[transaction.To]; ok {
				state_balances[transaction.To][transaction.CurrencyOrNftContractAddress+"-"+transaction.AmountOrNftTokenId] = "yes"
			} else {
				state_balances[transaction.To] = make(map[string]string)
				state_balances[transaction.To][transaction.CurrencyOrNftContractAddress+"-"+transaction.AmountOrNftTokenId] = "yes"
			}
			if _, ok := state_balances[transaction.From]; ok {
				delete(state_balances[transaction.From], transaction.CurrencyOrNftContractAddress+"-"+transaction.AmountOrNftTokenId)
			}
			if len(state_balances[transaction.From]) == 0 {
				delete(state_balances, transaction.From)
			}
		case "nft_withdrawal":
			users_updated_map[transaction.From] = true
			has_withdraw = true
			if _, ok := state_balances[transaction.From]; ok {
				delete(state_balances[transaction.From], transaction.CurrencyOrNftContractAddress+"-"+transaction.AmountOrNftTokenId)
			}
			if len(state_balances[transaction.From]) == 0 {
				delete(state_balances, transaction.From)
			}
		}
		switch transaction.Type {
		case "transfer":
			users_updated_map[transaction.From] = true
			users_updated_map[transaction.To] = true
			if _, ok := state_balances[transaction.From]; ok {
				if _, ok := state_balances[transaction.From][transaction.CurrencyOrNftContractAddress]; ok {
					amount, ok := new(big.Int).SetString(transaction.AmountOrNftTokenId, 10)
					if !ok {
						return state_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(state_balances[transaction.From][transaction.CurrencyOrNftContractAddress], 10)
					if !ok {
						return state_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Sub(current_balance, amount)
					state_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = new_amt.String()
				} else {
					state_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = "-" + transaction.AmountOrNftTokenId
				}
			} else {
				state_balances[transaction.From] = make(map[string]string)
				state_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = "-" + transaction.AmountOrNftTokenId
			}
			if _, ok := state_balances[transaction.To]; ok {
				if _, ok := state_balances[transaction.To][transaction.CurrencyOrNftContractAddress]; ok {
					amount, ok := new(big.Int).SetString(transaction.AmountOrNftTokenId, 10)
					if !ok {
						return state_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(state_balances[transaction.To][transaction.CurrencyOrNftContractAddress], 10)
					if !ok {
						return state_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Add(amount, current_balance)
					state_balances[transaction.To][transaction.CurrencyOrNftContractAddress] = new_amt.String()
				} else {
					state_balances[transaction.To][transaction.CurrencyOrNftContractAddress] = transaction.AmountOrNftTokenId
				}
			} else {
				state_balances[transaction.To] = make(map[string]string)
				state_balances[transaction.To][transaction.CurrencyOrNftContractAddress] = transaction.AmountOrNftTokenId
			}
		case "deposit":
			users_updated_map[transaction.To] = true
			has_deposit = true
			if _, ok := state_balances[transaction.To]; ok {
				if _, ok := state_balances[transaction.To][transaction.CurrencyOrNftContractAddress]; ok {
					amount, ok := new(big.Int).SetString(transaction.AmountOrNftTokenId, 10)
					if !ok {
						return state_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(state_balances[transaction.To][transaction.CurrencyOrNftContractAddress], 10)
					if !ok {
						return state_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Add(amount, current_balance)
					state_balances[transaction.To][transaction.CurrencyOrNftContractAddress] = new_amt.String()
				} else {
					state_balances[transaction.To][transaction.CurrencyOrNftContractAddress] = transaction.AmountOrNftTokenId
				}
			} else {
				state_balances[transaction.To] = make(map[string]string)
				state_balances[transaction.To][transaction.CurrencyOrNftContractAddress] = transaction.AmountOrNftTokenId
			}
		case "withdrawal":
			users_updated_map[transaction.From] = true
			has_withdraw = true
			if _, ok := state_balances[transaction.From]; ok {
				if _, ok := state_balances[transaction.From][transaction.CurrencyOrNftContractAddress]; ok {
					amount, ok := new(big.Int).SetString(transaction.AmountOrNftTokenId, 10)
					if !ok {
						return state_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(state_balances[transaction.From][transaction.CurrencyOrNftContractAddress], 10)
					if !ok {
						return state_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Sub(current_balance, amount)
					state_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = new_amt.String()
				} else {
					state_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = "-" + transaction.AmountOrNftTokenId
				}
			} else {
				state_balances[transaction.From] = make(map[string]string)
				state_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = "-" + transaction.AmountOrNftTokenId
			}
		case "contract_withdrawal":
			users_updated_map[transaction.From] = true
			has_contract_withdrawal = true
			if _, ok := state_balances[transaction.From]; ok {
				if _, ok := state_balances[transaction.From][transaction.CurrencyOrNftContractAddress]; ok {
					amount, ok := new(big.Int).SetString(transaction.AmountOrNftTokenId, 10)
					if !ok {
						return state_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(state_balances[transaction.From][transaction.CurrencyOrNftContractAddress], 10)
					if !ok {
						return state_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Sub(current_balance, amount)
					state_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = new_amt.String()
				} else {
					state_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = "-" + transaction.AmountOrNftTokenId
				}
			} else {
				state_balances[transaction.From] = make(map[string]string)
				state_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = "-" + transaction.AmountOrNftTokenId
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
	if has_nft_deposit {
		settlement_type = 8
	}
	if has_nft_withdraw {
		settlement_type = 9
	}

	return state_balances, settlement_type, users_updated_map, user_nonce_tracker, nil
}
