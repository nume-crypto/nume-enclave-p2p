package main

import (
	"fmt"
	"math/big"
	"strconv"
	"time"
)

func CheckNonce(last_nonce, current_nonce uint64) bool {
	return last_nonce < current_nonce
}

func TransitionState(state_balances map[string]map[string]string, transactions []interface{}, currencies []string, nft_collections []map[string]interface{}, used_lister_nonce map[string][]uint, meta_data map[string]interface{}) (map[string]map[string]string, HasProcess, map[string]bool, map[string]uint64, error) {
	users_updated_map := make(map[string]bool)
	user_nonce_tracker := make(map[string]uint64)
	nft_collections_map := make(map[string]map[string]interface{})
	defer TimeTrack(time.Now(), "TransitionState")
	for _, nft_collection := range nft_collections {
		nft_collections_map[nft_collection["ContractAddress"].(string)] = nft_collection
	}
	nume_address := meta_data["nume_user"].(string)
	users_updated_map[nume_address] = true
	fee_currency_token := meta_data["fee_currency_token"].(string)
	has_process := HasProcess{}
	for i, tx := range transactions {
		var transaction Transaction
		var trade Trade
		if t, ok := tx.(map[string]interface{}); ok {
			if t["Type"] == "nft_trade" {
				trade = Trade{
					Id:                 uint(t["Id"].(float64)),
					From:               t["From"].(string),
					To:                 t["To"].(string),
					ListAmount:         t["ListAmount"].(string),
					BuyAmount:          t["BuyAmount"].(string),
					Currency:           t["Currency"].(string),
					ListerNonce:        uint(t["ListerNonce"].(float64)),
					BuyerNonce:         uint(t["BuyerNonce"].(float64)),
					NftTokenId:         t["NftTokenId"].(string),
					NftContractAddress: t["NftContractAddress"].(string),
					Type:               t["Type"].(string),
					ListSignature:      t["ListSignature"].(string),
					BuySignature:       t["BuySignature"].(string),
					L2Minted:           t["L2Minted"].(bool),
					RoyaltyAmount:      t["RoyaltyAmount"].(string),
					NumeFees:           t["NumeFees"].(string),
				}
			} else {
				transaction = Transaction{
					Id:                           uint(t["Id"].(float64)),
					From:                         t["From"].(string),
					To:                           t["To"].(string),
					AmountOrNftTokenId:           t["AmountOrNftTokenId"].(string),
					Nonce:                        uint(t["Nonce"].(float64)),
					CurrencyOrNftContractAddress: t["CurrencyOrNftContractAddress"].(string),
					Type:                         t["Type"].(string),
					Signature:                    t["Signature"].(string),
					IsInvalid:                    t["IsInvalid"].(bool),
					L2Minted:                     t["L2Minted"].(bool),
					Data:                         t["Data"].(string),
					NumeFees:                     t["NumeFees"].(string),
					MintFees:                     t["MintFees"].(string),
					MintFeesToken:                t["MintFeesToken"].(string),
				}
			}
		} else {
			return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("invalid transaction type")
		}
		if transaction.IsInvalid {
			if transaction.Type == "contract_withdrawal" {
				has_process.HasContractWithdrawal = true
			}
			if transaction.Type == "nft_contract_withdrawal" {
				has_process.HasNFTContractWithdrawal = true
			}
			continue
		}
		if transaction.Type != "" && transaction.Type != "nft_deposit" && transaction.Type != "nft_mint" && transaction.Type != "deposit" && transaction.Type != "contract_withdrawal" && transaction.Type != "nft_contract_withdrawal" {
			verified, err := VerifyData(transaction, currencies)
			if !verified || err != nil {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("digital signature verification failed for transaction number %v %s %s", i+1, transaction.From, err)
			}
		}
		if !CheckNonce(user_nonce_tracker[transaction.From], uint64(transaction.Nonce)) && transaction.Type != "nft_deposit" && transaction.Type != "deposit" && transaction.Type != "contract_withdrawal" && transaction.Type != "nft_contract_withdrawal" && transaction.Type != "" {
			return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("nonce check failed for transaction number %v", i+1)
		}
		updateHasProcess(&has_process, transaction)
		if transaction.Type == "nft_mint" {
			user_nonce_tracker[transaction.To] = uint64(transaction.Nonce)
		} else {
			user_nonce_tracker[transaction.From] = uint64(transaction.Nonce)
		}

		// Handle Nume Fees wherever applicable
		var nume_fees *big.Int
		var ok bool
		var error_in_fee error
		if trade.Type == "nft_trade" {
			nume_fees, ok = new(big.Int).SetString(trade.NumeFees, 10)
			if !ok {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
			}
			state_balances, error_in_fee = DeductFees(state_balances, trade.To, fee_currency_token, nume_fees, nume_address)
			if error_in_fee != nil {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, error_in_fee
			}
		} else if transaction.Type == "transfer" || transaction.Type == "nft_transfer" || transaction.Type == "nft_mint" {
			nume_fees, ok = new(big.Int).SetString(transaction.NumeFees, 10)
			if !ok {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
			}
			sender := transaction.From
			if transaction.Type == "nft_mint" {
				mint_fee_amount_bi, ok := new(big.Int).SetString(transaction.MintFees, 10)
				if !ok {
					return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
				}
				state_balances, error_in_fee = DeductFees(state_balances, transaction.To, transaction.MintFeesToken, mint_fee_amount_bi, nft_collections_map[transaction.CurrencyOrNftContractAddress]["Owner"].(string))
				if error_in_fee != nil {
					return state_balances, has_process, users_updated_map, user_nonce_tracker, error_in_fee
				}
				sender = transaction.To
			}
			state_balances, error_in_fee = DeductFees(state_balances, sender, fee_currency_token, nume_fees, nume_address)
			if error_in_fee != nil {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, error_in_fee
			}
		}

		if transaction.Type == "nft_deposit" || transaction.Type == "nft_transfer" || transaction.Type == "nft_mint" || trade.Type == "nft_trade" {
			if transaction.Type == "nft_mint" {
				err := verifyMintData(transaction, nft_collections_map)
				if err != nil {
					return state_balances, has_process, users_updated_map, user_nonce_tracker, err
				}
			}
			tx_receiver := transaction.To
			tx_nft_contract := transaction.CurrencyOrNftContractAddress
			tx_nft_token_id := transaction.AmountOrNftTokenId
			if trade.Type == "nft_trade" {
				tx_receiver = trade.To
				tx_nft_contract = trade.NftContractAddress
				tx_nft_token_id = trade.NftTokenId
			}
			users_updated_map[tx_receiver] = true
			if _, ok := state_balances[tx_receiver]; ok {
				state_balances[tx_receiver][tx_nft_contract+"-"+tx_nft_token_id] = "yes"
			} else {
				state_balances[tx_receiver] = make(map[string]string)
				state_balances[tx_receiver][tx_nft_contract+"-"+tx_nft_token_id] = "yes"
			}
		}
		if transaction.Type == "nft_contract_withdrawal" || transaction.Type == "nft_withdrawal" || transaction.Type == "nft_transfer" || trade.Type == "nft_trade" {
			tx_sender := transaction.From
			tx_nft_contract := transaction.CurrencyOrNftContractAddress
			tx_nft_token_id := transaction.AmountOrNftTokenId
			if trade.Type == "nft_trade" {
				tx_sender = trade.From
				tx_nft_contract = trade.NftContractAddress
				tx_nft_token_id = trade.NftTokenId
			}
			users_updated_map[tx_sender] = true
			if _, ok := state_balances[tx_sender]; ok {
				delete(state_balances[tx_sender], tx_nft_contract+"-"+tx_nft_token_id)
			}
		}

		if transaction.Type == "deposit" || transaction.Type == "transfer" || trade.Type == "nft_trade" {
			tx_receiver := transaction.To
			tx_currency := transaction.CurrencyOrNftContractAddress
			tx_amt := transaction.AmountOrNftTokenId
			if trade.Type == "nft_trade" {
				tx_receiver = trade.From
				tx_currency = trade.Currency
				trade_buy_amt_bi, ok := new(big.Int).SetString(trade.BuyAmount, 10)
				if !ok {
					return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
				}
				trade_royalty_bi, ok := new(big.Int).SetString(trade.RoyaltyAmount, 10)
				if !ok {
					return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
				}
				tx_amt = new(big.Int).Sub(trade_buy_amt_bi, trade_royalty_bi).String()
			}
			if _, ok := state_balances[tx_receiver]; ok {
				if _, ok := state_balances[tx_receiver][tx_currency]; ok {
					amount, ok := new(big.Int).SetString(tx_amt, 10)
					if !ok {
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(state_balances[tx_receiver][tx_currency], 10)
					if !ok {
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting current_balance to big int")
					}
					new_amt := new(big.Int)
					new_amt.Add(amount, current_balance)
					state_balances[tx_receiver][tx_currency] = new_amt.String()
				} else {
					state_balances[tx_receiver][tx_currency] = tx_amt
				}
			} else {
				state_balances[tx_receiver] = make(map[string]string)
				state_balances[tx_receiver][tx_currency] = tx_amt
			}
		}
		if transaction.Type == "contract_withdrawal" || transaction.Type == "withdrawal" || transaction.Type == "transfer" || trade.Type == "nft_trade" {
			tx_sender := transaction.From
			tx_currency := transaction.CurrencyOrNftContractAddress
			tx_amt := transaction.AmountOrNftTokenId
			if trade.Type == "nft_trade" {
				tx_sender = trade.To
				tx_currency = trade.Currency
				trade_buy_amt_bi, ok := new(big.Int).SetString(trade.BuyAmount, 10)
				if !ok {
					return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
				}
				trade_royalty_bi, ok := new(big.Int).SetString(trade.RoyaltyAmount, 10)
				if !ok {
					return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
				}
				tx_amt = new(big.Int).Sub(trade_buy_amt_bi, trade_royalty_bi).String()
			}
			users_updated_map[tx_sender] = true
			if _, ok := state_balances[tx_sender]; ok {
				if _, ok := state_balances[tx_sender][tx_currency]; ok {
					amount, ok := new(big.Int).SetString(tx_amt, 10)
					if !ok {
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(state_balances[tx_sender][tx_currency], 10)
					if !ok {
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting current_balance to big int")
					}
					new_amt := new(big.Int)
					new_amt.Sub(current_balance, amount)
					state_balances[tx_sender][tx_currency] = new_amt.String()
					if new_amt.Cmp(big.NewInt(0)) == -1 {
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error: user balance negative")
					}
				} else {
					return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error: user does not have currency to transfer")
				}
			} else {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error: user does not have balance")
			}
		}

		if trade.Type == "nft_trade" {
			if !CheckNonce(user_nonce_tracker[trade.To], uint64(trade.BuyerNonce)) {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("nonce check failed for transaction number %v", i+1)
			}
			invalid_nonce := binarySearch(used_lister_nonce[trade.From], trade.ListerNonce)
			if invalid_nonce {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("invalid lister nonce for transaction number %v", i+1)
			}
			used_lister_nonce[trade.From] = append(used_lister_nonce[trade.From], trade.ListerNonce)
			// VERIFY LIST SIGNATURE AND BUY SIGNATURE
			list_message := NftTradeMessage(trade.From, trade.NftContractAddress, trade.NftTokenId, trade.Currency, trade.ListAmount, strconv.Itoa(int(trade.ListerNonce)))
			if !EthVerify(list_message, trade.ListSignature, trade.From) {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("invalid list signature")
			}
			buy_message := NftTradeMessage(trade.To, trade.NftContractAddress, trade.NftTokenId, trade.Currency, trade.BuyAmount, strconv.Itoa(int(trade.BuyerNonce)))
			if !EthVerify(buy_message, trade.BuySignature, trade.To) {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("invalid buy signature")
			}

			amount_bi, ok := new(big.Int).SetString(trade.BuyAmount, 10)
			if !ok {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
			}
			listed_amt_bi, ok := new(big.Int).SetString(trade.ListAmount, 10)
			if !ok {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
			}
			if amount_bi.Cmp(listed_amt_bi) < 0 {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("list amount must be less than or equal to buy amount")
			}

			// Handle ROYALTY fee
			royalty_amount_bi, ok := new(big.Int).SetString(trade.RoyaltyAmount, 10)
			if !ok {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
			}
			state_balances, error_in_fee = DeductFees(state_balances, trade.To, trade.Currency, royalty_amount_bi, nft_collections_map[trade.NftContractAddress]["Owner"].(string))
			if error_in_fee != nil {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, error_in_fee
			}
		}
	}

	return state_balances, has_process, users_updated_map, user_nonce_tracker, nil
}

func DeductFees(state_balances map[string]map[string]string, sender string, fee_currency_token string, fees *big.Int, receiver string) (map[string]map[string]string, error) {
	// Deduct from Sender
	if _, ok := state_balances[sender]; ok {
		if _, ok := state_balances[sender][fee_currency_token]; ok {
			current_balance, ok := new(big.Int).SetString(state_balances[sender][fee_currency_token], 10)
			if !ok {
				return state_balances, fmt.Errorf("error converting amount to big int")
			}
			new_amt := new(big.Int)
			new_amt.Sub(current_balance, fees)
			state_balances[sender][fee_currency_token] = new_amt.String()
			if new_amt.Cmp(big.NewInt(0)) == -1 {
				return state_balances, fmt.Errorf("user has negative balance to pay fees")
			}
		} else {
			return state_balances, fmt.Errorf("user does not have enough balance to pay fees " + sender + " " + fee_currency_token)
		}
	} else {
		return state_balances, fmt.Errorf("user does not have any balance to pay fees " + sender)
	}

	// Credit To Receiver
	if _, ok := state_balances[receiver]; ok {
		if _, ok := state_balances[receiver][fee_currency_token]; ok {
			current_balance, ok := new(big.Int).SetString(state_balances[receiver][fee_currency_token], 10)
			if !ok {
				return state_balances, fmt.Errorf("error converting amount to big int")
			}
			new_amt := new(big.Int)
			new_amt.Add(fees, current_balance)
			state_balances[receiver][fee_currency_token] = new_amt.String()
		} else {
			state_balances[receiver][fee_currency_token] = fees.String()
		}
	} else {
		state_balances[receiver] = make(map[string]string)
		state_balances[receiver][fee_currency_token] = fees.String()
	}
	return state_balances, nil
}
