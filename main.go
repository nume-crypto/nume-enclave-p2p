package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"
)

type ResponseBody struct {
	SettlementId                  uint              `json:"settlementId" binding:"required"`
	Root                          string            `json:"root" binding:"required"`
	AggregatedSignature           string            `json:"aggregatedSignature" binding:"required"`
	AggregatedPublicKeyComponents []string          `json:"aggregatedPublicKeyComponents" binding:"required"`
	BlockNumber                   string            `json:"blockNumber" binding:"required"`
	Type                          string            `json:"type" binding:"required"`
	QueueHash                     string            `json:"queueHash" binding:"required"`
	QueueIndex                    int               `json:"queueIndex" binding:"required"`
	WithdrawalHash                string            `json:"withdrawalHash" binding:"required"`
	WithdrawalAmounts             []string          `json:"withdrawalAmounts" binding:"required"`
	WithdrawalAddresses           []string          `json:"withdrawalAddresses" binding:"required"`
	WithdrawalTokenIndex          []uint            `json:"withdrawalTokenIndex" binding:"required"`
	Message                       string            `json:"message" binding:"required"`
	UsersUpdated                  map[string]string `json:"usersUpdated" binding:"required"`
	MissingUserIds                []uint            `json:"missingUserIds" binding:"required"`
	ActiveUserIds                 []uint            `json:"activeUserIds" binding:"required"`
	SignatureRecordedAt           time.Time         `json:"signatureRecordedAt" binding:"required"`
	SettlementStartedAt           time.Time         `json:"settlementStartedAt" binding:"required"`
}

