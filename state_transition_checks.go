package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"

	solsha3 "github.com/miguelmota/go-solidity-sha3"
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

// func GenerateTransactionPublicData(transactions []Transaction, address_pubkey_map map[string]string, block_number float64) TransactionPublicData {
// 	var transferTransactions []TransferTransaction
// 	for _, transaction := range transactions {
// 		switch transaction.Type {
// 		case "transfer":
// 			if address_pubkey_map[transaction.From] != "" && address_pubkey_map[transaction.To] != "" {
// 				encryptedTxSender, err := EncryptTransactionWithPubKey(&transaction, block_number, address_pubkey_map[transaction.From])
// 				if err != nil {
// 					fmt.Println("unable to encrypt message for ", transaction.From)
// 					continue
// 				}
// 				encryptedTxReceiver, err := EncryptTransactionWithPubKey(&transaction, block_number, address_pubkey_map[transaction.To])
// 				if err != nil {
// 					fmt.Println("unable to encrypt message for ", transaction.To)
// 					continue
// 				}
// 				transferTransaction := TransferTransaction{
// 					ID:                transaction.Id,
// 					EncryptionOutputs: []string{encryptedTxSender, encryptedTxReceiver},
// 				}
// 				transferTransactions = append(transferTransactions, transferTransaction)
// 			}
// 		}
// 	}

// 	transactionPublicData := TransactionPublicData{
// 		Transactions: transferTransactions,
// 	}
// 	return transactionPublicData
// }

type HasProcess struct {
	HasDeposit               bool
	HasWithdrawal            bool
	HasContractWithdrawal    bool
	HasNFTDeposit            bool
	HasNFTContractWithdrawal bool
}

