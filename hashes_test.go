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
	hash, queue_index, ok := QueueHash(transactions)
	if !ok {
		t.Errorf("Failed to hash queue")
		return
	}
	if queue_index != 3 {
		t.Errorf("Failed to get queue index expected 3 got %d", queue_index)
		return
	}
	expected_hash := "1a9e0398f7c42398c98ddfc4a1a680e3d879823f91c04c32007a21a1c37f3d2c"
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
	withdrawal_hash, withdrawal_amounts, withdrawal_l2_minted, withdrawal_addresses, withdrawal_tokens, ok := WithdrawalHash(transactions)
	_ = withdrawal_l2_minted
	if !ok {
		t.Errorf("Failed to hash withdrawal")
		return
	}
	expected_withdrawal_amounts := []string{"64000", "123", "12124"}
	if !reflect.DeepEqual(withdrawal_amounts, expected_withdrawal_amounts) {
		t.Errorf("Failed to get withdrawal amounts expected %v got %v", expected_withdrawal_amounts, withdrawal_amounts)
		return
	}
	expected_withdrawal_addresses := []string{"0x5D9aE648d74dF50746951Cbb2FB296c836A7A0e9", "0xD28c7E81b6E8ba287307c034133EB7d72302eBf3", "0x0132b7813B8D84C6f95253B056329151eBF42E76"}
	if !reflect.DeepEqual(withdrawal_addresses, expected_withdrawal_addresses) {
		t.Errorf("Failed to get withdrawal addresses expected %v got %v", expected_withdrawal_addresses, withdrawal_addresses)
		return
	}
	expected_withdrawal_tokens := []string{"0xCE47C48fDF8c9355FDbE4DacC1e1954914D65Be6", "0x799c6832d187243f3367902079A72fb3Fd61cdF7", "0xE9573B8A0AF951431bcBD194E8cc3AeE654Cd723"}
	if !reflect.DeepEqual(withdrawal_tokens, expected_withdrawal_tokens) {
		t.Errorf("Failed to get withdrawal tokens expected %v got %v", expected_withdrawal_tokens, withdrawal_tokens)
		return
	}
	expected_hash := "323ad72842bb3ddae1baa8dd308ef0a686dd4fa5670cf7ecabc12291bbdade7d"
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
	expected_cw_queue_len := 4
	if cw_queue_len != expected_cw_queue_len {
		t.Errorf("Failed to get withdrawal queue len expected %v got %v", expected_cw_queue_len, cw_queue_len)
		return
	}
	expected_cw_addresses := []string{"0x0132b7813B8D84C6f95253B056329151eBF42E76", "0x0132b7813B8D84C6f95253B056329151eBF42E76", "0x2e3925Ff5246f5Cc131860311b249C3712a9D789", "0x0132b7813B8D84C6f95253B056329151eBF42E76"}
	if !reflect.DeepEqual(cw_addresses, expected_cw_addresses) {
		t.Errorf("Failed to get withdrawal addresses expected %v got %v", expected_cw_addresses, cw_addresses)
		return
	}
	expected_cw_amounts := []string{"90", "889127589", "11", "12124"}
	if !reflect.DeepEqual(cw_amounts, expected_cw_amounts) {
		t.Errorf("Failed to get withdrawal amounts expected %v got %v", expected_cw_amounts, cw_amounts)
		return
	}
	expected_cw_token_ids := []string{"0xCE47C48fDF8c9355FDbE4DacC1e1954914D65Be6", "0x799c6832d187243f3367902079A72fb3Fd61cdF7", "0x799c6832d187243f3367902079A72fb3Fd61cdF7", "0xE9573B8A0AF951431bcBD194E8cc3AeE654Cd723"}
	if !reflect.DeepEqual(cw_token_ids, expected_cw_token_ids) {
		t.Errorf("Failed to get withdrawal token ids expected %v got %v", expected_cw_token_ids, cw_token_ids)
		return
	}
	expected_hash := "96100de070a91868ed66757b1e21171ad19b6aeae809cc6c76c7f3b367a98ae3"
	if hex.EncodeToString(cw_queue_hash) != expected_hash {
		t.Errorf("Failed to hash hash expected %s got %s", expected_hash, hex.EncodeToString(cw_queue_hash))
		return
	}

}
