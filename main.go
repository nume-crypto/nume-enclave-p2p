package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

type ResponseBody struct {
	MerkleRoot                    string    `json:"merkleRoot" binding:"required"`
	AggregatedSignature           string    `json:"aggregatedSignature" binding:"required"`
	AggregatedPublicKeyComponents []string  `json:"aggregatedPublicKeyComponents" binding:"required"`
	BlockNumber                   string    `json:"blockNumber" binding:"required"`
	IsVerified                    bool      `json:"isVerified" binding:"required"`
	MissingUserIds                []uint    `json:"missingUserIds" binding:"required"`
	SignatureRecordedAt           time.Time `json:"signatureRecordedAt" binding:"required"`
	SettlementStartedAt           time.Time `json:"settlementStartedAt" binding:"required"`
	ActiveUserIds                 []uint    `json:"activeUserIds" binding:"required"`
}

func main() {

	defer TimeTrack(time.Now(), "main")
	settlement_started_at := time.Now()

	input_data, md5_sum_str, err := GetData("./data")
	if err != nil {
		fmt.Println("read err", err)
	}
	hFunc := NewMiMC()

	max_num_balances, err := strconv.Atoi(input_data.MetaData["max_num_balances"])
	if err != nil {
		fmt.Println("error in max_num_balances")
		return
	}
	max_num_users, err := strconv.Atoi(input_data.MetaData["max_num_users"])
	if err != nil {
		fmt.Println("error in max_num_users")
		return
	}
	var prev_val_hash = make([][]byte, max_num_users)
	var val_hash = make([][]byte, max_num_users)
	var empty_balances_data = make([][]byte, max_num_balances)
	for i := 0; i < max_num_balances; i++ {
		cb, ok := new(big.Int).SetString("0", 16)
		if !ok {
			fmt.Println("error in decoding cb hash")
			return
		}
		hFunc.Write(cb.Bytes())
		empty_balances_data[i] = hFunc.Sum(nil)
		hFunc.Reset()
	}

	empty_balances_tree := NewMerkleTree(empty_balances_data, hFunc)

	for i := 0; i < max_num_users; i++ {
		leaf, ok := GetLeafHash(strconv.FormatUint(uint64(i), 16), hex.EncodeToString(empty_balances_tree.Root))
		if !ok {
			fmt.Println("error in getting leaf hash")
			return
		}
		prev_val_hash[i] = leaf
		val_hash[i] = leaf
	}

	prev_tree := NewMerkleTree(prev_val_hash, hFunc)

	new_balances, err := TransitionState(input_data.OldUserBalances, input_data.Transactions, input_data.UserKeys)
	if err != nil {
		fmt.Println(err)
		fmt.Println("error in transition state")
		return
	}
	result := NestedMapsEqual(new_balances, input_data.NewUserBalances)
	if !result {
		fmt.Println("new_balances", new_balances, input_data.NewUserBalances)
		fmt.Println("new_balances and input_data.NewUserBalances are not equal")
		return
	}
	tree := NewMerkleTree(val_hash, hFunc)
	bn, err := strconv.Atoi(input_data.MetaData["block_number"])
	if err != nil {
		fmt.Println("error in block number")
		return
	}

	message := ""
	switch input_data.MetaData["message_type"] {
	case "notarizeSettlementSignedByAllUsers":
		message = SettlementSignedByAllUsersMessage(hex.EncodeToString(prev_tree.Root), hex.EncodeToString(tree.Root), md5_sum_str, bn)
	case "notarizeSettlementWithDepositsAndWithdrawals":
		last_handled_queue_index, err := strconv.Atoi(input_data.MetaData["last_handled_queue_index"])
		if err != nil {
			fmt.Println("error in queue size")
			return
		}
		queue_hash, queue_index, ok := QueueHash(input_data.Transactions)
		if !ok {
			fmt.Println("error in getting queue hash")
			return
		}
		withdrawal_hash, ok := WithdrawalHash(input_data.Transactions)
		if !ok {
			fmt.Println("error in getting withdrawal_hash")
			return
		}
		message = SettlementWithDepositsAndWithdrawalsMessage(hex.EncodeToString(prev_tree.Root), hex.EncodeToString(tree.Root), md5_sum_str, uint(queue_index+last_handled_queue_index), hex.EncodeToString(queue_hash), hex.EncodeToString(withdrawal_hash), bn)
	case "notarizeSettlementWithDeposits":
		last_handled_queue_index, err := strconv.Atoi(input_data.MetaData["last_handled_queue_index"])
		if err != nil {
			fmt.Println("error in queue size")
			return
		}
		queue_hash, queue_index, ok := QueueHash(input_data.Transactions)
		if !ok {
			fmt.Println("error in getting queue hash")
			return
		}
		message = SettlementWithDepositsMessage(hex.EncodeToString(prev_tree.Root), hex.EncodeToString(tree.Root), md5_sum_str, uint(queue_index+last_handled_queue_index), hex.EncodeToString(queue_hash), bn)
	case "notarizeSettlementWithWithdrawals":
		withdrawal_hash, ok := WithdrawalHash(input_data.Transactions)
		if !ok {
			fmt.Println("error in getting withdrawal_hash")
			return
		}
		message = SettlementWithWithdrawalsMessage(hex.EncodeToString(prev_tree.Root), hex.EncodeToString(tree.Root), md5_sum_str, hex.EncodeToString(withdrawal_hash), bn)
	default:
		fmt.Println("error in message type")
		return
	}
	signature, aggregated_public_key, failed_to_decrypt, successfully_decrypted, err := SignMessage(message, input_data.UserKeys)
	if err != nil {
		fmt.Println(err)
		return
	}

	signature_recorded_at := time.Now()
	response := ResponseBody{
		MerkleRoot:                    hex.EncodeToString(tree.Root),
		AggregatedSignature:           signature,
		AggregatedPublicKeyComponents: aggregated_public_key,
		BlockNumber:                   input_data.MetaData["block_number"],
		SignatureRecordedAt:           signature_recorded_at,
		SettlementStartedAt:           settlement_started_at,
		MissingUserIds:                failed_to_decrypt,
		ActiveUserIds:                 successfully_decrypted,
	}
	PrettyPrint(response)

}
