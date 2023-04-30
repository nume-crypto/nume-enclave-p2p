package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
)

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func GetLeafHash(address string, root string, nonce uint) []byte {
	nonce_bi := big.NewInt(int64(nonce))
	hash := solsha3.SoliditySHA3(
		[]string{"address", "bytes32", "uint256"},
		[]interface{}{
			address,
			root,
			nonce_bi,
		},
	)
	return hash
}

func NestedMapsEqual(m1, m2 map[string]map[string]bool) bool {
	if len(m1) != len(m2) {
		fmt.Println("len(m1)", len(m1), "len(m2)", len(m2))
		return false
	}
	for k, v1 := range m1 {
		if v2, ok := m2[k]; !ok || !MapsEqual(v1, v2) {
			fmt.Println("v1", v1, "v2", v2)
			return false
		}
	}
	return reflect.DeepEqual(m1, m2)
}

func MapsEqual(m1, m2 map[string]bool) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v1 := range m1 {
		if v2, ok := m2[k]; !ok || v1 != v2 {
			return false
		}
	}
	return reflect.DeepEqual(m1, m2)
}

func GetBalancesRoot(balances map[string]bool, user_balance_order []string, max_num_balances int) (string, bool) {

	balances_tree := &MerkleTree{}
	var balances_data = make([][]byte, max_num_balances)
	for i := 0; i < max_num_balances; i++ {
		if i < len(balances) {
			nft_address := strings.Split(user_balance_order[i], "-")[0]
			nft_token_id := strings.Split(user_balance_order[i], "-")[0]
			hash := solsha3.SoliditySHA3(
				[]string{"address", "uint256", "bytes32"},
				[]interface{}{
					nft_address,
					nft_token_id,
					"0x0000000000000000000000000000000000000000",
				},
			)
			balances_data[i] = hash
		} else {
			hash := solsha3.SoliditySHA3(
				[]string{"address", "uint256", "bytes32"},
				[]interface{}{
					"0x0000000000000000000000000000000000000000",
					"0",
					"0x0000000000000000000000000000000000000000",
				},
			)
			balances_data[i] = hash
		}
	}
	balances_tree = NewMerkleTree(balances_data)
	return hex.EncodeToString(balances_tree.Root), true
}

// func EncryptTransactionKMSPubkey(tx *Transaction, block_number float64, public_key_hex string) (string, error) {
// 	hashed_message := DigitalSignatureMessage(tx.From, tx.To, tx.Currency, tx.Amount, uint64(tx.Nonce), int64(block_number))
// 	kstr2, err := hex.DecodeString(public_key_hex)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to decode public key: %w", err)
// 	}

// 	pk, err := PemToPubkey(kstr2)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to convert public key: %w", err)
// 	}

// 	message := []byte(hashed_message)

// 	publicKeyEcies := ecies.ImportECDSAPublic(pk)

// 	encryptedMessage, err := ecies.Encrypt(rand.Reader, publicKeyEcies, message, nil, nil)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to encrypt message: %w", err)
// 	}

// 	return hex.EncodeToString(encryptedMessage), nil
// }

var (
	OidPublicKeyECDSA = asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}
)

type PublicKeyInfo struct {
	Raw       asn1.RawContent
	Algorithm pkix.AlgorithmIdentifier
	PublicKey asn1.BitString
}

func PemToPubkey(publicKey []byte) (*ecdsa.PublicKey, error) {
	var pub PublicKeyInfo
	rest, err := asn1.Unmarshal(publicKey, &pub)
	if err != nil || len(rest) > 0 {
		return nil, fmt.Errorf("error unmarshaling public key: %w", err)
	}
	if !pub.Algorithm.Algorithm.Equal(OidPublicKeyECDSA) {
		return nil, errors.New("not a ECDSA public key")
	}

	// Convert to ecdsa.PublicKey
	pk, err := crypto.UnmarshalPubkey(pub.PublicKey.Bytes)
	fmt.Println("reflect.TypeOf ", reflect.TypeOf(pk.Curve))

	publicKeyBytes := elliptic.Marshal(pk.Curve, pk.X, pk.Y)
	publicKeyHex := hex.EncodeToString(publicKeyBytes)
	fmt.Printf("ECDSA public key: %s\n", publicKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal secp256k1 curve point: %w", err)
	}

	return pk, nil
}

func ToECDSAPub(pub []byte) *ecdsa.PublicKey {
	if len(pub) == 0 {
		return nil
	}
	curve := crypto.S256()
	x, y := elliptic.Unmarshal(curve, pub)
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}
}

// func EncryptTransactionECDSAPubKey(tx *Transaction, block_number float64, ecdsaPublicKeyHex string) (string, error) {
// 	hashed_message := DigitalSignatureMessage(tx.From, tx.To, tx.Currency, tx.Amount, uint64(tx.Nonce), int64(block_number))
// 	// ecdsaPublicKeyHex = "040fc7ea6980106e7e7e303f27b70773182ae0c1e2681e0170d2f0d1704adeab1d031ad29c18d7a1292e20c6d29439c65b54833f4897780bdc6c43864289d8d134"
// 	// decode the hex string into a byte array
// 	publicKeyBytes, _ := hex.DecodeString(ecdsaPublicKeyHex)

// 	publicKey := ToECDSAPub(publicKeyBytes)

// 	message := []byte(hashed_message)

// 	publicKeyEcies := ecies.ImportECDSAPublic(publicKey)
// 	encryptedMessage, err := ecies.Encrypt(rand.Reader, publicKeyEcies, message, nil, nil)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("encryptedMessage")
// 	str := hex.EncodeToString(encryptedMessage)
// 	fmt.Println(str)

// 	return hex.EncodeToString(encryptedMessage), nil
// }
