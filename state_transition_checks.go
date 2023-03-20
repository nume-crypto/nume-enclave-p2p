package main

import (
	"fmt"
	"math/big"
	"strconv"
)

type Transaction struct {
	From       string
	To         string
	Amount     string
	Nonce      uint64
	CurrencyId uint
	Type       string
}

func CheckNonce(last_nonce, current_nonce uint64) bool {
	return last_nonce < current_nonce
}

func GetDeltaBalances(transactions []Transaction) (map[string]map[uint]string, error) {
	user_nonce_tracker := make(map[string]uint64)
	delta_balances := make(map[string]map[uint]string)
	for i, transaction := range transactions {
		if !CheckNonce(user_nonce_tracker[transaction.From], transaction.Nonce) && transaction.Type != "deposit" {
			return delta_balances, fmt.Errorf("nonce check failed for transaction number %v", i+1)
		}
		user_nonce_tracker[transaction.From] = transaction.Nonce
		switch transaction.Type {
		case "transfer":
			if _, ok := delta_balances[transaction.From]; ok {
				if _, ok := delta_balances[transaction.From][transaction.CurrencyId]; ok {
					amount, ok := new(big.Int).SetString(transaction.Amount, 10)
					if !ok {
						return delta_balances, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.From][transaction.CurrencyId], 10)
					if !ok {
						return delta_balances, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Sub(current_balance, amount)
					delta_balances[transaction.From][transaction.CurrencyId] = new_amt.String()
				} else {
					delta_balances[transaction.From][transaction.CurrencyId] = "-" + transaction.Amount
				}
			} else {
				delta_balances[transaction.From] = make(map[uint]string)
				delta_balances[transaction.From][transaction.CurrencyId] = "-" + transaction.Amount
			}
			if _, ok := delta_balances[transaction.To]; ok {
				if _, ok := delta_balances[transaction.To][transaction.CurrencyId]; ok {
					amount, ok := new(big.Int).SetString(transaction.Amount, 10)
					if !ok {
						return delta_balances, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.To][transaction.CurrencyId], 10)
					if !ok {
						return delta_balances, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Add(amount, current_balance)
					delta_balances[transaction.To][transaction.CurrencyId] = new_amt.String()
				} else {
					delta_balances[transaction.To][transaction.CurrencyId] = transaction.Amount
				}
			} else {
				delta_balances[transaction.To] = make(map[uint]string)
				delta_balances[transaction.To][transaction.CurrencyId] = transaction.Amount
			}
		case "deposit":
			if _, ok := delta_balances[transaction.To]; ok {
				if _, ok := delta_balances[transaction.To][transaction.CurrencyId]; ok {
					amount, ok := new(big.Int).SetString(transaction.Amount, 10)
					if !ok {
						return delta_balances, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.To][transaction.CurrencyId], 10)
					if !ok {
						return delta_balances, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Add(amount, current_balance)
					delta_balances[transaction.To][transaction.CurrencyId] = new_amt.String()
				} else {
					delta_balances[transaction.To][transaction.CurrencyId] = transaction.Amount
				}
			} else {
				delta_balances[transaction.To] = make(map[uint]string)
				delta_balances[transaction.To][transaction.CurrencyId] = transaction.Amount
			}
		case "withdraw":
			if _, ok := delta_balances[transaction.From]; ok {
				if _, ok := delta_balances[transaction.From][transaction.CurrencyId]; ok {
					amount, ok := new(big.Int).SetString(transaction.Amount, 10)
					if !ok {
						return delta_balances, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.From][transaction.CurrencyId], 10)
					if !ok {
						return delta_balances, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Sub(current_balance, amount)
					delta_balances[transaction.From][transaction.CurrencyId] = new_amt.String()
				} else {
					delta_balances[transaction.From][transaction.CurrencyId] = "-" + transaction.Amount
				}
			} else {
				delta_balances[transaction.From] = make(map[uint]string)
				delta_balances[transaction.From][transaction.CurrencyId] = "-" + transaction.Amount
			}
		}
	}
	return delta_balances, nil
}

func TransitionState(state_balances map[string]map[string]string, transactions []Transaction) (map[string]map[string]string, error) {
	delta_balances, err := GetDeltaBalances(transactions)
	if err != nil {
		return state_balances, err
	}
	for user, balances := range delta_balances {
		if _, ok := state_balances[user]; ok {
			for currency, balance := range balances {
				c := strconv.Itoa(int(currency))
				if _, ok := state_balances[user][c]; ok {
					amount, ok := new(big.Int).SetString(balance, 10)
					if !ok {
						return state_balances, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(state_balances[user][c], 10)

					if !ok {
						return state_balances, fmt.Errorf("error converting amount to big int")
					}
					new_amt := new(big.Int)
					new_amt.Add(amount, current_balance)
					state_balances[user][c] = new_amt.String()
				} else {
					state_balances[user][c] = balance
				}
			}
		} else {
			state_balances[user] = make(map[string]string)
			for currency, balance := range balances {
				c := strconv.Itoa(int(currency))
				state_balances[user][c] = balance
			}
		}
	}

	return state_balances, nil
}
