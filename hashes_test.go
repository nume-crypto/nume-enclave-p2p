package main

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestDigitalSignatureMessage(t *testing.T) {
	from := "190ce0ac817bf41f26c665f414e0fc1c955864aff93b28132b2f73ed65522a29"
	to := "196d9f92fc71303cd2ac01eaec5dfef3590e526fd19cc6b78b51c1fbb4cb326a"
	currency := "0"
	amount := "1880000000000000000"
	nonce := 1
	bn := 1
	hashed_message := DigitalSignatureMessage(from, to, currency, amount, uint64(nonce), int64(bn))
	expected_hash := "190ce0ac817bf41f26c665f414e0fc1c955864aff93b28132b2f73ed65522a29196d9f92fc71303cd2ac01eaec5dfef3590e526fd19cc6b78b51c1fbb4cb326a00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001a171a0a11bc000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001"
	if hashed_message != expected_hash {
		t.Errorf("Failed to hash message expected %s got %s", expected_hash, hashed_message)
		return
	}
}

func TestQueueItemHash(t *testing.T) {
	pub_key := "26796d7073f12c5cdf95f5b30b071cbf5fc6e2f69d26e1af048a6b3bdcddc855"
	token := "token_id"
	amount := "99"
	hash, ok := QueueItemHash(pub_key, token, amount)
	if !ok {
		t.Errorf("Failed to hash queue item")
		return
	}
	expected_hash := "219730f65f0e4f1747ee765930f120dee7b792a3e7bd8371ad7e10ad46850c2a"
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
	hash, queue_index, ok := QueueHash(transactions)
	if !ok {
		t.Errorf("Failed to hash queue")
		return
	}
	if queue_index != 5 {
		t.Errorf("Failed to get queue index expected 2 got %d", queue_index)
		return
	}
	expected_hash := "26712cdb7e9e91dafb6abdf438fa70718c82be5a8c0e16d0306c2c5c1008f8e6"
	if hex.EncodeToString(hash) != expected_hash {
		t.Errorf("Failed to hash hash expected %s got %s", expected_hash, hex.EncodeToString(hash))
		return
	}
}

func TestWithdrawalItemHash(t *testing.T) {
	address := "26796d7073f12c5cdf95f5b30b071cbf5fc6e2f69d26e1af048a6b3bdcddc855"
	token_id := 1
	amount := "99"
	hash, ok := WithdrawalItemHash(amount, uint(token_id), address)
	if !ok {
		t.Errorf("Failed to hash withdrawal item")
		return
	}
	expected_hash := "269e489766dad8078c78461af75cc3bbf2087a2649d1f0b71e012b58f3ab6ab3"
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
	withdrawal_hash, withdrawal_amounts, withdrawal_addresses, withdrawal_tokens, ok := WithdrawalHash(transactions)
	if !ok {
		t.Errorf("Failed to hash withdrawal")
		return
	}
	expected_withdrawal_amounts := []string{"180000000000000000", "230000", "60000", "270000000000000032", "460000"}
	if !reflect.DeepEqual(withdrawal_amounts, expected_withdrawal_amounts) {
		t.Errorf("Failed to get withdrawal amounts expected %v got %v", expected_withdrawal_amounts, withdrawal_amounts)
		return
	}
	expected_withdrawal_addresses := []string{"447bF33F7c7C925eb7674bCF590AeD4Aa57e656b", "447bF33F7c7C925eb7674bCF590AeD4Aa57e656b", "447bF33F7c7C925eb7674bCF590AeD4Aa57e656b", "447bF33F7c7C925eb7674bCF590AeD4Aa57e656b", "447bF33F7c7C925eb7674bCF590AeD4Aa57e656b"}
	if !reflect.DeepEqual(withdrawal_addresses, expected_withdrawal_addresses) {
		t.Errorf("Failed to get withdrawal addresses expected %v got %v", expected_withdrawal_addresses, withdrawal_addresses)
		return
	}
	expected_withdrawal_tokens := []uint{1, 3, 3, 1, 2}
	if !reflect.DeepEqual(withdrawal_tokens, expected_withdrawal_tokens) {
		t.Errorf("Failed to get withdrawal tokens expected %v got %v", expected_withdrawal_tokens, withdrawal_tokens)
		return
	}
	expected_hash := "19ec7dd4d023e9cee21d0b8bbf9b97e2b6774e5cf2d00ee12bdb9d7ee564d880"
	if hex.EncodeToString(withdrawal_hash) != expected_hash {
		t.Errorf("Failed to hash hash expected %s got %s", expected_hash, hex.EncodeToString(withdrawal_hash))
		return
	}
}

func TestWithdrawalQueueItemHash(t *testing.T) {
	from := "190ce0ac817bf41f26c665f414e0fc1c955864aff93b28132b2f73ed65522a29"
	to := "447bF33F7c7C925eb7674bCF590AeD4Aa57e656b"
	token := "1"
	amt := "1880000000000000000"
	hash, ok := WithdrawalQueueItemHash(from, to, token, amt)
	if !ok {
		t.Errorf("Failed to hash withdrawal queue item")
		return
	}
	expected_hash := "0e73e13f2581b6e50190f469f8db5d22f60ef3f7bd4f301cb3b34f0a11a6668e"
	if hex.EncodeToString(hash) != expected_hash {
		t.Errorf("Failed to hash hash expected %s got %s", expected_hash, hex.EncodeToString(hash))
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
	cw_queue_hash, cw_queue_len, cw_addresses, cw_amounts, cw_token_ids, cw_bls_keys, ok := WithdrawalQueueHash(transactions)
	if !ok {
		t.Errorf("Failed to hash withdrawal")
		return
	}
	expected_cw_queue_len := 1
	if cw_queue_len != expected_cw_queue_len {
		t.Errorf("Failed to get withdrawal queue len expected %v got %v", expected_cw_queue_len, cw_queue_len)
		return
	}
	expected_cw_addresses := []string{"447bF33F7c7C925eb7674bCF590AeD4Aa57e656b"}
	if !reflect.DeepEqual(cw_addresses, expected_cw_addresses) {
		t.Errorf("Failed to get withdrawal addresses expected %v got %v", expected_cw_addresses, cw_addresses)
		return
	}
	expected_cw_amounts := []string{"9140000"}
	if !reflect.DeepEqual(cw_amounts, expected_cw_amounts) {
		t.Errorf("Failed to get withdrawal amounts expected %v got %v", expected_cw_amounts, cw_amounts)
		return
	}
	expected_cw_token_ids := []uint{2}
	if !reflect.DeepEqual(cw_token_ids, expected_cw_token_ids) {
		t.Errorf("Failed to get withdrawal token ids expected %v got %v", expected_cw_token_ids, cw_token_ids)
		return
	}
	expected_cw_bls_keys := []string{"282e55a1f3608a3f92eca0f582aa10b02da7d79d47a5d3101095c1583bc328ad"}
	if !reflect.DeepEqual(cw_bls_keys, expected_cw_bls_keys) {
		t.Errorf("Failed to get withdrawal bls keys expected %v got %v", expected_cw_bls_keys, cw_bls_keys)
		return
	}
	expected_hash := "071ed934b08b377203ad38838bb62023631f135b8e4cef189d2a3e660bd0e34a"
	if hex.EncodeToString(cw_queue_hash) != expected_hash {
		t.Errorf("Failed to hash hash expected %s got %s", expected_hash, hex.EncodeToString(cw_queue_hash))
		return
	}

}
