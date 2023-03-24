package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
)

func CheckNonce(last_nonce, current_nonce uint64) bool {
	return last_nonce < current_nonce
}

func GetDeltaBalances(transactions []Transaction, user_keys map[string]UserKeys) (map[string]map[uint]string, error) {
	user_nonce_tracker := make(map[string]uint64)
	delta_balances := make(map[string]map[uint]string)
	for i, transaction := range transactions {

		if transaction.Type != "deposit" {
			ds_message, ok := DigitalSignatureMessageHash(transaction.From, transaction.To, strconv.FormatUint(uint64(transaction.CurrencyTokenOrder), 10), transaction.Amount, strconv.FormatUint(uint64(transaction.Nonce), 10))
			if !ok {
				fmt.Println(transaction.From, transaction.To, transaction.Amount, strconv.FormatUint(uint64(transaction.Nonce), 10), strconv.FormatUint(uint64(transaction.CurrencyTokenOrder), 10))
				return delta_balances, fmt.Errorf("error generating digital signature message hash")
			}
			if !VerifyDigitalSignature(hex.EncodeToString(ds_message), transaction.Signature, user_keys[transaction.From].BlsG2PublicKey) {
				return delta_balances, fmt.Errorf("digital signature verification failed for transaction number %v", i+1)
			}
		}
		if !CheckNonce(user_nonce_tracker[transaction.From], uint64(transaction.Nonce)) && transaction.Type != "deposit" {
			return delta_balances, fmt.Errorf("nonce check failed for transaction number %v", i+1)
		}
		user_nonce_tracker[transaction.From] = uint64(transaction.Nonce)
		switch transaction.Type {
		case "transfer":
			if _, ok := delta_balances[transaction.From]; ok {
				if _, ok := delta_balances[transaction.From][transaction.CurrencyTokenOrder]; ok {
					amount, ok := new(big.Int).SetString(transaction.Amount, 10)
					if !ok {
						return delta_balances, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.From][transaction.CurrencyTokenOrder], 10)
					if !ok {
						return delta_balances, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Sub(current_balance, amount)
					delta_balances[transaction.From][transaction.CurrencyTokenOrder] = new_amt.String()
				} else {
					delta_balances[transaction.From][transaction.CurrencyTokenOrder] = "-" + transaction.Amount
				}
			} else {
				delta_balances[transaction.From] = make(map[uint]string)
				delta_balances[transaction.From][transaction.CurrencyTokenOrder] = "-" + transaction.Amount
			}
			if _, ok := delta_balances[transaction.To]; ok {
				if _, ok := delta_balances[transaction.To][transaction.CurrencyTokenOrder]; ok {
					amount, ok := new(big.Int).SetString(transaction.Amount, 10)
					if !ok {
						return delta_balances, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.To][transaction.CurrencyTokenOrder], 10)
					if !ok {
						return delta_balances, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Add(amount, current_balance)
					delta_balances[transaction.To][transaction.CurrencyTokenOrder] = new_amt.String()
				} else {
					delta_balances[transaction.To][transaction.CurrencyTokenOrder] = transaction.Amount
				}
			} else {
				delta_balances[transaction.To] = make(map[uint]string)
				delta_balances[transaction.To][transaction.CurrencyTokenOrder] = transaction.Amount
			}
		case "deposit":
			if _, ok := delta_balances[transaction.To]; ok {
				if _, ok := delta_balances[transaction.To][transaction.CurrencyTokenOrder]; ok {
					amount, ok := new(big.Int).SetString(transaction.Amount, 10)
					if !ok {
						return delta_balances, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.To][transaction.CurrencyTokenOrder], 10)
					if !ok {
						return delta_balances, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Add(amount, current_balance)
					delta_balances[transaction.To][transaction.CurrencyTokenOrder] = new_amt.String()
				} else {
					delta_balances[transaction.To][transaction.CurrencyTokenOrder] = transaction.Amount
				}
			} else {
				delta_balances[transaction.To] = make(map[uint]string)
				delta_balances[transaction.To][transaction.CurrencyTokenOrder] = transaction.Amount
			}
		case "withdrawal":
			if _, ok := delta_balances[transaction.From]; ok {
				if _, ok := delta_balances[transaction.From][transaction.CurrencyTokenOrder]; ok {
					amount, ok := new(big.Int).SetString(transaction.Amount, 10)
					if !ok {
						return delta_balances, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.From][transaction.CurrencyTokenOrder], 10)
					if !ok {
						return delta_balances, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Sub(current_balance, amount)
					delta_balances[transaction.From][transaction.CurrencyTokenOrder] = new_amt.String()
				} else {
					delta_balances[transaction.From][transaction.CurrencyTokenOrder] = "-" + transaction.Amount
				}
			} else {
				delta_balances[transaction.From] = make(map[uint]string)
				delta_balances[transaction.From][transaction.CurrencyTokenOrder] = "-" + transaction.Amount
			}
		}
	}
	return delta_balances, nil
}

func TransitionState(state_balances map[string]map[uint]string, transactions []Transaction, user_keys map[string]UserKeys) (map[string]map[uint]string, error) {
	delta_balances, err := GetDeltaBalances(transactions, user_keys)
	if err != nil {
		return state_balances, err
	}
	for user, balances := range delta_balances {
		if _, ok := state_balances[user]; ok {
			for currency, balance := range balances {
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
			state_balances[user] = make(map[uint]string)
			for currency, balance := range balances {
				state_balances[user][currency] = balance
			}
		}
	}

	return state_balances, nil
}
