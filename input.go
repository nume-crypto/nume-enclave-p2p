package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"os"
	"time"
)

type Transaction struct {
	From               string
	To                 string
	Amount             string
	Nonce              uint
	CurrencyTokenOrder uint
	Type               string
	Signature          string
	CreatedAt          time.Time
}

type UserKeys struct {
	BlsG1PublicKey      []string
	BlsG2PublicKey      []string
	HashedPublicKey     string
	EncryptedPrivateKey string
	CMKId               string
}

type InputData struct {
	MetaData        map[string]string
	NewUserBalances map[string]map[uint]string
	OldUserBalances map[string]map[uint]string
	Transactions    []Transaction
	UserKeys        map[string]UserKeys
}

func GetData(path string) (InputData, string, error) {
	var input_data InputData
	plan, err := os.ReadFile(path + "/transactions.json")
	if err != nil {
		return input_data, "", err
	}
	md5_sum := md5.Sum(bytes.TrimRight(plan, "\n"))
	md5_sum_str := hex.EncodeToString(md5_sum[:])
	err = json.Unmarshal(plan, &input_data.Transactions)
	if err != nil {
		return input_data, "", err
	}

	plan, err = os.ReadFile(path + "/prev_balances.json")
	if err != nil {
		return input_data, "", err
	}
	err = json.Unmarshal(plan, &input_data.OldUserBalances)
	if err != nil {
		return input_data, "", err
	}

	plan, err = os.ReadFile(path + "/new_balances.json")
	if err != nil {
		return input_data, "", err
	}
	err = json.Unmarshal(plan, &input_data.NewUserBalances)
	if err != nil {
		return input_data, "", err
	}

	plan, err = os.ReadFile(path + "/meta_data.json")
	if err != nil {
		return input_data, "", err
	}
	err = json.Unmarshal(plan, &input_data.MetaData)
	if err != nil {
		return input_data, "", err
	}
	plan, err = os.ReadFile(path + "/user_keys.json")
	if err != nil {
		return input_data, "", err
	}
	err = json.Unmarshal(plan, &input_data.UserKeys)
	if err != nil {
		return input_data, "", err
	}

	return input_data, md5_sum_str, nil
}
