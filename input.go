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
	Id                           uint
	From                         string
	To                           string
	AmountOrNftTokenId           string
	Nonce                        uint
	CurrencyOrNftContractAddress string
	Type                         string
	Signature                    string
	IsInvalid                    bool
	Data                         string
	L2Minted                     bool
	CreatedAt                    time.Time
}

type ValidatorKeys struct {
	BlsG1PublicKey      []string
	BlsG2PublicKey      []string
	HashedPublicKey     string
	EncryptedPrivateKey string
	CMKId               string
}

type InputData struct {
	MetaData             map[string]interface{}
	NewUserBalances      map[string]map[string]string
	OldUserBalances      map[string]map[string]string
	NewUserBalanceOrder  map[string][]string
	OldUserBalanceOrder  map[string][]string
	Transactions         []Transaction
	ValidatorKeys        map[string]ValidatorKeys
	AddressPublicKeyData map[string]string
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
	plan, err = os.ReadFile(path + "/new_user_balance_order.json")
	if err != nil {
		return input_data, "", err
	}
	err = json.Unmarshal(plan, &input_data.NewUserBalanceOrder)
	if err != nil {
		return input_data, "", err
	}
	plan, err = os.ReadFile(path + "/old_user_balance_order.json")
	if err != nil {
		return input_data, "", err
	}
	err = json.Unmarshal(plan, &input_data.OldUserBalanceOrder)
	if err != nil {
		return input_data, "", err
	}
	plan, err = os.ReadFile(path + "/address_public_key_map.json")
	if err != nil {
		return input_data, "", err
	}
	err = json.Unmarshal(plan, &input_data.AddressPublicKeyData)
	if err != nil {
		return input_data, "", err
	}

	return input_data, md5_sum_str, nil
}
