package main

import (
	"encoding/hex"
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
		t.Errorf("Failed to hash message")
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
		t.Errorf("Failed to hash message")
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
		t.Errorf("Failed to hash message")
		return
	}
}
