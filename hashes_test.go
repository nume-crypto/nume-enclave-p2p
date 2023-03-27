package main

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func TestDigitalSignatureMessageHash(t *testing.T) {
	from := "00"
	to := "00"
	currency := "0"
	amount := "0"
	nonce := "0"
	hashed_message, ok := DigitalSignatureMessageHash(from, to, currency, amount, nonce)
	if !ok {
		t.Errorf("Failed to hash message")
		return
	}
	expected_hash := "0ba788e8a57932d9ba121cdc539a55a8d03541c192b08701fbf3af57681de759"
	if hex.EncodeToString(hashed_message) != expected_hash {
		t.Errorf("Failed to hash message expected %s got %s", expected_hash, hex.EncodeToString(hashed_message))
		return
	}
}

func TestLeafHash(t *testing.T) {
	pub_key := "00"
	balance_root := "00"
	hashed_message, ok := LeafHash(pub_key, balance_root)
	if !ok {
		t.Errorf("Failed to hash message")
		return
	}
	expected_hash := "302927ba94dfa8136f80c1896185578157b5811cf031c06ab9686f5a1d89b94d"
	if hex.EncodeToString(hashed_message) != expected_hash {
		t.Errorf("Failed to hash message expected %s got %s", expected_hash, hex.EncodeToString(hashed_message))
		return
	}
}

func TestG1Hash(t *testing.T) {
	pub_key_g1 := [2]string{
		"22e9eda228ccc6368167df61fc8daffffc08e3b0a573787c236a64699671e000",
		"2c532e2d6cb2c03dd41d61632d2c8d726cb49d08eac94233df96e4f77a1b6c1f",
	}
	hashed_message, ok := G1Hash(pub_key_g1)
	if !ok {
		t.Errorf("Failed to hash message")
		return
	}
	expected_hash := "196d9f92fc71303cd2ac01eaec5dfef3590e526fd19cc6b78b51c1fbb4cb326a"
	if hex.EncodeToString(hashed_message) != expected_hash {
		t.Errorf("Failed to hash message expected %s got %s", expected_hash, hex.EncodeToString(hashed_message))
		return
	}
}

func TestSettlementWithDepositsMessage(t *testing.T) {
	prev_root := "26796d7073f12c5cdf95f5b30b071cbf5fc6e2f69d26e1af048a6b3bdcddc855"
	current_root := "1960a8abb428a13ebf10b6632e14748aa4e38b758edeaf2d831aee00739bb708"
	tree_hash := "9e653ebf9b254458e53e7b073ff881a"
	queue_index := uint(1)
	queue_hash := "0f6bc10b6e3d746d5f040e2b93cbb10e10ec22d4bb5aca47d4f64db86906fd9e"
	block_number := 2
	message := SettlementWithDepositsMessage(prev_root, current_root, tree_hash, queue_index, queue_hash, block_number)
	expected_message := "26796d7073f12c5cdf95f5b30b071cbf5fc6e2f69d26e1af048a6b3bdcddc8551960a8abb428a13ebf10b6632e14748aa4e38b758edeaf2d831aee00739bb7080000000000000000000000000000000009e653ebf9b254458e53e7b073ff881a00000000000000000000000000000000000000000000000000000000000000010f6bc10b6e3d746d5f040e2b93cbb10e10ec22d4bb5aca47d4f64db86906fd9e0000000000000000000000000000000000000000000000000000000000000002"
	if message != expected_message {
		t.Errorf("Failed to hash message expected %s got %s", expected_message, message)
		return
	}
}

func TestSettlementSignedByAllUsersMessage(t *testing.T) {
	prev_root := "26796d7073f12c5cdf95f5b30b071cbf5fc6e2f69d26e1af048a6b3bdcddc855"
	current_root := "1960a8abb428a13ebf10b6632e14748aa4e38b758edeaf2d831aee00739bb708"
	tree_hash := "9e653ebf9b254458e53e7b073ff881a"
	block_number := 2
	message := SettlementSignedByAllUsersMessage(prev_root, current_root, tree_hash, block_number)
	expected_message := "26796d7073f12c5cdf95f5b30b071cbf5fc6e2f69d26e1af048a6b3bdcddc8551960a8abb428a13ebf10b6632e14748aa4e38b758edeaf2d831aee00739bb7080000000000000000000000000000000009e653ebf9b254458e53e7b073ff881a0000000000000000000000000000000000000000000000000000000000000002"
	if message != expected_message {
		t.Errorf("Failed to hash message expected %s got %s", expected_message, message)
		return
	}
}

