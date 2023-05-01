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
 				transferTransaction := TransferTransaction {
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

func GetDeltaBalances(transactions []Transaction, currencies []string) (map[string]map[string]string, uint, map[string]bool, map[string]uint64, error) {
	users_updated_map := make(map[string]bool)
	settlement_type := uint(0)
	user_nonce_tracker := make(map[string]uint64)
	delta_balances := make(map[string]map[string]string)
	has_deposit := false
	has_withdraw := false
	has_contract_withdrawal := false
	for i, transaction := range transactions {
		if transaction.IsInvalid {
			if transaction.Type == "contract_withdrawal" {
				has_contract_withdrawal = true
			}
			continue
		}
		if transaction.Type != "deposit" && transaction.Type != "contract_withdrawal" {
			verified, err := VerifyData(transaction, currencies)
			if !verified || err != nil {
				return delta_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("digital signature verification failed for transaction number %v %s %s", i+1, transaction.From, err)
			}
		}
		if !CheckNonce(user_nonce_tracker[transaction.From], uint64(transaction.Nonce)) && transaction.Type != "deposit" && transaction.Type != "contract_withdrawal" {
			return delta_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("nonce check failed for transaction number %v", i+1)
		}
		user_nonce_tracker[transaction.From] = uint64(transaction.Nonce)
		switch transaction.Type {
		case "transfer":
			users_updated_map[transaction.From] = true
			users_updated_map[transaction.To] = true
			if _, ok := delta_balances[transaction.From]; ok {
				if _, ok := delta_balances[transaction.From][transaction.CurrencyOrNftContractAddress]; ok {
					amount, ok := new(big.Int).SetString(transaction.AmountOrNftTokenId, 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.From][transaction.CurrencyOrNftContractAddress], 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Sub(current_balance, amount)
					delta_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = new_amt.String()
				} else {
					delta_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = "-" + transaction.AmountOrNftTokenId
				}
			} else {
				delta_balances[transaction.From] = make(map[string]string)
				delta_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = "-" + transaction.AmountOrNftTokenId
			}
			if _, ok := delta_balances[transaction.To]; ok {
				if _, ok := delta_balances[transaction.To][transaction.CurrencyOrNftContractAddress]; ok {
					amount, ok := new(big.Int).SetString(transaction.AmountOrNftTokenId, 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.To][transaction.CurrencyOrNftContractAddress], 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Add(amount, current_balance)
					delta_balances[transaction.To][transaction.CurrencyOrNftContractAddress] = new_amt.String()
				} else {
					delta_balances[transaction.To][transaction.CurrencyOrNftContractAddress] = transaction.AmountOrNftTokenId
				}
			} else {
				delta_balances[transaction.To] = make(map[string]string)
				delta_balances[transaction.To][transaction.CurrencyOrNftContractAddress] = transaction.AmountOrNftTokenId
			}
		case "deposit":
			users_updated_map[transaction.To] = true
			has_deposit = true
			if _, ok := delta_balances[transaction.To]; ok {
				if _, ok := delta_balances[transaction.To][transaction.CurrencyOrNftContractAddress]; ok {
					amount, ok := new(big.Int).SetString(transaction.AmountOrNftTokenId, 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.To][transaction.CurrencyOrNftContractAddress], 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Add(amount, current_balance)
					delta_balances[transaction.To][transaction.CurrencyOrNftContractAddress] = new_amt.String()
				} else {
					delta_balances[transaction.To][transaction.CurrencyOrNftContractAddress] = transaction.AmountOrNftTokenId
				}
			} else {
				delta_balances[transaction.To] = make(map[string]string)
				delta_balances[transaction.To][transaction.CurrencyOrNftContractAddress] = transaction.AmountOrNftTokenId
			}
		case "withdrawal":
			users_updated_map[transaction.From] = true
			has_withdraw = true
			if _, ok := delta_balances[transaction.From]; ok {
				if _, ok := delta_balances[transaction.From][transaction.CurrencyOrNftContractAddress]; ok {
					amount, ok := new(big.Int).SetString(transaction.AmountOrNftTokenId, 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.From][transaction.CurrencyOrNftContractAddress], 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Sub(current_balance, amount)
					delta_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = new_amt.String()
				} else {
					delta_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = "-" + transaction.AmountOrNftTokenId
				}
			} else {
				delta_balances[transaction.From] = make(map[string]string)
				delta_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = "-" + transaction.AmountOrNftTokenId
			}
		case "contract_withdrawal":
			users_updated_map[transaction.From] = true
			has_contract_withdrawal = true
			if _, ok := delta_balances[transaction.From]; ok {
				if _, ok := delta_balances[transaction.From][transaction.CurrencyOrNftContractAddress]; ok {
					amount, ok := new(big.Int).SetString(transaction.AmountOrNftTokenId, 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.From][transaction.CurrencyOrNftContractAddress], 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Sub(current_balance, amount)
					delta_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = new_amt.String()
				} else {
					delta_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = "-" + transaction.AmountOrNftTokenId
				}
			} else {
				delta_balances[transaction.From] = make(map[string]string)
				delta_balances[transaction.From][transaction.CurrencyOrNftContractAddress] = "-" + transaction.AmountOrNftTokenId
			}
		}
	}
	// process ID 0: only L2 transactions (+4)
	// process ID 1: deposit (+2)
	// process ID 2: withdrawal (backend) (+1)
	// process ID 3: withdrawal (contract) (+2)
	// process ID 4: deposit + withdrawal (backend)
	// process ID 5: deposit + withdrawal (contract)
	// process ID 6: withdrawal (backend) + withdrawal (contract)
	// process ID 7: deposit + withdrawal (backend) + withdrawal (contract)
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

	return delta_balances, settlement_type, users_updated_map, user_nonce_tracker, nil
}

// part of original main
func OldTransitionState(state_balances map[string]map[string]string, transactions []Transaction, currencies []string) (map[string]map[string]string, uint, map[string]bool, map[string]uint64, error) {
	delta_balances, settlement_type, users_updated, user_nonce_tracker, err := GetDeltaBalances(transactions, currencies)
	if err != nil {
		return state_balances, settlement_type, users_updated, user_nonce_tracker, err
	}
	for user, balances := range delta_balances {
		if _, ok := state_balances[user]; ok {
			for currency, balance := range balances {
				if _, ok := state_balances[user][currency]; ok {
					amount, ok := new(big.Int).SetString(balance, 10)
					if !ok {
						return state_balances, settlement_type, users_updated, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(state_balances[user][currency], 10)

					if !ok {
						return state_balances, settlement_type, users_updated, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Add(amount, current_balance)
					state_balances[user][currency] = new_amt.String()
				} else {
					state_balances[user][currency] = balance
				}
			}
		} else {
			state_balances[user] = make(map[string]string)
			for currency, balance := range balances {
				state_balances[user][currency] = balance
			}
		}
	}
	for user, balances := range state_balances {
		for currency, balance := range balances {
			amount, ok := new(big.Int).SetString(balance, 10)
			if !ok {
				return state_balances, settlement_type, users_updated, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
			}
			if amount.Cmp(big.NewInt(0)) < 0 {
				return state_balances, settlement_type, users_updated, user_nonce_tracker, fmt.Errorf("balance for user %s and currency %s is less than 0, %s", user, currency, amount)
			}
		}
	}

	return state_balances, settlement_type, users_updated, user_nonce_tracker, nil
}

// part of yesterdays commit
func NFTTransitionState(state_balances map[string]map[string]string, transactions []Transaction, currencies []string) (map[string]map[string]string, uint, map[string]bool, map[string]uint64, error) {
	users_updated_map := make(map[string]bool)
	settlement_type := uint(0)
	user_nonce_tracker := make(map[string]uint64)
	has_deposit := false
	has_withdraw := false
	has_contract_withdrawal := false
	for i, transaction := range transactions {
		if transaction.Type != "nft_deposit" && transaction.Type != "nft_mint" {
			verified, err := VerifyData(transaction, currencies)
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
				state_balances[transaction.To][transaction.AmountOrNftTokenId+"-"+transaction.AmountOrNftTokenId] = "yes"
			} else {
				state_balances[transaction.To] = make(map[string]string)
				state_balances[transaction.To][transaction.AmountOrNftTokenId+"-"+transaction.AmountOrNftTokenId] = "yes"
			}
		case "nft_mint":
			users_updated_map[transaction.To] = true
			if _, ok := state_balances[transaction.To]; ok {
				state_balances[transaction.To][transaction.AmountOrNftTokenId+"-"+transaction.AmountOrNftTokenId] = "yes"
			} else {
				state_balances[transaction.To] = make(map[string]string)
				state_balances[transaction.To][transaction.AmountOrNftTokenId+"-"+transaction.AmountOrNftTokenId] = "yes"
			}
		case "nft_transfer":
			users_updated_map[transaction.From] = true
			users_updated_map[transaction.To] = true
			if _, ok := state_balances[transaction.To]; ok {
				state_balances[transaction.To][transaction.AmountOrNftTokenId+"-"+transaction.AmountOrNftTokenId] = "yes"
			} else {
				state_balances[transaction.To] = make(map[string]string)
				state_balances[transaction.To][transaction.AmountOrNftTokenId+"-"+transaction.AmountOrNftTokenId] = "yes"
			}
			if _, ok := state_balances[transaction.From]; ok {
				delete(state_balances[transaction.From], transaction.AmountOrNftTokenId+"-"+transaction.AmountOrNftTokenId)
			}
			if len(state_balances[transaction.From]) == 0 {
				delete(state_balances, transaction.From)
			}
		case "nft_withdrawal":
			users_updated_map[transaction.From] = true
			has_withdraw = true
			if _, ok := state_balances[transaction.From]; ok {
				delete(state_balances[transaction.From], transaction.AmountOrNftTokenId+"-"+transaction.AmountOrNftTokenId)
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

func UpdateStateBalances(state_balances map[string]map[string]string, user string, delta_balance map[string]string) (map[string]map[string]string, error) {
	if _, ok := state_balances[user]; ok {
		for currency, balance := range delta_balance {
			if _, ok := state_balances[user][currency]; ok {
				amount, ok := new(big.Int).SetString(balance, 10)
				if !ok {
					return state_balances, fmt.Errorf("error converting amount to big int")
				}
				current_balance, ok := new(big.Int).SetString(state_balances[user][currency], 10)

				if !ok {
					return state_balances, fmt.Errorf("error converting amount to big int")
				}
				new_amt := new(big.Int)
				new_amt.Add(amount, current_balance)
				state_balances[user][currency] = new_amt.String()
			} else {
				state_balances[user][currency] = balance
			}
		}
	} else {
		state_balances[user] = make(map[string]string)
		for currency, balance := range delta_balance {
			state_balances[user][currency] = balance
		}
	}

	for currency, balance := range state_balances[user] {
		amount, ok := new(big.Int).SetString(balance, 10)
		if !ok {
			return state_balances, fmt.Errorf("error converting amount to big int")
		}
		if amount.Cmp(big.NewInt(0)) < 0 {
			return state_balances, fmt.Errorf("balance for user %s and currency %s is less than 0, %s", user, currency, amount)
		}
	}

	return state_balances, nil
}


func HandleNFTTransaction(state_balances map[string]map[string]string, transaction Transaction) (map[string]map[string]string, error) {
	switch transaction.Type {
	case "nft_deposit", "nft_mint":
		if _, ok := state_balances[transaction.To]; ok {
			state_balances[transaction.To][transaction.CurrencyOrNftContractAddress+"-"+transaction.AmountOrNftTokenId] = "yes"
		} else {
			state_balances[transaction.To] = make(map[string]string)
			state_balances[transaction.To][transaction.CurrencyOrNftContractAddress+"-"+transaction.AmountOrNftTokenId] = "yes"
		}
	case "nft_transfer":
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
		if _, ok := state_balances[transaction.From]; ok {
			delete(state_balances[transaction.From], transaction.CurrencyOrNftContractAddress+"-"+transaction.AmountOrNftTokenId)
		}
		if len(state_balances[transaction.From]) == 0 {
			delete(state_balances, transaction.From)
		}
	default:
		return state_balances, fmt.Errorf("unknown NFT transaction type: %s", transaction.Type)
	}

	return state_balances, nil
}

// merged method MAIN - NFT
func TransitionState(state_balances map[string]map[string]string, transactions []Transaction, currencies []string) (map[string]map[string]string, uint, map[string]bool, map[string]uint64, error) {
	delta_balances, settlement_type, users_updated, user_nonce_tracker, err := GetDeltaBalances(transactions, currencies)
	if err != nil {
		return state_balances, settlement_type, users_updated, user_nonce_tracker, err
	}

	for _, transaction := range transactions {
		if transaction.Type != "nft_deposit" && transaction.Type != "nft_mint" {
			verified, err := VerifyData(transaction, currencies)
			if !verified || err != nil {
				return state_balances, settlement_type, users_updated, user_nonce_tracker, fmt.Errorf("digital signature verification failed for transaction %s %s", transaction.From, err)
			}
		}
		if !CheckNonce(user_nonce_tracker[transaction.From], uint64(transaction.Nonce)) && transaction.Type != "nft_deposit" && transaction.Type != "nft_mint" {
			return state_balances, settlement_type, users_updated, user_nonce_tracker, fmt.Errorf("nonce check failed for transaction %s", transaction.From)
		}
		user_nonce_tracker[transaction.From] = uint64(transaction.Nonce)

		switch transaction.Type {
		case "nft_deposit", "nft_mint", "nft_transfer":
			state_balances, err = HandleNFTTransaction(state_balances, transaction)
			if err != nil {
				return state_balances, settlement_type, users_updated, user_nonce_tracker, err
			}
			users_updated[transaction.From] = true
			users_updated[transaction.To] = true
		default:
			users_updated[transaction.From] = true
			delta_balance := delta_balances[transaction.From]
			state_balances, err = UpdateStateBalances(state_balances, transaction.From, delta_balance)
			if err != nil {
				return state_balances, settlement_type, users_updated, user_nonce_tracker, err
			}
		}
	}

	return state_balances, settlement_type, users_updated, user_nonce_tracker, nil
}

