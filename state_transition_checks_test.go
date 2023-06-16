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

func TestTransitionState(t *testing.T) {
	input_data, _, err := GetData("./test_data")
	if err != nil {
		t.Errorf("Error in GetData " + err.Error())
		return
	}
	currencies := []string{}
	for _, c := range input_data.MetaData["currencies"].([]interface{}) {
		currencies = append(currencies, c.(string))
	}
	new_balances, _, _, _, err := TransitionState(input_data.OldUserBalances, input_data.Transactions, currencies, append(input_data.OldNftCollections, input_data.NewNftCollections...), input_data.UserListerNonce, input_data.MetaData)
	if err != nil {
		t.Errorf("Error in TransitionState " + err.Error())
		return
	}
	for _, v := range input_data.MetaData["users_ordered"].([]interface{}) {
		if _, ok := new_balances[v.(string)]; !ok {
			new_balances[v.(string)] = make(map[string]string)
			new_balances[v.(string)][input_data.MetaData["fee_currency_token"].(string)] = "0"
		} else {
			if _, ok := new_balances[v.(string)][input_data.MetaData["fee_currency_token"].(string)]; !ok {
				new_balances[v.(string)][input_data.MetaData["fee_currency_token"].(string)] = "0"
			}
		}
	}
	result := NestedMapsEqual(new_balances, input_data.NewUserBalances)
	if !result {
		t.Errorf("NestedMapsEqual(new_balances, new_balances_desired) = %t, want %t", result, true)
		PrettyPrint("new_balances", new_balances)
		PrettyPrint("new_balances_desired", input_data.NewUserBalances)
		return
	}

}
