package main

import (
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

// func TestTransitionState(t *testing.T) {
// 	prev_balances := make(map[string]map[string]string)
// 	prev_balances_file, err := os.Open("test_data/prev_balances.json")
// 	if err != nil {
// 		t.Errorf("Error opening test_data/prev_balances.json")
// 		return
// 	}
// 	err = json.NewDecoder(prev_balances_file).Decode(&prev_balances)
// 	if err != nil {
// 		t.Errorf("Error decoding json from test_data/prev_balances.json")
// 		return
// 	}
// 	defer prev_balances_file.Close()

// 	transactions := make([]NftTransaction, 0)
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

// 	new_balances_desired := make(map[string]map[string]string)
// 	new_balances_file, err := os.Open("test_data/new_balances.json")
// 	if err != nil {
// 		t.Errorf("Error opening test_data/new_balances.json")
// 		return
// 	}
// 	err = json.NewDecoder(new_balances_file).Decode(&new_balances_desired)
// 	if err != nil {
// 		t.Errorf("Error decoding json from test_data/new_balances.json")
// 		return
// 	}
// 	defer new_balances_file.Close()

// 	meta_data := make(map[string]interface{})
// 	meta_data_file, err := os.Open("test_data/meta_data.json")
// 	if err != nil {
// 		t.Errorf("Error opening test_data/meta_data.json")
// 		return
// 	}
// 	err = json.NewDecoder(meta_data_file).Decode(&meta_data)
// 	if err != nil {
// 		t.Errorf("Error decoding json from test_data/meta_data.json")
// 		return
// 	}
// 	defer meta_data_file.Close()
// 	currencies := []string{}
// 	for _, currency := range meta_data["currencies"].([]interface{}) {
// 		currencies = append(currencies, currency.(string))
// 	}
// 	new_balances, settlement_type, _, _, err := TransitionState(prev_balances, transactions, currencies)
// 	if err != nil {
// 		t.Errorf("Error in TransitionState " + err.Error())
// 		return
// 	}
// 	if settlement_type != 7 {
// 		t.Errorf("settlement_type = %d, want %d", settlement_type, 4)
// 		return
// 	}

// 	result := NestedMapsEqual(new_balances, new_balances_desired)
// 	if !result {
// 		t.Errorf("NestedMapsEqual(new_balances, new_balances_desired) = %t, want %t", result, true)
// 		return
// 	}

// }
