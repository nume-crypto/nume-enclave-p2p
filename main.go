package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	solsha3 "github.com/miguelmota/go-solidity-sha3"
	progressbar "github.com/schollz/progressbar/v3"
)

type SettlementRequest struct {
	SettlementId                         uint              `json:"settlementId" binding:"required"`
	Root                                 string            `json:"root" binding:"required"`
	NftRoot                              string            `json:"nftRoot" binding:"required"`
	AggregatedSignature                  string            `json:"aggregatedSignature" binding:"required"`
	AggregatedPublicKeyComponents        []string          `json:"aggregatedPublicKeyComponents" binding:"required"`
	BlockNumber                          string            `json:"blockNumber" binding:"required"`
	QueueHash                            string            `json:"queueHash" binding:"required"` // deposit
	QueueIndex                           int               `json:"queueIndex"`
	NftQueueHash                         string            `json:"nftQueueHash" binding:"required"` // deposit nft
	NftQueueIndex                        int               `json:"nftQueueIndex"`
	WithdrawalHash                       string            `json:"withdrawalHash" binding:"required"` // withdrawal
	WithdrawalAmountOrTokenId            []string          `json:"withdrawalAmountOrTokenId" binding:"required"`
	WithdrawalAddresses                  []string          `json:"withdrawalAddresses" binding:"required"`
	WithdrawalCurrencyOrNftContract      []string          `json:"withdrawalCurrencyOrNftContract" binding:"required"`
	WithdrawalL2Minted                   []bool            `json:"withdrawalL2Minted" binding:"required"`
	WithdrawalType                       []int             `json:"withdrawalType" binding:"required"`
	ContractWithdrawalAddresses          []string          `json:"contractWithdrawalAddresses" binding:"required"` // contract withdrawal
	ContractWithdrawalAmounts            []string          `json:"contractWithdrawalAmounts" binding:"required"`
	ContractWithdrawalTokens             []string          `json:"contractWithdrawalTokes" binding:"required"`
	ContractWithdrawalQueueIndex         int               `json:"contractWithdrawalQueueIndex"`
	NftContractWithdrawalAddresses       []string          `json:"nftContractWithdrawalAddresses" binding:"required"` // nft contract withdrawal
	NftContractWithdrawalTokensIds       []string          `json:"nftContractWithdrawalTokenIds" binding:"required"`
	NftContractWithdrawalContractAddress []string          `json:"nftContractWithdrawalContractAddress" binding:"required"`
	NftContractWithdrawalQueueIndex      int               `json:"nftContractWithdrawalQueueIndex"`
	NftContractWithdrawalL2Minted        []bool            `json:"nftContractWithdrawalL2Minted" binding:"required"`
	Message                              string            `json:"message" binding:"required"` // message
	UsersUpdated                         map[string]string `json:"usersUpdated" binding:"required"`
	NftCollectionsCreated                map[int]string    `json:"nftCollectionsCreated" binding:"required"`
	SignatureRecordedAt                  time.Time         `json:"signatureRecordedAt" binding:"required"`
	SettlementStartedAt                  time.Time         `json:"settlementStartedAt" binding:"required"`
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
	max_num_collections, err := strconv.Atoi(input_data.MetaData["max_num_collections"].(string))
	if err != nil {
		fmt.Println("error in max_num_collections")
		return
	}
	var nft_collection_data = make([][]byte, max_num_collections)
	nft_zero_hash := solsha3.SoliditySHA3(
		[]string{"address", "address", "bytes32"},
		[]interface{}{
			"0x0000000000000000000000000000000000000000",
			"0x0000000000000000000000000000000000000000",
			"0x0000000000000000000000000000000000000000",
		},
	)
	for i := 0; i < len(input_data.OldNftCollections); i++ {
		types := []string{}
		values := []interface{}{}
		for _, t := range input_data.NewNftCollections[i]["MintUsers"].([]interface{}) {
			types = append(types, "address")
			values = append(values, t)
		}
		mint_users_hash := solsha3.SoliditySHA3(
			types,
			values)
		meta_hash := solsha3.SoliditySHA3(
			[]string{"uint256", "uint256", "bytes32", "bytes32"},
			[]interface{}{
				input_data.NewNftCollections[i]["MintStart"],
				input_data.NewNftCollections[i]["MintEnd"],
				solsha3.SoliditySHA3("string", input_data.NewNftCollections[i]["BaseUri"]),
				mint_users_hash,
			},
		)
		hash := solsha3.SoliditySHA3(
			[]string{"address", "address", "bytes32"},
			[]interface{}{
				input_data.NewNftCollections[i]["Owner"],
				input_data.NewNftCollections[i]["ContractAddress"],
				meta_hash,
			},
		)
		nft_collection_data[i] = hash
	}
	for i := len(input_data.OldNftCollections); i < max_num_collections; i++ {
		nft_collection_data[i] = nft_zero_hash
	}
	nft_collection_tree := NewMerkleTree(nft_collection_data)

	var prev_val_hash = make([][]byte, max_num_users)
	var empty_balances_data = make([][]byte, max_num_balances)
	zero_hash := solsha3.SoliditySHA3(
		[]string{"address", "uint256", "uint256"},
		[]interface{}{
			"0x0000000000000000000000000000000000000000",
			"0",
			"0",
		},
	)
	for i := 0; i < max_num_balances; i++ {
		empty_balances_data[i] = zero_hash
	}
	bar := progressbar.Default(int64(max_num_users))
	empty_balances_tree := NewMerkleTree(empty_balances_data)
	old_user_nonce := input_data.MetaData["old_users_nonce"].(map[string]interface{})
	var wg sync.WaitGroup
	for i, u := range input_data.MetaData["users_ordered"].([]interface{}) {
		wg.Add(1)
		go func(i int, u interface{}) {
			bar.Add(1)
			balances_root, ok := GetBalancesRoot(input_data.OldUserBalances[u.(string)], input_data.OldUserBalanceOrder[u.(string)], max_num_balances)
			if !ok {
				fmt.Println("error in getting balances root")
				return
			}
			nonce := uint(0)
			if old_user_nonce[u.(string)] != nil {
				nonce = uint(old_user_nonce[u.(string)].(float64))
			}
			leaf := GetLeafHash(fmt.Sprintf("%040s", u.(string)), "0x"+balances_root, nonce)
			prev_val_hash[i] = leaf
			wg.Done()
		}(i, u)
	}
	wg.Wait()
	for i := len(input_data.OldUserBalances); i < max_num_users; i++ {
		wg.Add(1)
		go func(i int) {
			bar.Add(1)
			leaf := GetLeafHash("0x"+fmt.Sprintf("%040s", strconv.FormatUint(uint64(i), 16)), "0x"+hex.EncodeToString(empty_balances_tree.Root), 0)
			prev_val_hash[i] = leaf
			wg.Done()
		}(i)
	}
	wg.Wait()
	tree := NewMerkleTree(prev_val_hash)
	currencies := []string{}
	for _, c := range input_data.MetaData["currencies"].([]interface{}) {
		currencies = append(currencies, c.(string))
	}
	input_transactions := []Transaction{}
	for _, tx := range input_data.Transactions {
		if t, ok := tx.(map[string]interface{}); ok {
			if t["Type"] != "nft_trade" {
				transaction := Transaction{
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
				input_transactions = append(input_transactions, transaction)
			}
		}
	}
	init_state_balances := input_data.OldUserBalances
	new_balances, has_process, users_updated_map, user_nonce_tracker, err := TransitionState(init_state_balances, input_data.Transactions, currencies, input_data.NewNftCollections)
	if err != nil {
		fmt.Println(err)
		fmt.Println("error in transition state")
		return
	}

	for _, v := range input_data.MetaData["users_ordered"].([]interface{}) {
		if _, ok := new_balances[v.(string)]; !ok {
			new_balances[v.(string)] = make(map[string]string)
			new_balances[v.(string)]["0x0000000000000000000000000000000000000000"] = "0"
		} else {
			if _, ok := new_balances[v.(string)]["0x0000000000000000000000000000000000000000"]; !ok {
				new_balances[v.(string)]["0x0000000000000000000000000000000000000000"] = "0"
			}
		}
	}

	result := NestedMapsEqual(new_balances, input_data.NewUserBalances)
	if !result {
		PrettyPrint("new_balances", new_balances)
		PrettyPrint("input_data.NewUserBalances", input_data.NewUserBalances)
		fmt.Println("new_balances and input_data.NewUserBalances are not equal")
		return
	}
	bn := int(input_data.MetaData["block_number"].(float64))
	users_updated := make(map[string]string)
	var prev_tree_root []byte
	var prev_ctree_root []byte
	prev_ctree_root = append(prev_ctree_root, nft_collection_tree.Root...)
	prev_tree_root = append(prev_tree_root, tree.Root...)
	for i, u := range input_data.MetaData["users_ordered"].([]interface{}) {
		balances_root, ok := GetBalancesRoot(input_data.NewUserBalances[u.(string)], input_data.NewUserBalanceOrder[u.(string)], max_num_balances)
		if !ok {
			fmt.Println("error in getting balances root new")
			return
		}
		leaf := GetLeafHash(u.(string), "0x"+balances_root, uint(user_nonce_tracker[u.(string)]))
		if users_updated_map[u.(string)] {
			users_updated[u.(string)] = hex.EncodeToString(leaf)
		} else if !bytes.Equal(leaf, prev_val_hash[i]) {
			users_updated[u.(string)] = hex.EncodeToString(leaf)
		}
		tree.UpdateLeaf(i, hex.EncodeToString(leaf))
	}
	updated_ntf_collections := make(map[int]string)
	for i := len(input_data.OldNftCollections); i < len(input_data.NewNftCollections); i++ {
		types := []string{}
		values := []interface{}{}
		for _, t := range input_data.NewNftCollections[i]["MintUsers"].([]interface{}) {
			types = append(types, "address")
			values = append(values, t)
		}
		mint_users_hash := solsha3.SoliditySHA3(
			types,
			values)
		meta_hash := solsha3.SoliditySHA3(
			[]string{"uint256", "uint256", "bytes32", "bytes32"},
			[]interface{}{
				input_data.NewNftCollections[i]["MintStart"],
				input_data.NewNftCollections[i]["MintEnd"],
				solsha3.SoliditySHA3("string", input_data.NewNftCollections[i]["BaseUri"]),
				mint_users_hash,
			},
		)
		hash := solsha3.SoliditySHA3(
			[]string{"address", "address", "bytes32"},
			[]interface{}{
				input_data.NewNftCollections[i]["Owner"],
				input_data.NewNftCollections[i]["ContractAddress"],
				meta_hash,
			},
		)
		nft_collection_tree.UpdateLeaf(i, hex.EncodeToString(hash))
		updated_ntf_collections[i] = hex.EncodeToString(hash)
	}

	// find leaves and upload
	leafMap := make(map[int]string)
	for i := 0; i < len(tree.Nodes[0]); i++ {
		leaf := tree.Nodes[0][i].Data
		leafMap[i] = hex.EncodeToString(leaf)
	}

	// Convert the `leafMap` map to a JSON string.
	leafData, err := json.Marshal(leafMap)
	if err != nil {
		panic(err)
	}
	// Upload Merkle leafData to S3
	md5_leaf_data := md5.Sum(bytes.TrimRight(leafData, "\n"))
	md5_leaf_data_str := hex.EncodeToString(md5_leaf_data[:])
	fmt.Println("md5_leaf_data_str", md5_leaf_data_str)

	//Upload Public Transaction Data to S3
	// public_transaction_data := GenerateTransactionPublicData(input_transactions, input_data.AddressPublicKeyData, input_data.MetaData["block_number"].(float64))
	// _ = public_transaction_data

	var new_tree_root []byte
	new_tree_root = append(new_tree_root, tree.Root...)
	var new_ctree_root []byte
	new_ctree_root = append(new_ctree_root, nft_collection_tree.Root...)
	fmt.Println(hex.EncodeToString(prev_tree_root), hex.EncodeToString(new_tree_root), md5_sum_str, bn, hex.EncodeToString(prev_ctree_root), hex.EncodeToString(new_ctree_root))

	message := ""
	var queue_hash []byte
	var queue_index int
	var queue_len int

	var nft_queue_hash []byte
	var nft_queue_index int
	var nft_queue_len int

	var withdrawal_hash []byte
	withdrawal_amounts_or_token_id := make([]string, 0)
	withdrawal_addresses := make([]string, 0)
	withdrawal_currency_or_nft_contract := make([]string, 0)
	withdrawal_l2_minted := make([]bool, 0)
	withdrawal_type := make([]int, 0)

	cw_addresses := make([]string, 0)
	cw_token_ids := make([]string, 0)
	cw_amounts := make([]string, 0)
	var cw_queue_hash []byte
	var cw_queue_index int
	var cw_queue_len int

	nft_cw_addresses := make([]string, 0)
	nft_cw_token_ids := make([]string, 0)
	nft_cw_amounts := make([]string, 0)
	var nft_cw_queue_hash []byte
	var nft_cw_queue_index int
	var nft_cw_queue_len int
	nft_cw_l2_minted := make([]bool, 0)
	var ok bool

	md5_sum_str = "0000000000000000000000000000000000000000000000000000000000000000"
	md5_leaf_data_str = "0000000000000000000000000000000000000000000000000000000000000000"

	message = hex.EncodeToString(prev_tree_root) + hex.EncodeToString(new_tree_root) + fmt.Sprintf("%064s", md5_sum_str) + fmt.Sprintf("%064x", bn) + hex.EncodeToString(prev_ctree_root) + hex.EncodeToString(new_ctree_root)
	if has_process.HasDeposit {
		last_handled_queue_index, err := strconv.Atoi(input_data.MetaData["last_handled_queue_index"].(string))
		if err != nil {
			fmt.Println("error in queue size")
			return
		}
		queue_hash, queue_len, ok = QueueHash(input_transactions)
		if !ok {
			fmt.Println("error in getting queue hash")
			return
		}
		message += fmt.Sprintf("%064x", queue_len+last_handled_queue_index) + fmt.Sprintf("%064s", hex.EncodeToString(queue_hash))
		queue_index = queue_len + last_handled_queue_index
	}
	if has_process.HasContractWithdrawal {
		last_handled_cw_queue_index, err := strconv.Atoi(input_data.MetaData["last_handled_cw_queue_index"].(string))
		if err != nil {
			fmt.Println("error in queue size")
			return
		}
		cw_queue_hash, cw_queue_len, cw_addresses, cw_amounts, cw_token_ids, ok = WithdrawalQueueHash(input_transactions)
		if !ok {
			fmt.Println("error in getting cw queue hash")
			return
		}
		message += fmt.Sprintf("%064x", cw_queue_len+last_handled_cw_queue_index) + fmt.Sprintf("%064s", hex.EncodeToString(cw_queue_hash))
		cw_queue_index = cw_queue_len + last_handled_cw_queue_index
	}
	if has_process.HasNFTDeposit {
		last_handled_nft_queue_index, err := strconv.Atoi(input_data.MetaData["last_handled_nft_queue_index"].(string))
		if err != nil {
			fmt.Println("error in queue size")
			return
		}
		nft_queue_hash, nft_queue_len, ok = QueueHash(input_transactions)
		if !ok {
			fmt.Println("error in getting queue hash")
			return
		}
		message += fmt.Sprintf("%064x", nft_queue_len+last_handled_nft_queue_index) + fmt.Sprintf("%064s", hex.EncodeToString(queue_hash))
		nft_queue_index = nft_queue_len + last_handled_nft_queue_index
	}
	if has_process.HasNFTContractWithdrawal {
		last_handled_nft_cw_queue_index, err := strconv.Atoi(input_data.MetaData["last_handled_nft_cw_queue_index"].(string))
		if err != nil {
			fmt.Println("error in queue size")
			return
		}
		nft_cw_queue_hash, nft_cw_queue_len, nft_cw_addresses, nft_cw_amounts, nft_cw_token_ids, nft_cw_l2_minted, ok = NftWithdrawalQueueHash(input_transactions)
		if !ok {
			fmt.Println("error in getting cw queue hash")
			return
		}
		message += fmt.Sprintf("%064x", nft_cw_queue_len+last_handled_nft_cw_queue_index) + fmt.Sprintf("%064s", hex.EncodeToString(nft_cw_queue_hash))
		nft_cw_queue_index = nft_cw_queue_len + last_handled_nft_cw_queue_index
	}
	if has_process.HasWithdrawal {
		withdrawal_hash, withdrawal_amounts_or_token_id, withdrawal_l2_minted, withdrawal_addresses, withdrawal_currency_or_nft_contract, withdrawal_type, ok = WithdrawalHash(input_transactions)
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
		SettlementId:                         uint(input_data.MetaData["settlement_id"].(float64)),
		Root:                                 hex.EncodeToString(new_tree_root),
		NftRoot:                              hex.EncodeToString(new_ctree_root),
		AggregatedSignature:                  signature,
		AggregatedPublicKeyComponents:        aggregated_public_key,
		Message:                              message,
		BlockNumber:                          bn_str,
		SignatureRecordedAt:                  signature_recorded_at,
		SettlementStartedAt:                  settlement_started_at,
		QueueHash:                            "0x" + hex.EncodeToString(queue_hash),
		QueueIndex:                           queue_index,
		NftQueueHash:                         "0x" + hex.EncodeToString(nft_queue_hash),
		NftQueueIndex:                        nft_queue_index,
		WithdrawalHash:                       "0x" + hex.EncodeToString(withdrawal_hash),
		WithdrawalAmountOrTokenId:            withdrawal_amounts_or_token_id,
		WithdrawalAddresses:                  withdrawal_addresses,
		WithdrawalCurrencyOrNftContract:      withdrawal_currency_or_nft_contract,
		WithdrawalL2Minted:                   withdrawal_l2_minted,
		WithdrawalType:                       withdrawal_type,
		ContractWithdrawalAddresses:          cw_addresses,
		ContractWithdrawalQueueIndex:         cw_queue_index,
		ContractWithdrawalAmounts:            cw_amounts,
		ContractWithdrawalTokens:             cw_token_ids,
		NftContractWithdrawalAddresses:       nft_cw_addresses,
		NftContractWithdrawalQueueIndex:      nft_cw_queue_index,
		NftContractWithdrawalTokensIds:       nft_cw_amounts,
		NftContractWithdrawalContractAddress: nft_cw_token_ids,
		NftContractWithdrawalL2Minted:        nft_cw_l2_minted,
		UsersUpdated:                         users_updated,
		NftCollectionsCreated:                updated_ntf_collections,
	}
	PrettyPrint("response", response)

}
