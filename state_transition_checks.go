package main

func CheckNonce(last_nonce, current_nonce uint64) bool {
	return last_nonce < current_nonce
}
