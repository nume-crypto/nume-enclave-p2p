package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"time"

	solsha3 "github.com/miguelmota/go-solidity-sha3"
)

type SettlementRequest struct {
	SettlementId                    uint              `json:"settlementId" binding:"required"`
	Root                            string            `json:"root" binding:"required"`
	AggregatedSignature             string            `json:"aggregatedSignature" binding:"required"`
	AggregatedPublicKeyComponents   []string          `json:"aggregatedPublicKeyComponents" binding:"required"`
	BlockNumber                     string            `json:"blockNumber" binding:"required"`
	ProcessId                       uint              `json:"processId" binding:"required"`
	QueueHash                       string            `json:"queueHash" binding:"required"` // deposit
	QueueIndex                      int               `json:"queueIndex"`
	WithdrawalHash                  string            `json:"withdrawalHash" binding:"required"` // withdrawal
	WithdrawalAmounts               []string          `json:"withdrawalAmounts" binding:"required"`
	WithdrawalAddresses             []string          `json:"withdrawalAddresses" binding:"required"`
	WithdrawalTokens                []string          `json:"withdrawalTokes" binding:"required"`
	ContractWithdrawalAddresses     []string          `json:"contractWithdrawalAddresses" binding:"required"` // contract withdrawal
	ContractWithdrawalAmounts       []string          `json:"contractWithdrawalAmounts" binding:"required"`
	ContractWithdrawalTokens        []string          `json:"contractWithdrawalTokes" binding:"required"`
	ContractWithdrawalQueueIndex    int               `json:"contractWithdrawalQueueIndex"`
	ContractWithdrawalHashedBlsKeys []string          `json:"contractWithdrawalHashedBlsKeys" binding:"required"`
	Message                         string            `json:"message" binding:"required"` // message
	UsersUpdated                    map[string]string `json:"usersUpdated" binding:"required"`
	SignatureRecordedAt             time.Time         `json:"signatureRecordedAt" binding:"required"`
	SettlementStartedAt             time.Time         `json:"settlementStartedAt" binding:"required"`
}

