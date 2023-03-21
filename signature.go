package main

import (
	"encoding/base64"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

func DecryptKeys(data map[uint]string, cmkIds []string, kms_client *kms.KMS, merchant_status map[uint]bool) ([]string, []uint, []uint, error) {

	var keys []string
	var failed_to_decrypt []uint
	var successfully_decrypted []uint

	for k, v := range data {
		b, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return keys, failed_to_decrypt, successfully_decrypted, err
		}
		input := &kms.DecryptInput{
			CiphertextBlob: b,
			GrantTokens: aws.StringSlice([]string{
				"GrantTokenType",
			}),
		}
		result, err := kms_client.Decrypt(input)
		if err != nil {
			if merchant_status[k] {
				failed_to_decrypt = append(failed_to_decrypt, k)
			}
		} else {
			if !merchant_status[k] {
				successfully_decrypted = append(successfully_decrypted, k)
			}
			keys = append(keys, string(result.Plaintext))
		}

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

func SignMessage(message string, merchant_key_id_map map[uint]string, merchant_cmk_id_map map[uint]string, merchant_status map[uint]bool) (string, []string, []uint, []uint, error) {

	sess := session.Must(session.NewSession())
	kms_client := kms.New(sess, aws.NewConfig().WithRegion("us-east-1"))

	var aggregated_public_key_components []string
	cmk_list := make([]string, 0, len(merchant_cmk_id_map))

	for _, cmks := range merchant_cmk_id_map {
		cmk_list = append(cmk_list, cmks)
	}

	keys, failed_to_decrypt, successfully_decrypted, err := DecryptKeys(merchant_key_id_map, cmk_list, kms_client, merchant_status)
	if err != nil {
		return "", aggregated_public_key_components, failed_to_decrypt, successfully_decrypted, err
	}
	signature, aggregated_public_key_components, err := AggregateSignature(message, keys)
	if err != nil {
		return "", aggregated_public_key_components, failed_to_decrypt, successfully_decrypted, err
	}
	return signature, aggregated_public_key_components, failed_to_decrypt, successfully_decrypted, nil

}
