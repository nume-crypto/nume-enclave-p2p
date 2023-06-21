package main

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestQueueItemHash(t *testing.T) {
	pub_key := "0x447bF33F7c7C925eb7674bCF590AeD4Aa57e656b"
	token := "0x0b6D9aB4c80889b65A61050470CBC5523d8Ce48D"
	amount := "25003"
	hash, ok := QueueItemHash(pub_key, token, amount)
	if !ok {
		t.Errorf("Failed to hash queue item")
		return
	}
	expected_hash := "19321d66228a380018281c8afcc4590447ca41a34162e29e501f7670b30d036b"
	if hex.EncodeToString(hash) != expected_hash {
		t.Errorf("Failed to hash hash expected %s got %s", expected_hash, hex.EncodeToString(hash))
		return
	}
}

func TestQueueHash(t *testing.T) {
	transactions := make([]Transaction, 0)
	transactions_file, err := os.Open("test_data/transactions.json")
	if err != nil {
		t.Errorf("Error opening test_data/transactions.json")
		return
	}
	err = json.NewDecoder(transactions_file).Decode(&transactions)
	if err != nil {
		t.Errorf("Error decoding json from test_data/transactions.json")
		return
	}
	defer transactions_file.Close()
	hash, queue_index, ok := QueueHash(transactions, "deposit")
	if !ok {
		t.Errorf("Failed to hash queue")
		return
	}
	if queue_index != 8 {
		t.Errorf("Failed to get queue index expected 16 got %d", queue_index)
		return
	}
	expected_hash := "86bd2b51b681b773e091c7bebabc96031a2ad3bb4d51f888d5c672b0a7635d36"
	if hex.EncodeToString(hash) != expected_hash {
		t.Errorf("Failed to hash hash expected %s got %s", expected_hash, hex.EncodeToString(hash))
		return
	}
}

func TestWithdrawalHash(t *testing.T) {
	transactions := make([]Transaction, 0)
	transactions_file, err := os.Open("test_data/transactions.json")
	if err != nil {
		t.Errorf("Error opening test_data/transactions.json")
		return
	}
	err = json.NewDecoder(transactions_file).Decode(&transactions)
	if err != nil {
		t.Errorf("Error decoding json from test_data/transactions.json")
		return
	}
	defer transactions_file.Close()
	withdrawal_hash, withdrawal_amt_or_token_id, withdrawal_l2_minted, withdrawal_addresses, withdrawal_currency_or_nft, _, ok := WithdrawalHash(transactions)
	_ = withdrawal_l2_minted
	if !ok {
		t.Errorf("Failed to hash withdrawal")
		return
	}
	expected_withdrawal_amt_or_token_id := []string{"900000000000000000", "520000", "550000000000000064", "11", "13"}
	if !reflect.DeepEqual(withdrawal_amt_or_token_id, expected_withdrawal_amt_or_token_id) {
		t.Errorf("Failed to get withdrawal amounts expected %v got %v", expected_withdrawal_amt_or_token_id, withdrawal_amt_or_token_id)
		return
	}
	expected_withdrawal_addresses := []string{"0xCcFf350Ef46B85228d6650a802107e58BF6A32Ab", "0x1b34B2f706cDA183E4818D2ceaF58253CcAb3428", "0x11c830B25a15E39006094377fDc409c11C002B48", "0xCcFf350Ef46B85228d6650a802107e58BF6A32Ab", "0x46714661eECB6F07065DCb4BF3D9b772dcefa63a"}
	if !reflect.DeepEqual(withdrawal_addresses, expected_withdrawal_addresses) {
		t.Errorf("Failed to get withdrawal addresses expected %v got %v", expected_withdrawal_addresses, withdrawal_addresses)
		return
	}
	expected_withdrawal_currency_or_nft := []string{"0xEe146Fac7b2fce5FdBE31C36d89cF92f6b006F80", "0xCE47C48fDF8c9355FDbE4DacC1e1954914D65Be6", "0xEe146Fac7b2fce5FdBE31C36d89cF92f6b006F80", "0xedb6375347e060b055d6af9842ba8c55e3d93e3a", "0xedb6375347e060b055d6af9842ba8c55e3d93e3a"}
	if !reflect.DeepEqual(withdrawal_currency_or_nft, expected_withdrawal_currency_or_nft) {
		t.Errorf("Failed to get withdrawal tokens expected %v got %v", expected_withdrawal_currency_or_nft, withdrawal_currency_or_nft)
		return
	}
	expected_hash := "0eec6909b065dfe79f3691feeb8116855d7cff16c8d1a9e265ec33457945008c"
	if hex.EncodeToString(withdrawal_hash) != expected_hash {
		t.Errorf("Failed to hash hash expected %s got %s", expected_hash, hex.EncodeToString(withdrawal_hash))
		return
	}
}

