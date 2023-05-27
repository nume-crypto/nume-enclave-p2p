package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type KmsRequest struct {
	AwsAccessKeyId     string `json:"AwsAccessKeyId"`
	AwsSecretAccessKey string `json:"AwsSecretAccessKey"`
	AwsSessionToken    string `json:"AwsSessionToken"`
	CipherText         string `json:"CipherText"`
}

func DecryptKeys(data map[string]ValidatorKeys, kms_payload KmsPayload) ([]string, []uint, []uint, error) {

	keys := make([]string, 0)
	failed_to_decrypt := make([]uint, 0)
	successfully_decrypted := make([]uint, 0)
	var global_err error
	var wg sync.WaitGroup
	for _, v := range data {
		wg.Add(1)
		go func(v ValidatorKeys) {
			request := KmsRequest{
				AwsAccessKeyId:     kms_payload.AccessKeyId,
				AwsSecretAccessKey: kms_payload.SecretAccessKey,
				AwsSessionToken:    kms_payload.Token,
				CipherText:         v.EncryptedPrivateKey,
			}
			payload_bytes, err := json.Marshal(request)
			if err != nil {
				fmt.Printf("Error: %s", err)
			}
			kmslib := "./kmstool_enclave"
			cmd := exec.Command(kmslib, string(payload_bytes))
			cmd.Dir = "app/"
			stdout, err := cmd.Output()
			if err != nil {
				fmt.Println("E1X : " + fmt.Sprint(err) + ": stdout : " + string(stdout))
			} else {
				keys = append(keys, string(stdout))
			}
			wg.Done()
		}(v)
	}
	wg.Wait()
	if global_err != nil {
		return keys, failed_to_decrypt, successfully_decrypted, global_err
	}
	return keys, failed_to_decrypt, successfully_decrypted, nil
}

func AggregateSignature(message string, keys []string) (string, []string, error) {
	app := "./bn256_aggregatesign"
	var aggregated_public_key_components []string
	var args []string
	args = append(args, message)
	args = append(args, keys...)
	cmd := exec.Command(app, args...)
	stdout, err := cmd.Output()

	if err != nil {
		return "", aggregated_public_key_components, err
	}
	result := string(stdout)
	subres := strings.Split(result, "Aggregated Signature In Hex of Length 66")[1]

	aggregated_public_key_components = append(aggregated_public_key_components, strings.TrimSpace(strings.Split(subres, `"`)[3]))
	aggregated_public_key_components = append(aggregated_public_key_components, strings.TrimSpace(strings.Split(subres, `"`)[5]))
	aggregated_public_key_components = append(aggregated_public_key_components, strings.TrimSpace(strings.Split(subres, `"`)[7]))
	aggregated_public_key_components = append(aggregated_public_key_components, strings.TrimSpace(strings.Split(subres, `"`)[9]))

	return strings.TrimSpace(strings.Split(subres, `"`)[1]), aggregated_public_key_components, err
}

func SignMessage(message string, user_keys map[string]ValidatorKeys, kms_payload KmsPayload) (string, []string, []uint, []uint, error) {

	var aggregated_public_key_components []string
	keys, failed_to_decrypt, successfully_decrypted, err := DecryptKeys(user_keys, kms_payload)
	if err != nil {
		return "", aggregated_public_key_components, failed_to_decrypt, successfully_decrypted, err
	}
	signature, aggregated_public_key_components, err := AggregateSignature(message, keys)
	if err != nil {
		return "", aggregated_public_key_components, failed_to_decrypt, successfully_decrypted, err
	}
	return signature, aggregated_public_key_components, failed_to_decrypt, successfully_decrypted, nil

}

func RecoverPlain(sighash common.Hash, R, S, Vb *big.Int, homestead bool) (string, string, common.Address, error) {
	signature := ""
	pubkey := ""
	if Vb.BitLen() > 8 {
		return signature, pubkey, common.Address{}, errors.New("invalid signature v byte")
	}
	V := byte(Vb.Uint64() - 27)
	if !crypto.ValidateSignatureValues(V, R, S, homestead) {
		return signature, pubkey, common.Address{}, errors.New("invalid signature")
	}
	// encode the signature in uncompressed format
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, crypto.SignatureLength)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V
	// recover the public key from the signature
	signature = hex.EncodeToString(sig)
	pub, err := crypto.Ecrecover(sighash[:], sig)
	pubkey = hex.EncodeToString(pub)
	if err != nil {
		return signature, pubkey, common.Address{}, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return signature, pubkey, common.Address{}, errors.New("invalid public key")
	}
	var addr common.Address
	copy(addr[:], crypto.Keccak256(pub[1:])[12:])
	return signature, pubkey, addr, nil
}

func decodeTransactionInputData(abi *abi.ABI, data []byte) (map[string]interface{}, string, error) {
	method_data := data[:4]
	input_data := data[4:]
	method, err := abi.MethodById(method_data)
	inputsMap := make(map[string]interface{})
	if err != nil {
		return inputsMap, "", err
	}
	if err := method.Inputs.UnpackIntoMap(inputsMap, input_data); err != nil {
		return inputsMap, "", err
	}
	return inputsMap, method.Name, nil
}

