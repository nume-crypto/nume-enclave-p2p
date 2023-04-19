package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mdlayher/vsock"
)

type Transaction struct {
	Id        uint
	From      string
	To        string
	Amount    string
	Nonce     uint
	Currency  string
	Type      string
	Signature string
	IsInvalid bool
	CreatedAt time.Time
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
	NewUserBalances  map[string]map[string]string
	OldUserBalances  map[string]map[string]string
	UserBalanceOrder map[string][]string
	Transactions     []Transaction
	ValidatorKeys    map[string]ValidatorKeys
	AddressPublicKeyData map[string]string
	KmsPayload       KmsPayload
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


func GetDataOverSocket(con *vsock.Conn) (InputData, string, error) {
	//path := "./temp_data"
	var input_data InputData

	fmt.Println("Sending to Server: transactions.json")
	con.Write([]byte("transactions.json" + "\n"))
	//plan, err := os.ReadFile(path + "/transactions.json")
	plan, err := bufio.NewReader(con).ReadBytes('\n')
	if err != nil {
		return input_data, "", err
	}
	md5_sum := md5.Sum(bytes.TrimRight(plan, "\n"))
	md5_sum_str := hex.EncodeToString(md5_sum[:])
	err = json.Unmarshal(plan, &input_data.Transactions)
	if err != nil {
		return input_data, "", err
	}

	fmt.Println("Sending to Server: prev_balances.json")
	con.Write([]byte("prev_balances.json" + "\n"))
	//plan, err = os.ReadFile(path + "/prev_balances.json")
	plan, err = bufio.NewReader(con).ReadBytes('\n')
	if err != nil {
		return input_data, "", err
	}
	err = json.Unmarshal(plan, &input_data.OldUserBalances)
	if err != nil {
		return input_data, "", err
	}

	fmt.Println("Sending to Server: new_balances.json")
	con.Write([]byte("new_balances.json" + "\n"))
	//plan, err = os.ReadFile(path + "/new_balances.json")
	plan, err = bufio.NewReader(con).ReadBytes('\n')
	if err != nil {
		return input_data, "", err
	}
	err = json.Unmarshal(plan, &input_data.NewUserBalances)
	if err != nil {
		return input_data, "", err
	}
	
	fmt.Println("Sending to Server: meta_data.json")
	con.Write([]byte("meta_data.json" + "\n"))
	//plan, err = os.ReadFile(path + "/meta_data.json")
	plan, err = bufio.NewReader(con).ReadBytes('\n')
	if err != nil {
		return input_data, "", err
	}
	err = json.Unmarshal(plan, &input_data.MetaData)
	if err != nil {
		return input_data, "", err
	}

	fmt.Println("Sending to Server: validators.json")
	con.Write([]byte("validators.json" + "\n"))
	//plan, err = os.ReadFile(path + "/validators.json")
	plan, err = bufio.NewReader(con).ReadBytes('\n')
	if err != nil {
		return input_data, "", err
	}
	err = json.Unmarshal(plan, &input_data.ValidatorKeys)
	if err != nil {
		return input_data, "", err
	}

	fmt.Println("Sending to Server: user_balance_order.json")
	con.Write([]byte("user_balance_order.json" + "\n"))
	//plan, err = os.ReadFile(path + "/user_balance_order.json")
	plan, err = bufio.NewReader(con).ReadBytes('\n')
	if err != nil {
		return input_data, "", err
	}
	err = json.Unmarshal(plan, &input_data.UserBalanceOrder)
	if err != nil {
		return input_data, "", err
	}

	fmt.Println("Sending to Server: address_public_key_map.json")
	con.Write([]byte("address_public_key_map.json" + "\n"))
	//plan, err = os.ReadFile(path + "/address_public_key_map.json")
	plan, err = bufio.NewReader(con).ReadBytes('\n')
	if err != nil {
		return input_data, "", err
	}
	err = json.Unmarshal(plan, &input_data.AddressPublicKeyData)
	if err != nil {
		return input_data, "", err
	}

	fmt.Println("Sending to Server: FETCH_CREDENTIALS")
	con.Write([]byte("FETCH_CREDENTIALS" + "\n"))
	plan, _ = bufio.NewReader(con).ReadBytes('\n')
	err = json.Unmarshal(plan, &input_data.KmsPayload)
	if err != nil {
		return input_data, "", err
	}
	fmt.Println("Message from server: ", string(plan))

	return input_data, md5_sum_str, nil
}