func TestWithdrawalQueueHash(t *testing.T) {
	transactions := make([]Transaction, 0)
	transactions_file, err := os.Open("test_data/transactions.json")
	if err != nil {
		t.Errorf("Error opening test_data/transactions.json")
		return
	}
	err = json.NewDecoder(transactions_file).Decode(&transactions)
	if err != nil {
		t.Errorf("Error decoding json from test_data/transactions.json")
		return
	}
	defer transactions_file.Close()
	cw_queue_hash, cw_queue_len, cw_addresses, cw_amounts, cw_token_ids, ok := WithdrawalQueueHash(transactions)
	if !ok {
		t.Errorf("Failed to hash withdrawal")
		return
	}
	expected_cw_queue_len := 2
	if cw_queue_len != expected_cw_queue_len {
		t.Errorf("Failed to get withdrawal queue len expected %v got %v", expected_cw_queue_len, cw_queue_len)
		return
	}
	expected_cw_addresses := []string{"0xe9e2d5240237955f5955c28cd9ee9d5f66800cf1", "0xe9e2d5240237955f5955c28cd9ee9d5f66800cf1"}
	if !reflect.DeepEqual(cw_addresses, expected_cw_addresses) {
		t.Errorf("Failed to get withdrawal addresses expected %v got %v", expected_cw_addresses, cw_addresses)
		return
	}
	expected_cw_amounts := []string{"10000", "10000"}
	if !reflect.DeepEqual(cw_amounts, expected_cw_amounts) {
		t.Errorf("Failed to get withdrawal amounts expected %v got %v", expected_cw_amounts, cw_amounts)
		return
	}
	expected_cw_token_ids := []string{"0xEe146Fac7b2fce5FdBE31C36d89cF92f6b006F80", "0xEe146Fac7b2fce5FdBE31C36d89cF92f6b006F80"}
	if !reflect.DeepEqual(cw_token_ids, expected_cw_token_ids) {
		t.Errorf("Failed to get withdrawal token ids expected %v got %v", expected_cw_token_ids, cw_token_ids)
		return
	}
	expected_hash := "992e822161de07540be8c4b2539cc32205bd9d9829fb384b95a99a2d51c6411c"
	if hex.EncodeToString(cw_queue_hash) != expected_hash {
		t.Errorf("Failed to hash hash expected %s got %s", expected_hash, hex.EncodeToString(cw_queue_hash))
		return
	}
}

func TestNftQueueItemHash(t *testing.T) {
	address := "0x447bF33F7c7C925eb7674bCF590AeD4Aa57e656b"
	token := "0xedb6375347e060b055d6af9842ba8c55e3d93e3a"
	token_id := "12"
	l2_minted := false
	hash, ok := NftQueueItemHash(address, token, token_id, l2_minted)
	if !ok {
		t.Errorf("Failed to hash queue item")
		return
	}
	expected_hash := "0f42b41acae7a1c7db41cb1bee05a07b1b0dd708b36c7be3df87d25235ea5eb2"
	if hex.EncodeToString(hash) != expected_hash {
		t.Errorf("Failed to hash hash expected %s got %s", expected_hash, hex.EncodeToString(hash))
		return
	}
}

