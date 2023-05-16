package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

func TestDecrypt(t *testing.T) {
	sess := session.Must(session.NewSession())
	kms_client := kms.New(sess, aws.NewConfig().WithRegion("us-east-1"))
	user_keys := make(map[string]ValidatorKeys)
	user_keys["1"] = ValidatorKeys{
		EncryptedPrivateKey: "AQICAHh2fn5fQzf0pR+JWPGR8yLKZjEywJ8b8umBI9kzCAFVdAEBaRXECZAR/aRfp8k2IeAhAAAAYjBgBgkqhkiG9w0BBwagUzBRAgEAMEwGCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMvHwEDnrIzDxSVPOQAgEQgB/0GX4mOgO5xq2emtxuQ/LzOtwhzFB0LyaQiFIrLgPv",
		CMKId:               "c53fe209-f0a7-42d2-baec-9d8f286f5ce1",
	}
	user_keys["2"] = ValidatorKeys{
		EncryptedPrivateKey: "AQICAHh2fn5fQzf0pR+JWPGR8yLKZjEywJ8b8umBI9kzCAFVdAEBaRXECZAR/aRfp8k2IeAhAAAAYjBgBgkqhkiG9w0BBwagUzBRAgEAMEwGCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMvHwEDnrIzDxSVPOQAgEQgB/0GX4mOgO5xq2emtxuQ/LzOtwhzFB0LyaQiFIrLgPv",
		CMKId:               "c53fe209-f0a7-42d2-baec-9d8f286f5ce1",
	}
	user_keys["3"] = ValidatorKeys{
		EncryptedPrivateKey: "AQICAHh2fn5fQzf0pR+JWPGR8yLKZjEywJ8b8umBI9kzCAFVdAEBaRXECZAR/aRfp8k2IeAhAAAAYjBgBgkqhkiG9w0BBwagUzBRAgEAMEwGCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMvHwEDnrIzDxSVPOQAgEQgB/0GX4mOgO5xq2emtxuQ/LzOtwhzFB0LyaQiFIrLgPp",
		CMKId:               "c53fe209-f0a7-42d2-baec-9d8f286f5ce1",
	}
	keys, _, _, err := DecryptKeys(user_keys, kms_client)
	if err != nil {
		t.Errorf("Error decrypting keys " + err.Error())
		return
	}
	if len(keys) != 2 {
		t.Errorf("Expected 2 key, got %d", len(keys))
		return
	}
	if keys[0] != "1234" {
		t.Errorf("Expected 1234, got %s", keys[0])
		return
	}
}

func TestAggregateSignature(t *testing.T) {
	signature, aggregated_public_key_components, err := AggregateSignature("10afdfd0a74398e23708f64b1ebdc41a78d85eebcb3b3d5fc7a9dd411f8f852d", []string{"d1f0f4e6df9803f1c94fe46214037c2fa926238de5504315abac0e9a5c189843", "b155212c78e165ab377c6e1c142ba828a94a05699b60c0d12c279cd7e5a3f4ae"})
	if err != nil {
		t.Errorf("Error aggregating signature" + err.Error())
		return
	}
	if len(signature) != 66 {
		t.Errorf("Expected 66 signature, got %d", len(signature))
		return
	}
	if len(aggregated_public_key_components) != 4 {
		t.Errorf("Expected 4 aggregated_public_key_components, got %d", len(aggregated_public_key_components))
		return
	}
	if signature != "02179840e62375b8f8f7caeeee710997fb9d4c365ba4330066bb767b2f6d514de0" {
		t.Errorf("Expected 02179840e62375b8f8f7caeeee710997fb9d4c365ba4330066bb767b2f6d514de0, got %s", signature)
		return
	}
	aggregated_public_key_components_expected := []string{
		"1aaa1aa48d2a03f922d5d1e851850c25681f94b72cb5e5974cbdc2648214cfd3",
		"18efbac529960229814ee922e00bd4fe57689ad2f4bca2522bd8d57cc9ed2d2a",
		"2a008c2e6f4d94f03490d077033c8d56e11e443f542ad59c7c58f2a68137db53",
		"008bc79f2e7b926b9f3824a2eadb2b92398e7e5450d2ab84226ea70c7fd177d9",
	}
	for i, pk := range aggregated_public_key_components {
		if len(pk) != 64 {
			t.Errorf("Expected 64 pk, got %d", len(pk))
		}
		if pk != aggregated_public_key_components_expected[i] {
			t.Errorf("Expected %s, got %s", aggregated_public_key_components_expected[i], pk)
			return
		}
	}
}

func TestSignMessage(t *testing.T) {
	user_keys := make(map[string]ValidatorKeys)
	user_keys["162ef608bf92f47846fbf53481f1b0504e3bd1f1678376b20139bd94cf0003eb"] = ValidatorKeys{
		EncryptedPrivateKey: "AQICAHh2fn5fQzf0pR+JWPGR8yLKZjEywJ8b8umBI9kzCAFVdAGN9n0CV+9w2tjYAqrVQWhcAAAAojCBnwYJKoZIhvcNAQcGoIGRMIGOAgEAMIGIBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDItcDL4lPiDkr7spewIBEIBbRb6Pltakos+qO7Ocpv0aiXT4GqF/8kMqm4pTFXMVO698rjL1u7PrudG09yiXvTVR3n/4hQrQf+LoGBi4CXTlc80z/f3OXTAB5tJCwNhOLAPgKZmo5X9MAT759A==",
		CMKId:               "c53fe209-f0a7-42d2-baec-9d8f286f5ce1",
	}
	signature, _, _, _, err := SignMessage("10afdfd0a74398e23708f64b1ebdc41a78d85eebcb3b3d5fc7a9dd411f8f852d", user_keys)
	if err != nil {
		t.Errorf("Error signing message" + err.Error())
		return
	}
	if len(signature) != 66 {
		t.Errorf("Expected 66 signature, got %d", len(signature))
		return
	}
	if signature != "031ccc5203cc6117f641d2aa552e2d42e77e5db495c94f904de52d0c673db754fb" {
		t.Errorf("Expected 031ccc5203cc6117f641d2aa552e2d42e77e5db495c94f904de52d0c673db754fb, got %s", signature)
		return
	}
}

func TestNftTradeMessage(t *testing.T) {
	sig := NftTradeMessage("0xCcFf350Ef46B85228d6650a802107e58BF6A32Ab", "0x5FbDB2315678afecb367f032d93F642f64180aa3", "1", "0x5FbDB2315678afecb367f032d93F642f64180aa3", "1", "1")
	if sig != "fb053228ebfe580d705665ce141bab48720b7a5dc320ddd125ff54cb2caadb1c" {
		t.Errorf("Expected fb053228ebfe580d705665ce141bab48720b7a5dc320ddd125ff54cb2caadb1c, got %s", sig)
		return
	}
}

func TestEthVerify(t *testing.T) {
	verify := EthVerify("fb053228ebfe580d705665ce141bab48720b7a5dc320ddd125ff54cb2caadb1c", "0xa72a93780d73208fd0790c614b9acd019d9be10fe3cbb0a058a964f9ba3fe27678e85d6f00ba93f9091bfac96f2c7e46ccbe3da6877de490f9664ae1ef9f2b091c", "0x13eB1ADfee2fa4813B658349dda1dcD051f89a34")
	if verify != true {
		t.Errorf("Expected true, got %t", verify)
		return
	}
}
