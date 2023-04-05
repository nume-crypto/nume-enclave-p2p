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

func GetDeltaBalances(transactions []Transaction, user_keys map[string]UserKeys) (map[string]map[uint]string, string, map[string]bool, error) {
	users_updated_map := make(map[string]bool)
	settlement_type := "notarizeSettlementSignedByAllUsers"
	user_nonce_tracker := make(map[string]uint64)
	delta_balances := make(map[string]map[uint]string)
	has_transfer := false
	has_deposit := false
	has_withdraw := false
	for i, transaction := range transactions {
		if transaction.Type != "deposit" {
			ds_message, ok := DigitalSignatureMessageHash(transaction.From, transaction.To, strconv.FormatUint(uint64(transaction.CurrencyTokenOrder), 10), transaction.Amount, strconv.FormatUint(uint64(transaction.Nonce), 10))
			if !ok {
				return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error generating digital signature message hash")
			}
			if !VerifyDigitalSignature(hex.EncodeToString(ds_message), transaction.Signature, user_keys[transaction.From].BlsG2PublicKey) {
				return delta_balances, settlement_type, users_updated_map, fmt.Errorf("digital signature verification failed for transaction number %v %s %s", i+1, hex.EncodeToString(ds_message), user_keys[transaction.From].HashedPublicKey)
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
			has_transfer = true
			if _, ok := delta_balances[transaction.From]; ok {
				if _, ok := delta_balances[transaction.From][transaction.CurrencyTokenOrder]; ok {
					amount, ok := new(big.Int).SetString(transaction.Amount, 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.From][transaction.CurrencyTokenOrder], 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
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
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.To][transaction.CurrencyTokenOrder], 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
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
			users_updated_map[transaction.To] = true
			has_deposit = true
			if _, ok := delta_balances[transaction.To]; ok {
				if _, ok := delta_balances[transaction.To][transaction.CurrencyTokenOrder]; ok {
					amount, ok := new(big.Int).SetString(transaction.Amount, 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.To][transaction.CurrencyTokenOrder], 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
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
			users_updated_map[transaction.From] = true
			has_withdraw = true
			if _, ok := delta_balances[transaction.From]; ok {
				if _, ok := delta_balances[transaction.From][transaction.CurrencyTokenOrder]; ok {
					amount, ok := new(big.Int).SetString(transaction.Amount, 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(delta_balances[transaction.From][transaction.CurrencyTokenOrder], 10)
					if !ok {
						return delta_balances, settlement_type, users_updated_map, fmt.Errorf("error converting amount to big int")
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
	if has_transfer && has_deposit && has_withdraw {
		settlement_type = "notarizeSettlementWithDepositsAndWithdrawals"
	} else if has_transfer && has_deposit {
		settlement_type = "notarizeSettlementWithDeposits"
	} else if has_transfer && has_withdraw {
		settlement_type = "notarizeSettlementWithWithdrawals"
	}
	return delta_balances, settlement_type, users_updated_map, nil
}

func TransitionState(state_balances map[string]map[uint]string, transactions []Transaction, user_keys map[string]UserKeys) (map[string]map[uint]string, string, map[string]bool, error) {
	delta_balances, settlement_type, users_updated, err := GetDeltaBalances(transactions, user_keys)
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
			state_balances[user] = make(map[uint]string)
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
				return state_balances, settlement_type, users_updated, fmt.Errorf("balance for user %s and currency %d is less than 0", user, currency)
			}
		}
	}

	return state_balances, settlement_type, users_updated, nil
}
