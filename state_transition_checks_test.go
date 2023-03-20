package main

import "testing"

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