func TestNftWithdrawalQueueHash(t *testing.T) {
	transactions := make([]Transaction, 0)
	transactions_file, err := os.Open("test_data/transactions.json")
	if err != nil {
		t.Errorf("Error opening test_data/transactions.json")
		return
	}
	err = json.NewDecoder(transactions_file).Decode(&transactions)
	if err != nil {
		t.Errorf("Error decoding json from test_data/transactions.json")
		return
	}
	defer transactions_file.Close()
	cw_queue_hash, cw_queue_len, cw_addresses, cw_amounts, cw_token_ids, l2_minted, ok := NftWithdrawalQueueHash(transactions)
	if !ok {
		t.Errorf("Failed to hash withdrawal")
		return
	}
	expected_cw_queue_len := 2
	if cw_queue_len != expected_cw_queue_len {
		t.Errorf("Failed to get withdrawal queue len expected %v got %v", expected_cw_queue_len, cw_queue_len)
		return
	}
	expected_cw_addresses := []string{"0x46714661eecb6f07065dcb4bf3d9b772dcefa63a", "0x46714661eecb6f07065dcb4bf3d9b772dcefa63a"}
	if !reflect.DeepEqual(cw_addresses, expected_cw_addresses) {
		t.Errorf("Failed to get withdrawal addresses expected %v got %v", expected_cw_addresses, cw_addresses)
		return
	}
	expected_cw_amounts := []string{"14", "14"}
	if !reflect.DeepEqual(cw_amounts, expected_cw_amounts) {
		t.Errorf("Failed to get withdrawal amounts expected %v got %v", expected_cw_amounts, cw_amounts)
		return
	}
	expected_cw_token_ids := []string{"0xedb6375347e060b055d6af9842ba8c55e3d93e3a", "0xedb6375347e060b055d6af9842ba8c55e3d93e3a"}
	if !reflect.DeepEqual(cw_token_ids, expected_cw_token_ids) {
		t.Errorf("Failed to get withdrawal token ids expected %v got %v", expected_cw_token_ids, cw_token_ids)
		return
	}
	expected_l2_minted := []bool{true, true}
	if !reflect.DeepEqual(l2_minted, expected_l2_minted) {
		t.Errorf("Failed to get withdrawal l2 minted expected %v got %v", expected_l2_minted, l2_minted)
		return
	}
	expected_hash := "a8f31375855b3c1dde4082378231a9fd871ae7fca431a06029a429d38fa7e549"
	if hex.EncodeToString(cw_queue_hash) != expected_hash {
		t.Errorf("Failed to hash hash expected %s got %s", expected_hash, hex.EncodeToString(cw_queue_hash))
		return
	}
}

func TestGetOptimizedNonce(t *testing.T) {
	used_lister_nonce := []uint{1, 2, 3, 4, 6, 8, 21}
	expected_optimized_nonce := []uint{0, 4, 6, 8, 21}
	optimized_nonce := GetOptimizedNonce(used_lister_nonce)
	if !reflect.DeepEqual(optimized_nonce, expected_optimized_nonce) {
		t.Errorf("Failed to get optimized nonce expected %v got %v", expected_optimized_nonce, optimized_nonce)
		return
	}
}

func TestProcessAndVerifyCollectionData(t *testing.T) {
	collection_data := map[string]interface{}{
		"Id":              1,
		"Owner":           "0x46714661eecb6f07065dcb4bf3d9b772dcefa63a",
		"ContractAddress": "0xedb6375347e060b055d6af9842ba8c55e3d93e3a",
		"Name":            "New NFT Collection",
		"BaseUri":         "https://ipfs.io/ipfs/QmVLbfDpBj9XxXCCgWwhshpAQE9X23skZ8SfpUPn29HhnQ",
		"MintStart":       "10",
		"MintEnd":         "100",
		"MintUsers": []interface{}{
			"0x46714661eecb6f07065dcb4bf3d9b772dcefa63a",
		},
		"MintFees":             "1000",
		"MintFeesToken":        "0xEe146Fac7b2fce5FdBE31C36d89cF92f6b006F80",
		"RoyaltyFeesPercetage": "10",
		"Signature":            "0x4e55606dd8904ffd61bb8aea4c1d8ab3fcec635dbe6abed6c3bce9a1afafc30914d756433fe1b5e748f1962f04a0f5cbf6e60c345a4c1a2e586eb66154a116061c",
	}
	verfied, hash := ProcessAndVerifyCollectionData(collection_data)
	if !verfied {
		t.Errorf("Failed to verify collection data")
		return
	}
	expected_hash := "314ac2223ebffacbec4f90750995e9f19326ddd9012716687061095372b6c59f"
	if hex.EncodeToString(hash) != expected_hash {
		t.Errorf("Failed to hash hash expected %s got %s", expected_hash, hex.EncodeToString(hash))
		return
	}
}