func main() {

	defer TimeTrack(time.Now(), "main")
	settlement_started_at := time.Now()

	input_data, md5_sum_str, err := GetData("./data")
	if err != nil {
		fmt.Println("read err", err)
	}
	hFunc := NewMiMC()

	max_num_balances, err := strconv.Atoi(input_data.MetaData["max_num_balances"].(string))
	if err != nil {
		fmt.Println("error in max_num_balances")
		return
	}
	max_num_users, err := strconv.Atoi(input_data.MetaData["max_num_users"].(string))
	if err != nil {
		fmt.Println("error in max_num_users")
		return
	}
	var prev_val_hash = make([][]byte, max_num_users)
	var empty_balances_data = make([][]byte, max_num_balances)
	zero, ok := new(big.Int).SetString("0", 16)
	if !ok {
		fmt.Println("error in decoding cb hash")
		return
	}
	hFunc.Write(zero.Bytes())
	zero_hash := hFunc.Sum(nil)
	hFunc.Reset()
	for i := 0; i < max_num_balances; i++ {
		empty_balances_data[i] = zero_hash
	}
	empty_balances_tree := NewMerkleTree(empty_balances_data, hFunc)

	var wg sync.WaitGroup
	for i, u := range input_data.MetaData["users_ordered"].([]interface{}) {
		wg.Add(1)
		go func(i int, u interface{}) {
			balances_root, ok := GetBalancesRoot(input_data.OldUserBalances[u.(string)], max_num_balances)
			if !ok {
				fmt.Println("error in getting balances root")
				return
			}
			leaf, ok := GetLeafHash(u.(string), balances_root)
			if !ok {
				fmt.Println("error in getting leaf hash")
				return
			}
			prev_val_hash[i] = leaf
			wg.Done()
		}(i, u)
	}
	wg.Wait()

	for i := len(input_data.OldUserBalances); i < max_num_users; i++ {
		wg.Add(1)
		go func(i int) {
			leaf, ok := GetLeafHash(strconv.FormatUint(uint64(i), 16), hex.EncodeToString(empty_balances_tree.Root))
			if !ok {
				fmt.Println("error in getting leaf hash")
				return
			}
			prev_val_hash[i] = leaf
			wg.Done()
		}(i)
	}
	wg.Wait()
	tree := NewMerkleTree(prev_val_hash, hFunc)
	// 5 is default currency index when they register and dont do any deposit or transactions
	init_state_balances := input_data.OldUserBalances
	for _, u := range input_data.MetaData["users_ordered"].([]interface{}) {
		if _, ok := init_state_balances[u.(string)]; !ok {
			init_state_balances[u.(string)] = make(map[uint]string)
			init_state_balances[u.(string)][5] = "0"
		}
	}

	new_balances, settlement_type, users_updated_map, err := TransitionState(init_state_balances, input_data.Transactions, input_data.UserKeys)
	if err != nil {
		fmt.Println(err)
		fmt.Println("error in transition state")
		return
	}
	result := NestedMapsEqual(new_balances, input_data.NewUserBalances)
	if !result {
		fmt.Println("new_balances and input_data.NewUserBalances are not equal")
		return
	}
	bn := int(input_data.MetaData["block_number"].(float64))
	users_updated := make(map[string]string)
	var prev_tree_root []byte
	prev_tree_root = append(prev_tree_root, tree.Root...)
	for i, u := range input_data.MetaData["users_ordered"].([]interface{}) {
		balances_root, ok := GetBalancesRoot(input_data.NewUserBalances[u.(string)], max_num_balances)
		if !ok {
			fmt.Println("error in getting balances root")
			return
		}
		leaf, ok := GetLeafHash(u.(string), balances_root)
		if !ok {
			fmt.Println("error in getting leaf hash")
			return
		}
		if users_updated_map[u.(string)] {
			users_updated[u.(string)] = hex.EncodeToString(leaf)
		} else if !bytes.Equal(leaf, prev_val_hash[i]) {
			users_updated[u.(string)] = hex.EncodeToString(leaf)
		}
		tree.UpdateLeaf(i, hex.EncodeToString(leaf))
	}
	var new_tree_root []byte
	new_tree_root = append(new_tree_root, tree.Root...)
	fmt.Println(hex.EncodeToString(prev_tree_root), hex.EncodeToString(new_tree_root), md5_sum_str, bn)

	message := ""
	var queue_hash []byte
	var queue_index int

	var withdrawal_hash []byte
	withdrawal_amounts := make([]string, 0)
	withdrawal_addresses := make([]string, 0)
	withdrawal_tokens := make([]uint, 0)
	var queue_len int

	md5_sum_str = "0000000000000000000000000000000000000000000000000000000000000000"
	switch settlement_type {
	case "notarizeSettlementSignedByAllUsers":
		message = SettlementSignedByAllUsersMessage(hex.EncodeToString(prev_tree_root), hex.EncodeToString(new_tree_root), md5_sum_str, bn)
	case "notarizeSettlementWithDepositsAndWithdrawals":
		last_handled_queue_index, err := strconv.Atoi(input_data.MetaData["last_handled_queue_index"].(string))
		if err != nil {
			fmt.Println("error in queue size")
			return
		}
		queue_hash, queue_len, ok = QueueHash(input_data.Transactions)
		if !ok {
			fmt.Println("error in getting queue hash")
			return
		}
		withdrawal_hash, withdrawal_amounts, withdrawal_addresses, withdrawal_tokens, ok = WithdrawalHash(input_data.Transactions)
		if !ok {
			fmt.Println("error in getting withdrawal_hash")
			return
		}
		queue_index = queue_len + last_handled_queue_index
		message = SettlementWithDepositsAndWithdrawalsMessage(hex.EncodeToString(prev_tree_root), hex.EncodeToString(new_tree_root), md5_sum_str, uint(queue_index), hex.EncodeToString(queue_hash), hex.EncodeToString(withdrawal_hash), bn)
	case "notarizeSettlementWithDeposits":
		last_handled_queue_index, err := strconv.Atoi(input_data.MetaData["last_handled_queue_index"].(string))
		if err != nil {
			fmt.Println("error in queue size")
			return
		}
		queue_hash, queue_index, ok = QueueHash(input_data.Transactions)
		if !ok {
			fmt.Println("error in getting queue hash")
			return
		}
		message = SettlementWithDepositsMessage(hex.EncodeToString(prev_tree_root), hex.EncodeToString(new_tree_root), md5_sum_str, uint(queue_index+last_handled_queue_index), hex.EncodeToString(queue_hash), bn)
	case "notarizeSettlementWithWithdrawals":
		withdrawal_hash, withdrawal_amounts, withdrawal_addresses, withdrawal_tokens, ok = WithdrawalHash(input_data.Transactions)
		if !ok {
			fmt.Println("error in getting withdrawal_hash")
			return
		}
		message = SettlementWithWithdrawalsMessage(hex.EncodeToString(prev_tree_root), hex.EncodeToString(new_tree_root), md5_sum_str, hex.EncodeToString(withdrawal_hash), bn)
	default:
		fmt.Println("error in message type", settlement_type)
		return
	}
	signature, aggregated_public_key, failed_to_decrypt, successfully_decrypted, err := SignMessage(message, input_data.UserKeys)
	if err != nil {
		fmt.Println(err)
		return
	}

	bn_str := strconv.Itoa(bn)
	signature_recorded_at := time.Now()
	response := ResponseBody{
		SettlementId:                  uint(input_data.MetaData["settlement_id"].(float64)),
		Root:                          hex.EncodeToString(new_tree_root),
		AggregatedSignature:           signature,
		AggregatedPublicKeyComponents: aggregated_public_key,
		Message:                       message,
		BlockNumber:                   bn_str,
		SignatureRecordedAt:           signature_recorded_at,
		SettlementStartedAt:           settlement_started_at,
		MissingUserIds:                failed_to_decrypt,
		ActiveUserIds:                 successfully_decrypted,
		Type:                          settlement_type,
		QueueHash:                     "0x" + hex.EncodeToString(queue_hash),
		QueueIndex:                    queue_index,
		WithdrawalHash:                "0x" + hex.EncodeToString(withdrawal_hash),
		WithdrawalAmounts:             withdrawal_amounts,
		WithdrawalAddresses:           withdrawal_addresses,
		WithdrawalTokenIndex:          withdrawal_tokens,
		UsersUpdated:                  users_updated,
	}
	PrettyPrint(response)

}
