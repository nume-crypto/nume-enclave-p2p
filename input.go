package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"os"
	"time"
)

type NftTransaction struct {
	Id                 uint
	From               string
	To                 string
	NftContractAddress string
	Nonce              uint
	NftTokenId         uint
	Type               string
	Signature          string
	IsInvalid          bool
	Data               string
	CreatedAt          time.Time
}

type ValidatorKeys struct {
	BlsG1PublicKey      []string
	BlsG2PublicKey      []string
	HashedPublicKey     string
	EncryptedPrivateKey string
	CMKId               string
}

type InputData struct {
	MetaData         map[string]interface{}
	NewUserBalances  map[string]map[string]bool
	OldUserBalances  map[string]map[string]bool
	UserBalanceOrder map[string][]string
	Transactions     []NftTransaction
	ValidatorKeys    map[string]ValidatorKeys
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
	plan, err = os.ReadFile(path + "/validators.json")
	if err != nil {
		return input_data, "", err
	}
	err = json.Unmarshal(plan, &input_data.ValidatorKeys)
	if err != nil {
		return input_data, "", err
	}
	plan, err = os.ReadFile(path + "/user_balance_order.json")
	if err != nil {
		return input_data, "", err
	}
	err = json.Unmarshal(plan, &input_data.UserBalanceOrder)
	if err != nil {
		return input_data, "", err
	}

	return input_data, md5_sum_str, nil
}
