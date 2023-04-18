package main

import (
	"fmt"
	"math/big"
)

func CheckNonce(last_nonce, current_nonce uint64) bool {
	return last_nonce < current_nonce
}

type TransferTransaction struct {
    ID               uint
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
				encryptedTxSender,err := EncryptTransactionECDSAPubKey(&transaction, block_number, address_pubkey_map[transaction.From])
				if err != nil {
					fmt.Println("unable to encrypt message for ", transaction.From)
					continue
				}
				encryptedTxReceiver,err := EncryptTransactionECDSAPubKey(&transaction, block_number, address_pubkey_map[transaction.To])
				if err != nil {
					fmt.Println("unable to encrypt message for ", transaction.To)
					continue
				}
				transferTransaction := TransferTransaction{
					ID: transaction.Id,
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

func GetDeltaBalances(transactions []Transaction, block_number int64) (map[string]map[string]string, uint, map[string]bool, error) {
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
			ds_message := DigitalSignatureMessage(transaction.From, transaction.To, transaction.Currency, transaction.Amount, uint64(transaction.Nonce), block_number)
			if !EthVerify(ds_message, transaction.Signature, transaction.From) {
				return delta_balances, settlement_type, users_updated_map, fmt.Errorf("digital signature verification failed for transaction number %v %s %s", i+1, ds_message, transaction.From)
			}
		}
		if !CheckNonce(user_nonce_tracker[transaction.From], uint64(transaction.Nonce)) && transaction.Type != "deposit" {
			return delta_balances, settlement_type, users_updated_map, fmt.Errorf("nonce check failed for transaction number %v", i+1)
		}
		user_nonce_tracker[transaction.From] = uint64(transaction.Nonce)
		switch transaction.Type {
		case "transfer":
			users_updated_map[transaction.From] = true
			users_updated_map[transaction.To] = true
			if _, ok := delta_balances[transaction.From]; ok {
				if _, ok := delta_balances[transaction.From][transaction.Currency]; ok {
					amount, ok := new(big.Int).SetString(transaction.Amount, 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.From][transaction.Currency], 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Sub(current_balance, amount)
					delta_balances[transaction.From][transaction.Currency] = new_amt.String()
				} else {
					delta_balances[transaction.From][transaction.Currency] = "-" + transaction.Amount
				}
			} else {
				delta_balances[transaction.From] = make(map[string]string)
				delta_balances[transaction.From][transaction.Currency] = "-" + transaction.Amount
			}
			if _, ok := delta_balances[transaction.To]; ok {
				if _, ok := delta_balances[transaction.To][transaction.Currency]; ok {
					amount, ok := new(big.Int).SetString(transaction.Amount, 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.To][transaction.Currency], 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Add(amount, current_balance)
					delta_balances[transaction.To][transaction.Currency] = new_amt.String()
				} else {
					delta_balances[transaction.To][transaction.Currency] = transaction.Amount
				}
			} else {
				delta_balances[transaction.To] = make(map[string]string)
				delta_balances[transaction.To][transaction.Currency] = transaction.Amount
			}
		case "deposit":
			users_updated_map[transaction.To] = true
			has_deposit = true
			if _, ok := delta_balances[transaction.To]; ok {
				if _, ok := delta_balances[transaction.To][transaction.Currency]; ok {
					amount, ok := new(big.Int).SetString(transaction.Amount, 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.To][transaction.Currency], 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Add(amount, current_balance)
					delta_balances[transaction.To][transaction.Currency] = new_amt.String()
				} else {
					delta_balances[transaction.To][transaction.Currency] = transaction.Amount
				}
			} else {
				delta_balances[transaction.To] = make(map[string]string)
				delta_balances[transaction.To][transaction.Currency] = transaction.Amount
			}
		case "withdrawal":
			users_updated_map[transaction.From] = true
			has_withdraw = true
			if _, ok := delta_balances[transaction.From]; ok {
				if _, ok := delta_balances[transaction.From][transaction.Currency]; ok {
					amount, ok := new(big.Int).SetString(transaction.Amount, 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.From][transaction.Currency], 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Sub(current_balance, amount)
					delta_balances[transaction.From][transaction.Currency] = new_amt.String()
				} else {
					delta_balances[transaction.From][transaction.Currency] = "-" + transaction.Amount
				}
			} else {
				delta_balances[transaction.From] = make(map[string]string)
				delta_balances[transaction.From][transaction.Currency] = "-" + transaction.Amount
			}
		case "contract_withdrawal":
			users_updated_map[transaction.From] = true
			has_contract_withdrawal = true
			if _, ok := delta_balances[transaction.From]; ok {
				if _, ok := delta_balances[transaction.From][transaction.Currency]; ok {
					amount, ok := new(big.Int).SetString(transaction.Amount, 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.From][transaction.Currency], 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Sub(current_balance, amount)
					delta_balances[transaction.From][transaction.Currency] = new_amt.String()
				} else {
					delta_balances[transaction.From][transaction.Currency] = "-" + transaction.Amount
				}
			} else {
				delta_balances[transaction.From] = make(map[string]string)
				delta_balances[transaction.From][transaction.Currency] = "-" + transaction.Amount
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

	return delta_balances, settlement_type, users_updated_map, nil
}

func TransitionState(state_balances map[string]map[string]string, transactions []Transaction, block_number int64) (map[string]map[string]string, uint, map[string]bool, error) {
	delta_balances, settlement_type, users_updated, err := GetDeltaBalances(transactions, block_number)
	if err != nil {
		return state_balances, settlement_type, users_updated, err
	}
	for user, balances := range delta_balances {
		if _, ok := state_balances[user]; ok {
			for currency, balance := range balances {
				if _, ok := state_balances[user][currency]; ok {
					amount, ok := new(big.Int).SetString(balance, 10)
					if !ok {
						return state_balances, settlement_type, users_updated, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(state_balances[user][currency], 10)

					if !ok {
						return state_balances, settlement_type, users_updated, fmt.Errorf("error converting amount to big int")
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
				return state_balances, settlement_type, users_updated, fmt.Errorf("error converting amount to big int")
			}
			if amount.Cmp(big.NewInt(0)) < 0 {
				return state_balances, settlement_type, users_updated, fmt.Errorf("balance for user %s and currency %s is less than 0, %s", user, currency, amount)
			}
		}
	}

	return state_balances, settlement_type, users_updated, nil
}