func GetAmountAndTokenAddress(tx *types.Transaction, currencies []string) (string, string, string, error) {
	amount := ""
	token_address := ""
	to := tx.To().Hex()
	flag := false
	for _, currency := range currencies {
		if strings.EqualFold(currency, to) {
			flag = true
			break
		}
	}
	if !flag {
		return tx.Value().String(), "0x0000000000000000000000000000000000000000", tx.To().Hex(), nil
	}
	if tx.Data() == nil {
		return amount, token_address, to, errors.New("invalid data")
	}
	abi_path := "erc20.abi"
	path, err := filepath.Abs(abi_path)
	if err != nil {
		return amount, token_address, to, err
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return amount, token_address, to, err
	}
	abi, err := abi.JSON(strings.NewReader(string(file)))
	if err != nil {
		return amount, token_address, to, err
	}
	input, method, err := decodeTransactionInputData(&abi, tx.Data())
	if err != nil {
		return amount, token_address, to, err
	}
	if method != "transfer" {
		return amount, token_address, to, errors.New("invalid method")
	}
	amount = input["amount"].(*big.Int).String()
	to = input["to"].(common.Address).Hex()
	token_address = tx.To().Hex()
	return amount, token_address, to, nil
}

func VerifyData(input_tx Transaction, currencies []string) (bool, error) {
	tx_bytes, err := hex.DecodeString(input_tx.Data[2:])
	if err != nil {
		return false, err
	}

	eth_tx := new(types.Transaction)
	if err := eth_tx.UnmarshalBinary(tx_bytes); err != nil {
		return false, err
	}
	signer := types.NewLondonSigner(eth_tx.ChainId())
	V, R, S := eth_tx.RawSignatureValues()

	V = new(big.Int).Add(V, big.NewInt(27))
	_, _, addr, err := RecoverPlain(signer.Hash(eth_tx), R, S, V, true)
	if err != nil {
		return false, err
	}
	// if transactio.tyupe contrains nft if condition
	var amt_or_token_id, token_address_or_currency, to, from string
	if strings.Contains(input_tx.Type, "nft") {
		amt_or_token_id, token_address_or_currency, to, from, err = GetNftFromAndTo(eth_tx)
		if err != nil {
			return false, err
		}
	} else {
		from = addr.Hex()
		amt_or_token_id, token_address_or_currency, to, err = GetAmountAndTokenAddress(eth_tx, currencies)
		if err != nil {
			return false, err
		}
	}

	gen_tx := Transaction{
		From:                         from,
		To:                           to,
		CurrencyOrNftContractAddress: token_address_or_currency,
		AmountOrNftTokenId:           amt_or_token_id,
		Nonce:                        uint(eth_tx.Nonce()),
	}
	if !strings.EqualFold(input_tx.From, gen_tx.From) {
		return false, errors.New("from not equal " + input_tx.From + " " + gen_tx.From)
	}
	if !strings.EqualFold(input_tx.To, gen_tx.To) {
		return false, errors.New("to not equal " + input_tx.To + " " + gen_tx.To)
	}
	if input_tx.AmountOrNftTokenId != gen_tx.AmountOrNftTokenId {
		return false, errors.New("amount not equal " + input_tx.AmountOrNftTokenId + " " + gen_tx.AmountOrNftTokenId)
	}
	if !strings.EqualFold(input_tx.CurrencyOrNftContractAddress, gen_tx.CurrencyOrNftContractAddress) {
		return false, errors.New("currency not equal " + input_tx.CurrencyOrNftContractAddress + " " + gen_tx.CurrencyOrNftContractAddress)
	}
	if input_tx.Nonce != gen_tx.Nonce {
		return false, errors.New("nonce not equal " + strconv.Itoa(int(input_tx.Nonce)) + " " + strconv.Itoa(int(gen_tx.Nonce)))
	}
	return true, nil
}

func GetNftFromAndTo(tx *types.Transaction) (string, string, string, string, error) {
	nft_token_id := "0"
	nft_token_address := ""
	to := tx.To().Hex()
	from := ""
	if tx.Data() == nil {
		return nft_token_id, nft_token_address, to, from, errors.New("invalid data")
	}
	abi_path := "erc721.abi"
	path, err := filepath.Abs(abi_path)
	if err != nil {
		return nft_token_id, nft_token_address, to, from, err
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return nft_token_id, nft_token_address, to, from, err
	}
	abi, err := abi.JSON(strings.NewReader(string(file)))
	if err != nil {
		return nft_token_id, nft_token_address, to, from, err
	}
	input, method, err := decodeTransactionInputData(&abi, tx.Data())
	if err != nil {
		return nft_token_id, nft_token_address, to, from, err
	}
	if method != "transferFrom" {
		return nft_token_id, nft_token_address, to, from, errors.New("invalid method")
	}
	nft_token_id = input["tokenId"].(*big.Int).String()
	to = input["to"].(common.Address).Hex()
	nft_token_address = tx.To().Hex()
	from = input["from"].(common.Address).Hex()
	return nft_token_id, nft_token_address, to, from, nil
}

func NftTradeMessage(user, nft_contract_address, nft_token_id, currency_address, amount, bn string) string {
	hash := solsha3.SoliditySHA3(
		[]string{"address", "address", "uint256", "address", "uint256", "uint256"},
		[]interface{}{
			user,
			nft_contract_address,
			nft_token_id,
			currency_address,
			amount,
			bn,
		},
	)
	return hex.EncodeToString(hash)
}

func EthVerify(message string, sig string, pubkey string) bool {
	msg_bytes := []byte(message)
	fullMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg_bytes), msg_bytes)
	hash := crypto.Keccak256Hash([]byte(fullMessage))
	sb, err := hex.DecodeString(sig[2:])
	if err != nil {
		fmt.Println("sb err:", err)
		return false
	}

	if sb[crypto.RecoveryIDOffset] == 27 || sb[crypto.RecoveryIDOffset] == 28 {
		sb[crypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1
	}

	recovered, err := crypto.SigToPub(hash.Bytes(), sb)
	if err != nil {
		return false
	}
	recoveredAddr := crypto.PubkeyToAddress(*recovered)
	if strings.EqualFold(recoveredAddr.String(), pubkey) {
		return true
	} else {
		return false
	}
}
