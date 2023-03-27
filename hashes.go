package main

import (
	"fmt"
	"math/big"
)

func DigitalSignatureMessageHash(from string, to string, currency string, amount string, nonce string) ([]byte, bool) {
	hFunc := NewMiMC()
	var hashed_message []byte
	cb, ok := new(big.Int).SetString(from, 16)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb, ok = new(big.Int).SetString(to, 16)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb, ok = new(big.Int).SetString(currency, 10)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb, ok = new(big.Int).SetString(amount, 10)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb, ok = new(big.Int).SetString(nonce, 10)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hashed_message = hFunc.Sum(nil)
	hFunc.Reset()
	return hashed_message, true
}

func LeafHash(pub_key string, balance_root string) ([]byte, bool) {
	hFunc := NewMiMC()
	var hashed_message []byte
	cb, ok := new(big.Int).SetString(pub_key, 16)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb, ok = new(big.Int).SetString(balance_root, 16)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hashed_message = hFunc.Sum(nil)
	hFunc.Reset()
	return hashed_message, true
}

func G1Hash(g1_keys [2]string) ([]byte, bool) {
	hFunc := NewMiMC()
	var hashed_message []byte
	cb, ok := new(big.Int).SetString(g1_keys[0], 16)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb, ok = new(big.Int).SetString(g1_keys[1], 16)
	if !ok {
		return hashed_message, false
	}
	hFunc.Write(cb.Bytes())
	hashed_message = hFunc.Sum(nil)
	hFunc.Reset()
	return hashed_message, true
}

func SettlementWithDepositsMessage(prev_root, new_root, tree_hash string, queue_index uint, queue_hash string, block_number int) string {
	message := prev_root + new_root + fmt.Sprintf("%064s", tree_hash) + fmt.Sprintf("%064x", queue_index) + fmt.Sprintf("%064s", queue_hash) + fmt.Sprintf("%064x", block_number)
	return message
}

func SettlementSignedByAllUsersMessage(prev_root, new_root, tree_hash string, block_number int) string {
	message := prev_root + new_root + fmt.Sprintf("%064s", tree_hash) + fmt.Sprintf("%064x", block_number)
	return message
}

func SettlementWithDepositsAndWithdrawalsMessage(prev_root, new_root, tree_hash string, queue_index uint, queue_hash string, withdrawal_hash string, block_number int) string {
	message := prev_root + new_root + fmt.Sprintf("%064s", tree_hash) + fmt.Sprintf("%064x", queue_index) + fmt.Sprintf("%064s", queue_hash) + fmt.Sprintf("%064s", withdrawal_hash) + fmt.Sprintf("%064x", block_number)
	return message
}

func SettlementWithWithdrawalsMessage(prev_root, new_root, tree_hash string, withdrawal_hash string, block_number int) string {
	message := prev_root + new_root + fmt.Sprintf("%064s", tree_hash) + fmt.Sprintf("%064s", withdrawal_hash) + fmt.Sprintf("%064x", block_number)
	return message
}

func QueueItemHash(pub_key string, token_id uint, amount string) ([]byte, bool) {
	var queue_hash []byte
	hFunc := NewMiMC()
	cb, ok := new(big.Int).SetString(pub_key, 16)
	if !ok {
		return queue_hash, ok
	}

	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb = new(big.Int).SetUint64(uint64(token_id))
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb, ok = new(big.Int).SetString(amount, 16)
	if !ok {
		return queue_hash, ok
	}
	hFunc.Write(cb.Bytes())
	queue_hash = hFunc.Sum(nil)

	hFunc.Reset()
	return queue_hash, true
}

func QueueHash(queue []Transaction) ([]byte, int, bool) {
	var queue_hash []byte
	hFunc := NewMiMC()
	var valid_queue [][]byte
	for i := 0; i < len(queue); i++ {
		if queue[i].Type == "deposit" {
			cb, ok := QueueItemHash(queue[i].From, queue[i].CurrencyTokenOrder, queue[i].Amount)
			if !ok {
				return queue_hash, 0, ok
			}
			valid_queue = append(valid_queue, cb)
		}
	}
	for _, item := range valid_queue {
		hFunc.Write(item)
		queue_hash = hFunc.Sum(nil)
	}
	hFunc.Reset()
	return queue_hash, len(valid_queue), true
}

func WithdrawalItemHash(amount string, token_id uint, address string) ([]byte, bool) {
	var withdrawal_hash []byte
	hFunc := NewMiMC()
	cb := new(big.Int).SetUint64(uint64(token_id))
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb, ok := new(big.Int).SetString(amount, 16)
	if !ok {
		return withdrawal_hash, ok
	}
	hFunc.Write(cb.Bytes())
	hFunc.Sum(nil)
	cb, ok = new(big.Int).SetString(address, 16)
	if !ok {
		return withdrawal_hash, ok
	}

	hFunc.Write(cb.Bytes())
	withdrawal_hash = hFunc.Sum(nil)

	hFunc.Reset()
	return withdrawal_hash, true
}

func WithdrawalHash(withdrawal []Transaction) ([]byte, bool) {
	var withdrawal_hash []byte
	hFunc := NewMiMC()
	var valid_withdrawal [][]byte
	for i := 0; i < len(withdrawal); i++ {
		if withdrawal[i].Type == "deposit" {
			cb, ok := WithdrawalItemHash(withdrawal[i].Amount, withdrawal[i].CurrencyTokenOrder, withdrawal[i].To)
			if !ok {
				return withdrawal_hash, ok
			}
			valid_withdrawal = append(valid_withdrawal, cb)
		}

	}
	for _, item := range valid_withdrawal {
		hFunc.Write(item)
		withdrawal_hash = hFunc.Sum(nil)
	}
	hFunc.Reset()
	return withdrawal_hash, true
}