func TestSettlementWithDepositsAndWithdrawalsMessage(t *testing.T) {
	prev_root := "26796d7073f12c5cdf95f5b30b071cbf5fc6e2f69d26e1af048a6b3bdcddc855"
	current_root := "1960a8abb428a13ebf10b6632e14748aa4e38b758edeaf2d831aee00739bb708"
	tree_hash := "9e653ebf9b254458e53e7b073ff881a"
	queue_index := uint(1)
	queue_hash := "0f6bc10b6e3d746d5f040e2b93cbb10e10ec22d4bb5aca47d4f64db86906fd9e"
	withdrawal_hash := "0f6bc10b6e3d746d5f040e2b93cbb10e10ec22d4bb5aca47d4f64db86906fd91"
	block_number := 2
	message := SettlementWithDepositsAndWithdrawalsMessage(prev_root, current_root, tree_hash, queue_index, queue_hash, withdrawal_hash, block_number)
	expected_message := "26796d7073f12c5cdf95f5b30b071cbf5fc6e2f69d26e1af048a6b3bdcddc8551960a8abb428a13ebf10b6632e14748aa4e38b758edeaf2d831aee00739bb7080000000000000000000000000000000009e653ebf9b254458e53e7b073ff881a00000000000000000000000000000000000000000000000000000000000000010f6bc10b6e3d746d5f040e2b93cbb10e10ec22d4bb5aca47d4f64db86906fd9e0f6bc10b6e3d746d5f040e2b93cbb10e10ec22d4bb5aca47d4f64db86906fd910000000000000000000000000000000000000000000000000000000000000002"
	if message != expected_message {
		t.Errorf("Failed to hash message expected %s got %s", expected_message, message)
		return
	}
}

func TestSettlementWithWithdrawalsMessage(t *testing.T) {
	prev_root := "26796d7073f12c5cdf95f5b30b071cbf5fc6e2f69d26e1af048a6b3bdcddc855"
	current_root := "1960a8abb428a13ebf10b6632e14748aa4e38b758edeaf2d831aee00739bb708"
	tree_hash := "9e653ebf9b254458e53e7b073ff881a"
	withdrawal_hash := "0f6bc10b6e3d746d5f040e2b93cbb10e10ec22d4bb5aca47d4f64db86906fd9e"
	block_number := 2
	message := SettlementWithWithdrawalsMessage(prev_root, current_root, tree_hash, withdrawal_hash, block_number)
	expected_message := "26796d7073f12c5cdf95f5b30b071cbf5fc6e2f69d26e1af048a6b3bdcddc8551960a8abb428a13ebf10b6632e14748aa4e38b758edeaf2d831aee00739bb7080000000000000000000000000000000009e653ebf9b254458e53e7b073ff881a0f6bc10b6e3d746d5f040e2b93cbb10e10ec22d4bb5aca47d4f64db86906fd9e0000000000000000000000000000000000000000000000000000000000000002"
	if message != expected_message {
		t.Errorf("Failed to hash message expected %s got %s", expected_message, message)
		return
	}
}

func TestQueueItemHash(t *testing.T) {
	pub_key := "26796d7073f12c5cdf95f5b30b071cbf5fc6e2f69d26e1af048a6b3bdcddc855"
	token_id := 1
	amount := "99"
	hash, ok := QueueItemHash(pub_key, uint(token_id), amount)
	if !ok {
		t.Errorf("Failed to hash queue item")
		return
	}
	expected_hash := "1bbbc4c8724008a8f4a904a25812f7c8f5443071a170dc5224becb84f0faaa4d"
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
	hash, ok := QueueHash(transactions, 30)
	if !ok {
		t.Errorf("Failed to hash queue")
		return
	}
	expected_hash := "2bab89cfc123b0d6b63760e7375f15785a2bad8826c8c6343fb1fa4231ed071d"
	if hex.EncodeToString(hash) != expected_hash {
		t.Errorf("Failed to hash hash expected %s got %s", expected_hash, hex.EncodeToString(hash))
		return
	}
}

func TestWithdrawalItemHash(t *testing.T) {
	pub_key := "26796d7073f12c5cdf95f5b30b071cbf5fc6e2f69d26e1af048a6b3bdcddc855"
	token_id := 1
	amount := "99"
	hash, ok := WithdrawalItemHash(pub_key, uint(token_id), amount)
	if !ok {
		t.Errorf("Failed to hash withdrawal item")
		return
	}
	expected_hash := "24c527e892f7d82bbe520c19db281ada68a4af22d82a7adfe43274e4401a5c38"
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
	hash, ok := WithdrawalHash(transactions)
	if !ok {
		t.Errorf("Failed to hash withdrawal")
		return
	}
	expected_hash := "097f3b54615a19171b4c89b6ec46ae6f610538dad1d5bfb6c860ddbd35e51260"
	if hex.EncodeToString(hash) != expected_hash {
		t.Errorf("Failed to hash hash expected %s got %s", expected_hash, hex.EncodeToString(hash))
		return
	}
}