func binarySearch(array []uint, to_search uint) bool {
	found := false
	low := 0
	high := len(array) - 1
	for low <= high {
		mid := (low + high) / 2
		if array[mid] == to_search {
			found = true
			break
		}
		if array[mid] < to_search {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return found
}

func TransitionState(state_balances map[string]map[string]string, transactions []interface{}, currencies []string, nft_collections []map[string]interface{}, used_lister_nonce map[string][]uint) (map[string]map[string]string, HasProcess, map[string]bool, map[string]uint64, error) {
	users_updated_map := make(map[string]bool)
	user_nonce_tracker := make(map[string]uint64)
	nft_collections_map := make(map[string]map[string]interface{})
	defer TimeTrack(time.Now(), "TransitionState")
	for _, nft_collection := range nft_collections {
		nft_collections_map[nft_collection["ContractAddress"].(string)] = nft_collection
	}
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
		if !CheckNonce(user_nonce_tracker[transaction.From], uint64(transaction.Nonce)) && transaction.Type != "nft_deposit" && transaction.Type != "nft_mint" && transaction.Type != "deposit" && transaction.Type != "contract_withdrawal" && transaction.Type != "nft_contract_withdrawal" && transaction.Type != "" {
			return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("nonce check failed for transaction number %v", i+1)
		}
		user_nonce_tracker[transaction.From] = uint64(transaction.Nonce)
		switch transaction.Type {
		case "nft_deposit":
			users_updated_map[transaction.To] = true
			has_process.HasNFTDeposit = true
			if _, ok := state_balances[transaction.To]; ok {
				state_balances[transaction.To][transaction.CurrencyOrNftContractAddress+"-"+transaction.AmountOrNftTokenId] = "yes"
			} else {
				state_balances[transaction.To] = make(map[string]string)
				state_balances[transaction.To][transaction.CurrencyOrNftContractAddress+"-"+transaction.AmountOrNftTokenId] = "yes"
			}
		case "nft_mint":
			message := solsha3.SoliditySHA3(
				[]string{"address", "address"},
				[]interface{}{transaction.CurrencyOrNftContractAddress, transaction.To},
			)
			if !EthVerify(hex.EncodeToString(message), transaction.Signature, transaction.From) {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("invalid list signature")
			}
			if _, ok := nft_collections_map[transaction.CurrencyOrNftContractAddress]; !ok {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("nft collection not found")
			}
			mint_end_specifed, ok := new(big.Int).SetString(nft_collections_map[transaction.CurrencyOrNftContractAddress]["MintEnd"].(string), 10)
			if !ok {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("nft collection mint end is not valid")
			}
			mint_start_specifed, ok := new(big.Int).SetString(nft_collections_map[transaction.CurrencyOrNftContractAddress]["MintStart"].(string), 10)
			if !ok {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("nft collection mint end is not valid")
			}
			token_id, ok := new(big.Int).SetString(transaction.AmountOrNftTokenId, 10)
			if !ok {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("nft token id is not valid")
			}
			if mint_end_specifed.Cmp(big.NewInt(0)) == 1 && mint_end_specifed.Cmp(token_id) != 1 {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("nft collection token id should be less than mint end")
			}
			if mint_start_specifed.Cmp(token_id) != -1 {
				return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("nft collection token id should be greater than mint start")
			}
			for _, v := range nft_collections_map[transaction.CurrencyOrNftContractAddress]["MintUsers"].([]interface{}) {
				if v.(string) == transaction.From {
					return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("user does not have minting rights")
				}
			}
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
			has_process.HasWithdrawal = true
			if _, ok := state_balances[transaction.From]; ok {
				delete(state_balances[transaction.From], transaction.CurrencyOrNftContractAddress+"-"+transaction.AmountOrNftTokenId)
			}
			if len(state_balances[transaction.From]) == 0 {
				delete(state_balances, transaction.From)
			}
		case "nft_contract_withdrawal":
			users_updated_map[transaction.From] = true
			has_process.HasNFTContractWithdrawal = true
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
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(state_balances[transaction.From][transaction.CurrencyOrNftContractAddress], 10)
					if !ok {
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
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
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(state_balances[transaction.To][transaction.CurrencyOrNftContractAddress], 10)
					if !ok {
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
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
			has_process.HasDeposit = true
			if _, ok := state_balances[transaction.To]; ok {
				if _, ok := state_balances[transaction.To][transaction.CurrencyOrNftContractAddress]; ok {
					amount, ok := new(big.Int).SetString(transaction.AmountOrNftTokenId, 10)
					if !ok {
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(state_balances[transaction.To][transaction.CurrencyOrNftContractAddress], 10)
					if !ok {
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
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
			has_process.HasWithdrawal = true
			if _, ok := state_balances[transaction.From]; ok {
				if _, ok := state_balances[transaction.From][transaction.CurrencyOrNftContractAddress]; ok {
					amount, ok := new(big.Int).SetString(transaction.AmountOrNftTokenId, 10)
					if !ok {
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(state_balances[transaction.From][transaction.CurrencyOrNftContractAddress], 10)
					if !ok {
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
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
			has_process.HasContractWithdrawal = true
			if _, ok := state_balances[transaction.From]; ok {
				if _, ok := state_balances[transaction.From][transaction.CurrencyOrNftContractAddress]; ok {
					amount, ok := new(big.Int).SetString(transaction.AmountOrNftTokenId, 10)
					if !ok {
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
					}
					current_balance, ok := new(big.Int).SetString(state_balances[transaction.From][transaction.CurrencyOrNftContractAddress], 10)
					if !ok {
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int")
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

			// NFT TRANSFER
			users_updated_map[trade.From] = true
			users_updated_map[trade.To] = true
			if _, ok := state_balances[trade.To]; ok {
				state_balances[trade.To][trade.NftContractAddress+"-"+trade.NftTokenId] = "yes"
			} else {
				state_balances[trade.To] = make(map[string]string)
				state_balances[trade.To][trade.NftContractAddress+"-"+trade.NftTokenId] = "yes"
			}
			if _, ok := state_balances[trade.From]; ok {
				delete(state_balances[trade.From], trade.NftContractAddress+"-"+trade.NftTokenId)
			}
			if len(state_balances[trade.From]) == 0 {
				delete(state_balances, trade.From)
			}

			// TOKEN TRANSFER
			users_updated_map[trade.From] = true
			users_updated_map[trade.To] = true
			if _, ok := state_balances[trade.To]; ok {
				if _, ok := state_balances[trade.To][trade.Currency]; ok {
					amount, ok := new(big.Int).SetString(trade.BuyAmount, 10)
					if !ok {
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int nft trade 1")
					}
					current_balance, ok := new(big.Int).SetString(state_balances[trade.To][trade.Currency], 10)
					if !ok {
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int nft trade 2")
					}
					new_amt := new(big.Int)
					new_amt.Sub(current_balance, amount)
					state_balances[trade.To][trade.Currency] = new_amt.String()
				} else {
					state_balances[trade.To][trade.Currency] = "-" + trade.BuyAmount
				}
			} else {
				state_balances[trade.To] = make(map[string]string)
				state_balances[trade.To][trade.Currency] = "-" + trade.BuyAmount
			}
			if _, ok := state_balances[trade.From]; ok {
				if _, ok := state_balances[trade.From][trade.Currency]; ok {
					amount, ok := new(big.Int).SetString(trade.BuyAmount, 10)
					if !ok {
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int nft trade 3")
					}
					current_balance, ok := new(big.Int).SetString(state_balances[trade.From][trade.Currency], 10)
					if !ok {
						return state_balances, has_process, users_updated_map, user_nonce_tracker, fmt.Errorf("error converting amount to big int nft trade 4")
					}
					new_amt := new(big.Int)
					new_amt.Add(amount, current_balance)
					state_balances[trade.From][trade.Currency] = new_amt.String()
				} else {
					state_balances[trade.From][trade.Currency] = trade.BuyAmount
				}
			} else {
				state_balances[trade.From] = make(map[string]string)
				state_balances[trade.From][trade.Currency] = trade.BuyAmount
			}
		}
	}

	return state_balances, has_process, users_updated_map, user_nonce_tracker, nil
}