func main() {

	defer TimeTrack(time.Now(), "main")
	settlement_started_at := time.Now()

	input_data, md5_sum_str, err := GetData("./data")
	if err != nil {
		fmt.Println("read err", err)
	}

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
	zero_hash := solsha3.SoliditySHA3(
		[]string{"address", "uint256"},
		[]interface{}{
			"0x0000000000000000000000000000000000000000",
			"0",
		},
	)
	for i := 0; i < max_num_balances; i++ {
		empty_balances_data[i] = zero_hash
	}
	empty_balances_tree := NewMerkleTree(empty_balances_data)

	var wg sync.WaitGroup
	for i, u := range input_data.MetaData["users_ordered"].([]interface{}) {
		wg.Add(1)
		go func(i int, u interface{}) {
			balances_root, ok := GetBalancesRoot(input_data.OldUserBalances[u.(string)], input_data.UserBalanceOrder[u.(string)], max_num_balances)
			if !ok {
				fmt.Println("error in getting balances root")
				return
			}
			leaf := GetLeafHash(fmt.Sprintf("%040s", u.(string)), "0x"+balances_root)
			prev_val_hash[i] = leaf
			wg.Done()
		}(i, u)
	}
	wg.Wait()
	for i := len(input_data.OldUserBalances); i < max_num_users; i++ {
		wg.Add(1)
		go func(i int) {
			leaf := GetLeafHash("0x"+fmt.Sprintf("%040s", strconv.FormatUint(uint64(i), 16)), "0x"+hex.EncodeToString(empty_balances_tree.Root))
			prev_val_hash[i] = leaf
			wg.Done()
		}(i)
	}
	wg.Wait()
	tree := NewMerkleTree(prev_val_hash)

	init_state_balances := input_data.OldUserBalances
	block_number := input_data.MetaData["block_number"].(float64)
	new_balances, settlement_type, users_updated_map, err := TransitionState(init_state_balances, input_data.Transactions, int64(block_number))
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
		balances_root, ok := GetBalancesRoot(input_data.NewUserBalances[u.(string)], input_data.UserBalanceOrder[u.(string)], max_num_balances)
		if !ok {
			fmt.Println("error in getting balances root")
			return
		}
		leaf := GetLeafHash(u.(string), "0x"+balances_root)
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
	withdrawal_tokens := make([]string, 0)
	var queue_len int

	cw_addresses := make([]string, 0)
	cw_token_ids := make([]string, 0)
	cw_amounts := make([]string, 0)
	cw_bls_keys := make([]string, 0)
	var cw_queue_hash []byte
	var cw_queue_index int
	var cw_queue_len int
	var ok bool

	md5_sum_str = "0000000000000000000000000000000000000000000000000000000000000000"
	fmt.Println("settlement_type", settlement_type)
	// process ID 0: only L2 transactions (+4)
	// process ID 1: deposit (+2)
	// process ID 2: withdrawal (backend) (+1)
	// process ID 3: withdrawal (contract) (+2)
	// process ID 4: deposit + withdrawal (backend)
	// process ID 5: deposit + withdrawal (contract)
	// process ID 6: withdrawal (backend) + withdrawal (contract)
	// process ID 7: deposit + withdrawal (backend) + withdrawal (contract)
	message = "10a3035aa29d8146314c6b29928a0b3e99ba2585e1bb8c74ca5170b4368819f4" + hex.EncodeToString(new_tree_root) + fmt.Sprintf("%064s", md5_sum_str) + fmt.Sprintf("%064x", bn)
	if settlement_type == 1 || settlement_type == 4 || settlement_type == 5 || settlement_type == 7 {
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
		message += fmt.Sprintf("%064x", queue_len+last_handled_queue_index) + fmt.Sprintf("%064s", hex.EncodeToString(queue_hash))
		queue_index = queue_len + last_handled_queue_index
	}
	if settlement_type == 3 || settlement_type == 5 || settlement_type == 6 || settlement_type == 7 {
		last_handled_cw_queue_index, err := strconv.Atoi(input_data.MetaData["last_handled_cw_queue_index"].(string))
		if err != nil {
			fmt.Println("error in queue size")
			return
		}
		cw_queue_hash, cw_queue_len, cw_addresses, cw_amounts, cw_token_ids, cw_bls_keys, ok = WithdrawalQueueHash(input_data.Transactions)
		if !ok {
			fmt.Println("error in getting queue hash")
			return
		}
		message += fmt.Sprintf("%064x", cw_queue_len+last_handled_cw_queue_index) + fmt.Sprintf("%064s", hex.EncodeToString(cw_queue_hash))
		cw_queue_index = cw_queue_len + last_handled_cw_queue_index
	}
	if settlement_type == 2 || settlement_type == 4 || settlement_type == 6 || settlement_type == 7 {
		withdrawal_hash, withdrawal_amounts, withdrawal_addresses, withdrawal_tokens, ok = WithdrawalHash(input_data.Transactions)
		if !ok {
			fmt.Println("error in getting withdrawal_hash")
			return
		}
		message += fmt.Sprintf("%064s", hex.EncodeToString(withdrawal_hash))
	}

	signature, aggregated_public_key, _, _, err := SignMessage(message, input_data.ValidatorKeys)
	if err != nil {
		fmt.Println(err)
		return
	}

	bn_str := strconv.Itoa(bn)
	signature_recorded_at := time.Now()
	response := SettlementRequest{
		SettlementId:                    uint(input_data.MetaData["settlement_id"].(float64)),
		Root:                            hex.EncodeToString(new_tree_root),
		AggregatedSignature:             signature,
		AggregatedPublicKeyComponents:   aggregated_public_key,
		Message:                         message,
		BlockNumber:                     bn_str,
		SignatureRecordedAt:             signature_recorded_at,
		SettlementStartedAt:             settlement_started_at,
		ProcessId:                       settlement_type,
		QueueHash:                       "0x" + hex.EncodeToString(queue_hash),
		QueueIndex:                      queue_index,
		WithdrawalHash:                  "0x" + hex.EncodeToString(withdrawal_hash),
		WithdrawalAmounts:               withdrawal_amounts,
		WithdrawalAddresses:             withdrawal_addresses,
		WithdrawalTokens:                withdrawal_tokens,
		ContractWithdrawalAddresses:     cw_addresses,
		ContractWithdrawalQueueIndex:    cw_queue_index,
		ContractWithdrawalAmounts:       cw_amounts,
		ContractWithdrawalTokens:        cw_token_ids,
		ContractWithdrawalHashedBlsKeys: cw_bls_keys,
		UsersUpdated:                    users_updated,
	}
	PrettyPrint(response)

}
