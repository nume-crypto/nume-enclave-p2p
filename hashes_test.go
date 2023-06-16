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
	if queue_index != 16 {
		t.Errorf("Failed to get queue index expected 16 got %d", queue_index)
		return
	}
	expected_hash := "40b7f56cbc8eb72c02be3981f983264735adece7bc5f12092796055b83e16e25"
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
	expected_withdrawal_amt_or_token_id := []string{"1420000000000000000", "1446", "917", "130000", "1660000000000000000", "2892", "11"}
	if !reflect.DeepEqual(withdrawal_amt_or_token_id, expected_withdrawal_amt_or_token_id) {
		t.Errorf("Failed to get withdrawal amounts expected %v got %v", expected_withdrawal_amt_or_token_id, withdrawal_amt_or_token_id)
		return
	}
	expected_withdrawal_addresses := []string{"0xCcFf350Ef46B85228d6650a802107e58BF6A32Ab", "0x1b34B2f706cDA183E4818D2ceaF58253CcAb3428", "0xa9B39CB5Ebf5DEB0818561e8bc64092FBde34613", "0x46714661eECB6F07065DCb4BF3D9b772dcefa63a", "0xc8902116932225CC2e5D21d090574A54C10DC21b", "0x995227bD4dBFCd247Fd7c97eDbA86C4Ad46bfB05", "0xDc42d1dd82217013B79EBA43673912C4a3fC7bEA"}
	if !reflect.DeepEqual(withdrawal_addresses, expected_withdrawal_addresses) {
		t.Errorf("Failed to get withdrawal addresses expected %v got %v", expected_withdrawal_addresses, withdrawal_addresses)
		return
	}
	expected_withdrawal_currency_or_nft := []string{"0xEe146Fac7b2fce5FdBE31C36d89cF92f6b006F80", "0x799c6832d187243f3367902079A72fb3Fd61cdF7", "0x799c6832d187243f3367902079A72fb3Fd61cdF7", "0xCE47C48fDF8c9355FDbE4DacC1e1954914D65Be6", "0xEe146Fac7b2fce5FdBE31C36d89cF92f6b006F80", "0x799c6832d187243f3367902079A72fb3Fd61cdF7", "0xedb6375347e060b055d6af9842ba8c55e3d93e3a"}
	if !reflect.DeepEqual(withdrawal_currency_or_nft, expected_withdrawal_currency_or_nft) {
		t.Errorf("Failed to get withdrawal tokens expected %v got %v", expected_withdrawal_currency_or_nft, withdrawal_currency_or_nft)
		return
	}
	expected_hash := "ce7b81ed1eaa3eef6d5169b6f50c59c1f06253d524c478adc7316262dd5c3024"
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
	expected_cw_token_ids := []string{"0xCE47C48fDF8c9355FDbE4DacC1e1954914D65Be6", "0xCE47C48fDF8c9355FDbE4DacC1e1954914D65Be6"}
	if !reflect.DeepEqual(cw_token_ids, expected_cw_token_ids) {
		t.Errorf("Failed to get withdrawal token ids expected %v got %v", expected_cw_token_ids, cw_token_ids)
		return
	}
	expected_hash := "d2946539299563a92641f5c12ba368c4228b67b2fb5e28b12ca4e0e240b28535"
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
	expected_cw_amounts := []string{"13", "13"}
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
	expected_hash := "ff3bb77ab1e10313530a123fa7fd57047c34b7656ccef47b55896456e7e86766"
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
