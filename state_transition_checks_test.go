package main

import (
	"encoding/json"
	"os"
	"testing"
)

type CheckNonceData struct {
	last_nonce    uint64
	current_nonce uint64
	result        bool
}

func TestCheckNonce(t *testing.T) {
	check_nonce_data := []CheckNonceData{
		{0, 1, true},
		{1, 1, false},
		{1, 0, false},
	}

	for _, data := range check_nonce_data {
		result := CheckNonce(data.last_nonce, data.current_nonce)
		if result != data.result {
			t.Errorf("CheckNonce(%d, %d) = %t, want %t", data.last_nonce, data.current_nonce, result, data.result)
		}
	}

}

func TestTransitionState(t *testing.T) {
	prev_balances := make(map[string]map[uint]string)
	prev_balances_file, err := os.Open("test_data/prev_balances.json")
	if err != nil {
		t.Errorf("Error opening test_data/prev_balances.json")
		return
	}
	err = json.NewDecoder(prev_balances_file).Decode(&prev_balances)
	if err != nil {
		t.Errorf("Error decoding json from test_data/prev_balances.json")
		return
	}
	defer prev_balances_file.Close()

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

	new_balances_desired := make(map[string]map[uint]string)
	new_balances_file, err := os.Open("test_data/new_balances.json")
	if err != nil {
		t.Errorf("Error opening test_data/new_balances.json")
		return
	}
	err = json.NewDecoder(new_balances_file).Decode(&new_balances_desired)
	if err != nil {
		t.Errorf("Error decoding json from test_data/new_balances.json")
		return
	}
	defer new_balances_file.Close()

	user_keys := make(map[string]UserKeys)
	user_keys_file, err := os.Open("test_data/user_keys.json")
	if err != nil {
		t.Errorf("Error opening test_data/user_keys.json")
		return
	}
	err = json.NewDecoder(user_keys_file).Decode(&user_keys)
	if err != nil {
		t.Errorf("Error decoding json from test_data/user_keys.json")
		return
	}
	defer new_balances_file.Close()

	block_number := int64(1)
	new_balances, settlement_type, _, err := TransitionState(prev_balances, transactions, user_keys, block_number)
	if err != nil {
		t.Errorf("Error in TransitionState " + err.Error())
		return
	}
	if settlement_type != 3 {
		t.Errorf("settlement_type = %d, want %d", settlement_type, 3)
		return
	}

	result := NestedMapsEqual(new_balances, new_balances_desired)
	if !result {
		t.Errorf("NestedMapsEqual(new_balances, new_balances_desired) = %t, want %t", result, true)
		return
	}

}
