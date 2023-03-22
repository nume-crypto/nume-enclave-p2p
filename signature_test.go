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
	data := make(map[uint]string)
	data[1] = "AQICAHh2fn5fQzf0pR+JWPGR8yLKZjEywJ8b8umBI9kzCAFVdAEBaRXECZAR/aRfp8k2IeAhAAAAYjBgBgkqhkiG9w0BBwagUzBRAgEAMEwGCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMvHwEDnrIzDxSVPOQAgEQgB/0GX4mOgO5xq2emtxuQ/LzOtwhzFB0LyaQiFIrLgPv"
	data[2] = "AQICAHh2fn5fQzf0pR+JWPGR8yLKZjEywJ8b8umBI9kzCAFVdAEBaRXECZAR/aRfp8k2IeAhAAAAYjBgBgkqhkiG9w0BBwagUzBRAgEAMEwGCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMvHwEDnrIzDxSVPOQAgEQgB/0GX4mOgO5xq2emtxuQ/LzOtwhzFB0LyaQiFIrLgPv"
	data[3] = "AQICAHh2fn5fQzf0pR+JWPGR8yLKZjEywJ8b8umBI9kzCAFVdAEBaRXECZAR/aRfp8k2IeAhAAAAYjBgBgkqhkiG9w0BBwagUzBRAgEAMEwGCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMvHwEDnrIzDxSVPOQAgEQgB/0GX4mOgO5xq2emtxuQ/LzOtwhzFB0LyaQiFIrLgPp"
	cmk_ids := make([]string, 0)
	cmk_ids = append(cmk_ids, "c53fe209-f0a7-42d2-baec-9d8f286f5ce1")
	cmk_ids = append(cmk_ids, "c53fe209-f0a7-42d2-baec-9d8f286f5ce1")
	cmk_ids = append(cmk_ids, "c53fe209-f0a7-42d2-baec-9d8f286f5ce1")
	merchant_status := make(map[uint]bool)
	merchant_status[1] = true
	merchant_status[2] = false
	merchant_status[3] = true
	keys, failed_to_decrypt, successfully_decrypted, err := DecryptKeys(data, cmk_ids, kms_client, merchant_status)
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
	if len(failed_to_decrypt) != 1 {
		t.Errorf("Expected 1 failed_to_decrypt, got %d", len(failed_to_decrypt))
		return
	}
	if len(successfully_decrypted) != 1 {
		t.Errorf("Expected 1 successfully_decrypted, got %d", len(successfully_decrypted))
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
	data := make(map[uint]string)
	data[1] = "AQICAHh2fn5fQzf0pR+JWPGR8yLKZjEywJ8b8umBI9kzCAFVdAGN9n0CV+9w2tjYAqrVQWhcAAAAojCBnwYJKoZIhvcNAQcGoIGRMIGOAgEAMIGIBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDItcDL4lPiDkr7spewIBEIBbRb6Pltakos+qO7Ocpv0aiXT4GqF/8kMqm4pTFXMVO698rjL1u7PrudG09yiXvTVR3n/4hQrQf+LoGBi4CXTlc80z/f3OXTAB5tJCwNhOLAPgKZmo5X9MAT759A=="
	cmk_ids := make(map[uint]string)
	cmk_ids[1] = "c53fe209-f0a7-42d2-baec-9d8f286f5ce1"
	merchant_status := make(map[uint]bool)
	merchant_status[1] = true
	signature, _, _, _, err := SignMessage("10afdfd0a74398e23708f64b1ebdc41a78d85eebcb3b3d5fc7a9dd411f8f852d", data, cmk_ids, merchant_status)
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

func TestVerifyDigitalSignature(t *testing.T) {
	message := "10afdfd0a74398e23708f64b1ebdc41a78d85eebcb3b3d5fc7a9dd411f8f852d"
	aggregated_public_key_components := []string{
		"0abbdf1e100b1020b1178060decc6f5b257f0a39ea6f7baf506f8039b38f4caf",
		"0af72d12a44c88aeff2d19488c258e3f1a5006f895362a587bb3a2156208df34",
		"098b5d055ad316e2870258ab279bc467b5b93c114b6aabd472c571079c7ac3af",
		"08cc1c223fc7044fb37e527835fa1c7cae3acff0b19fa97d3170dad41673524d"}
	sig := "020da5e5a5fb0ca69acbdb01554ab258199f17588e9d4aec7d79f353cdad987280"
	verified, err := VerifyDigitalSignature(message, sig, aggregated_public_key_components)
	if err != nil {
		t.Errorf("Error verifying signature " + err.Error())
		return
	}
	if !verified {
		t.Errorf("Expected true, got false")
		return
	}

}
