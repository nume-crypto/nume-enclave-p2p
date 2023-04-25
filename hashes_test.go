package main

import (
	"encoding/hex"
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

// func TestQueueHash(t *testing.T) {
// 	transactions := make([]Transaction, 0)
// 	transactions_file, err := os.Open("test_data/transactions.json")
// 	if err != nil {
// 		t.Errorf("Error opening test_data/transactions.json")
// 		return
// 	}
// 	err = json.NewDecoder(transactions_file).Decode(&transactions)
// 	if err != nil {
// 		t.Errorf("Error decoding json from test_data/transactions.json")
// 		return
// 	}
// 	defer transactions_file.Close()
// 	hash, queue_index, ok := QueueHash(transactions)
// 	if !ok {
// 		t.Errorf("Failed to hash queue")
// 		return
// 	}
// 	if queue_index != 5 {
// 		t.Errorf("Failed to get queue index expected 2 got %d", queue_index)
// 		return
// 	}
// 	expected_hash := "26712cdb7e9e91dafb6abdf438fa70718c82be5a8c0e16d0306c2c5c1008f8e6"
// 	if hex.EncodeToString(hash) != expected_hash {
// 		t.Errorf("Failed to hash hash expected %s got %s", expected_hash, hex.EncodeToString(hash))
// 		return
// 	}
// }

// func TestWithdrawalHash(t *testing.T) {
// 	transactions := make([]Transaction, 0)
// 	transactions_file, err := os.Open("test_data/transactions.json")
// 	if err != nil {
// 		t.Errorf("Error opening test_data/transactions.json")
// 		return
// 	}
// 	err = json.NewDecoder(transactions_file).Decode(&transactions)
// 	if err != nil {
// 		t.Errorf("Error decoding json from test_data/transactions.json")
// 		return
// 	}
// 	defer transactions_file.Close()
// 	withdrawal_hash, withdrawal_amounts, withdrawal_addresses, withdrawal_tokens, ok := WithdrawalHash(transactions)
// 	if !ok {
// 		t.Errorf("Failed to hash withdrawal")
// 		return
// 	}
// 	expected_withdrawal_amounts := []string{"180000000000000000", "230000", "60000", "270000000000000032", "460000"}
// 	if !reflect.DeepEqual(withdrawal_amounts, expected_withdrawal_amounts) {
// 		t.Errorf("Failed to get withdrawal amounts expected %v got %v", expected_withdrawal_amounts, withdrawal_amounts)
// 		return
// 	}
// 	expected_withdrawal_addresses := []string{"447bF33F7c7C925eb7674bCF590AeD4Aa57e656b", "447bF33F7c7C925eb7674bCF590AeD4Aa57e656b", "447bF33F7c7C925eb7674bCF590AeD4Aa57e656b", "447bF33F7c7C925eb7674bCF590AeD4Aa57e656b", "447bF33F7c7C925eb7674bCF590AeD4Aa57e656b"}
// 	if !reflect.DeepEqual(withdrawal_addresses, expected_withdrawal_addresses) {
// 		t.Errorf("Failed to get withdrawal addresses expected %v got %v", expected_withdrawal_addresses, withdrawal_addresses)
// 		return
// 	}
// 	expected_withdrawal_tokens := []uint{1, 3, 3, 1, 2}
// 	if !reflect.DeepEqual(withdrawal_tokens, expected_withdrawal_tokens) {
// 		t.Errorf("Failed to get withdrawal tokens expected %v got %v", expected_withdrawal_tokens, withdrawal_tokens)
// 		return
// 	}
// 	expected_hash := "19ec7dd4d023e9cee21d0b8bbf9b97e2b6774e5cf2d00ee12bdb9d7ee564d880"
// 	if hex.EncodeToString(withdrawal_hash) != expected_hash {
// 		t.Errorf("Failed to hash hash expected %s got %s", expected_hash, hex.EncodeToString(withdrawal_hash))
// 		return
// 	}
// }

// func TestWithdrawalQueueHash(t *testing.T) {
// 	transactions := make([]Transaction, 0)
// 	transactions_file, err := os.Open("test_data/transactions.json")
// 	if err != nil {
// 		t.Errorf("Error opening test_data/transactions.json")
// 		return
// 	}
// 	err = json.NewDecoder(transactions_file).Decode(&transactions)
// 	if err != nil {
// 		t.Errorf("Error decoding json from test_data/transactions.json")
// 		return
// 	}
// 	defer transactions_file.Close()
// 	cw_queue_hash, cw_queue_len, cw_addresses, cw_amounts, cw_token_ids, ok := WithdrawalQueueHash(transactions)
// 	if !ok {
// 		t.Errorf("Failed to hash withdrawal")
// 		return
// 	}
// 	expected_cw_queue_len := 1
// 	if cw_queue_len != expected_cw_queue_len {
// 		t.Errorf("Failed to get withdrawal queue len expected %v got %v", expected_cw_queue_len, cw_queue_len)
// 		return
// 	}
// 	expected_cw_addresses := []string{"447bF33F7c7C925eb7674bCF590AeD4Aa57e656b"}
// 	if !reflect.DeepEqual(cw_addresses, expected_cw_addresses) {
// 		t.Errorf("Failed to get withdrawal addresses expected %v got %v", expected_cw_addresses, cw_addresses)
// 		return
// 	}
// 	expected_cw_amounts := []string{"9140000"}
// 	if !reflect.DeepEqual(cw_amounts, expected_cw_amounts) {
// 		t.Errorf("Failed to get withdrawal amounts expected %v got %v", expected_cw_amounts, cw_amounts)
// 		return
// 	}
// 	expected_cw_token_ids := []uint{2}
// 	if !reflect.DeepEqual(cw_token_ids, expected_cw_token_ids) {
// 		t.Errorf("Failed to get withdrawal token ids expected %v got %v", expected_cw_token_ids, cw_token_ids)
// 		return
// 	}
// 	expected_hash := "071ed934b08b377203ad38838bb62023631f135b8e4cef189d2a3e660bd0e34a"
// 	if hex.EncodeToString(cw_queue_hash) != expected_hash {
// 		t.Errorf("Failed to hash hash expected %s got %s", expected_hash, hex.EncodeToString(cw_queue_hash))
// 		return
// 	}

// }
