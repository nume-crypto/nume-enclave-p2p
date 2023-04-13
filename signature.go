package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/ethereum/go-ethereum/crypto"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
)

func DecryptKeys(data map[string]ValidatorKeys, kms_client *kms.KMS) ([]string, []uint, []uint, error) {

	keys := make([]string, 0)
	failed_to_decrypt := make([]uint, 0)
	successfully_decrypted := make([]uint, 0)
	var global_err error
	var wg sync.WaitGroup
	for _, v := range data {
		wg.Add(1)
		go func(v ValidatorKeys) {
			b, err := base64.StdEncoding.DecodeString(v.EncryptedPrivateKey)
			if err != nil {
				global_err = err
			}
			input := &kms.DecryptInput{
				CiphertextBlob: b,
				GrantTokens: aws.StringSlice([]string{
					"GrantTokenType",
				}),
			}
			result, err := kms_client.Decrypt(input)
			if err != nil {
				fmt.Println(err, v.CMKId)
				// if user_status[k] {
				// 	failed_to_decrypt = append(failed_to_decrypt, k)
				// }
			} else {
				// if !user_status[k] {
				// 	successfully_decrypted = append(successfully_decrypted, k)
				// }
				keys = append(keys, string(result.Plaintext))
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

func SignMessage(message string, user_keys map[string]ValidatorKeys) (string, []string, []uint, []uint, error) {

	sess := session.Must(session.NewSession())
	kms_client := kms.New(sess, aws.NewConfig().WithRegion("us-east-1"))

	var aggregated_public_key_components []string
	keys, failed_to_decrypt, successfully_decrypted, err := DecryptKeys(user_keys, kms_client)
	if err != nil {
		return "", aggregated_public_key_components, failed_to_decrypt, successfully_decrypted, err
	}
	signature, aggregated_public_key_components, err := AggregateSignature(message, keys)
	if err != nil {
		return "", aggregated_public_key_components, failed_to_decrypt, successfully_decrypted, err
	}
	return signature, aggregated_public_key_components, failed_to_decrypt, successfully_decrypted, nil

}

func EthVerify(message string, sig string, pubkey string) bool {
	msg_bytes := []byte(message)
	full_msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg_bytes), msg_bytes)
	hash := solsha3.SoliditySHA3(
		[]string{"string"},
		[]interface{}{
			full_msg,
		},
	)
	sb, err := hex.DecodeString(sig[2:])
	if err != nil {
		fmt.Println("sb err:", err)
		return false
	}

	if sb[crypto.RecoveryIDOffset] == 27 || sb[crypto.RecoveryIDOffset] == 28 {
		sb[crypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1
	}

	recovered, err := crypto.SigToPub(hash, sb)
	if err != nil {
		return false
	}
	recovered_address := crypto.PubkeyToAddress(*recovered)
	// fmt.Println("recovered_address:", recovered_address.String())
	if recovered_address.String() == pubkey {
		return true
	} else {
		return false
	}
}
