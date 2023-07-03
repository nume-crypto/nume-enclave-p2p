package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	solsha3 "github.com/miguelmota/go-solidity-sha3"
)

type SettlementRequest struct {
	SettlementId                         uint                   `json:"settlementId" binding:"required"`
	Root                                 string                 `json:"root" binding:"required"`
	NftRoot                              string                 `json:"nftRoot" binding:"required"`
	AggregatedSignature                  string                 `json:"aggregatedSignature" binding:"required"`
	AggregatedPublicKeyComponents        []string               `json:"aggregatedPublicKeyComponents" binding:"required"`
	BlockNumber                          string                 `json:"blockNumber" binding:"required"`
	QueueHash                            string                 `json:"queueHash" binding:"required"` // deposit
	QueueIndex                           int                    `json:"queueIndex"`
	NftQueueHash                         string                 `json:"nftQueueHash" binding:"required"` // deposit nft
	NftQueueIndex                        int                    `json:"nftQueueIndex"`
	WithdrawalHash                       string                 `json:"withdrawalHash" binding:"required"` // withdrawal
	WithdrawalAmountOrTokenId            []string               `json:"withdrawalAmountOrTokenId" binding:"required"`
	WithdrawalAddresses                  []string               `json:"withdrawalAddresses" binding:"required"`
	WithdrawalCurrencyOrNftContract      []string               `json:"withdrawalCurrencyOrNftContract" binding:"required"`
	WithdrawalL2Minted                   []bool                 `json:"withdrawalL2Minted" binding:"required"`
	WithdrawalType                       []int                  `json:"withdrawalType" binding:"required"`
	ContractWithdrawalAddresses          []string               `json:"contractWithdrawalAddresses" binding:"required"` // contract withdrawal
	ContractWithdrawalAmounts            []string               `json:"contractWithdrawalAmounts" binding:"required"`
	ContractWithdrawalTokens             []string               `json:"contractWithdrawalTokes" binding:"required"`
	ContractWithdrawalQueueIndex         int                    `json:"contractWithdrawalQueueIndex"`
	NftContractWithdrawalAddresses       []string               `json:"nftContractWithdrawalAddresses" binding:"required"` // nft contract withdrawal
	NftContractWithdrawalTokensIds       []string               `json:"nftContractWithdrawalTokenIds" binding:"required"`
	NftContractWithdrawalContractAddress []string               `json:"nftContractWithdrawalContractAddress" binding:"required"`
	NftContractWithdrawalQueueIndex      int                    `json:"nftContractWithdrawalQueueIndex"`
	NftContractWithdrawalL2Minted        []bool                 `json:"nftContractWithdrawalL2Minted" binding:"required"`
	Message                              string                 `json:"message" binding:"required"` // message
	UsersUpdated                         map[string]interface{} `json:"usersUpdated" binding:"required"`
	NftCollectionsCreated                map[int]string         `json:"nftCollectionsCreated" binding:"required"`
	UserListerNonce                      map[string][]uint      `json:"usedListerNonce" binding:"required"`
	SignatureRecordedAt                  time.Time              `json:"signatureRecordedAt" binding:"required"`
	SettlementStartedAt                  time.Time              `json:"settlementStartedAt" binding:"required"`
}

func PrettyPrint(text string, v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	fmt.Println(text)
	if err == nil {
		fmt.Println(string(b))
	}
	return
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func GetLeafHash(address string, root string, nonce uint, used_lister_nonce []uint) []byte {
	types := []string{}
	values := []interface{}{}
	optimized_used_lister_nonce := GetOptimizedNonce(used_lister_nonce)

	for _, nonce := range optimized_used_lister_nonce {
		types = append(types, "uint256")
		values = append(values, big.NewInt(int64(nonce)))
	}
	used_lister_nonce_hash := solsha3.SoliditySHA3(types, values)
	if len(used_lister_nonce) == 0 {
		used_lister_nonce_hash = []byte{}
	}
	nonce_bi := big.NewInt(int64(nonce))
	hash := solsha3.SoliditySHA3(
		[]string{"address", "bytes32", "uint256", "bytes32"},
		[]interface{}{
			address,
			root,
			nonce_bi,
			used_lister_nonce_hash,
		},
	)
	return hash
}

func NestedMapsEqual(m1, m2 map[string]map[string]string) bool {
	defer TimeTrack(time.Now(), "NestedMapsEqual")
	if len(m1) != len(m2) {
		fmt.Println("len(m1)", len(m1), "len(m2)", len(m2))
		return false
	}
	for k, v1 := range m1 {
		if v2, ok := m2[k]; !ok || !MapsEqual(v1, v2) {
			PrettyPrint("v1", v1)
			PrettyPrint("v2", v2)
			return false
		}
	}
	return reflect.DeepEqual(m1, m2)
}

func MapsEqual(m1, m2 map[string]string) bool {
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

func GetBalancesRoot(balances map[string]string, user_balance_order []string, max_num_balances int) (string, bool) {
	balances_tree := &MerkleTree{}
	var balances_data = make([][]byte, max_num_balances)
	var wg sync.WaitGroup
	zero_hash := solsha3.SoliditySHA3(
		[]string{"address", "uint256", "uint256"},
		[]interface{}{
			"0x0000000000000000000000000000000000000000",
			"0",
			"0",
		},
	)
	for i := 0; i < max_num_balances; i++ {
		wg.Add(1)
		go func(i int) {
			if i < len(user_balance_order) && user_balance_order[i] != "0x0000000000000000000000000000000000000000" {
				amt_or_token_id := balances[user_balance_order[i]]
				currency_or_contract := user_balance_order[i]
				ctype := "0"
				if len(user_balance_order[i]) > 42 {
					amt_or_token_id = strings.Split(user_balance_order[i], "-")[1]
					currency_or_contract = strings.Split(user_balance_order[i], "-")[0]
					ctype = "1"
				}
				cb2, _ := new(big.Int).SetString(amt_or_token_id, 10)
				hash := solsha3.SoliditySHA3(
					[]string{"address", "uint256", "uint256"},
					[]interface{}{
						currency_or_contract,
						cb2,
						ctype,
					},
				)
				balances_data[i] = hash
			} else {
				balances_data[i] = zero_hash
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	balances_tree = NewMerkleTree(balances_data)
	return hex.EncodeToString(balances_tree.Root), true
}

type HasProcess struct {
	HasDeposit               bool
	HasWithdrawal            bool
	HasContractWithdrawal    bool
	HasNFTDeposit            bool
	HasNFTContractWithdrawal bool
}

func updateHasProcess(has_process *HasProcess, transaction Transaction) {
	switch transaction.Type {
	case "deposit":
		has_process.HasDeposit = true
	case "withdrawal":
		has_process.HasWithdrawal = true
	case "contract_withdrawal":
		has_process.HasContractWithdrawal = true
	case "nft_deposit":
		has_process.HasNFTDeposit = true
	case "nft_contract_withdrawal":
		has_process.HasNFTContractWithdrawal = true
	}
}

func verifyMintData(transaction Transaction, nft_collections_map map[string]map[string]interface{}) error {
	{
		message := solsha3.SoliditySHA3(
			[]string{"uint256", "address", "address", "uint256", "address", "uint256"},
			[]interface{}{strconv.Itoa(int(transaction.Nonce)), transaction.CurrencyOrNftContractAddress, transaction.To, transaction.MintFees, transaction.MintFeesToken, transaction.NumeFees},
		)
		if !EthVerify(hex.EncodeToString(message), transaction.Signature, transaction.To) {
			return fmt.Errorf("invalid mint signature")
		}
		if _, ok := nft_collections_map[transaction.CurrencyOrNftContractAddress]; !ok {
			return fmt.Errorf("nft collection not found")
		}
		mint_end_specifed, ok := new(big.Int).SetString(nft_collections_map[transaction.CurrencyOrNftContractAddress]["MintEnd"].(string), 10)
		if !ok {
			return fmt.Errorf("nft collection mint end is not valid")
		}
		mint_start_specifed, ok := new(big.Int).SetString(nft_collections_map[transaction.CurrencyOrNftContractAddress]["MintStart"].(string), 10)
		if !ok {
			return fmt.Errorf("nft collection mint end is not valid")
		}
		token_id, ok := new(big.Int).SetString(transaction.AmountOrNftTokenId, 10)
		if !ok {
			return fmt.Errorf("nft token id is not valid")
		}
		if mint_end_specifed.Cmp(big.NewInt(0)) == 1 && mint_end_specifed.Cmp(token_id) != 1 {
			return fmt.Errorf("nft collection token id should be less than mint end")
		}
		if mint_start_specifed.Cmp(token_id) != -1 {
			return fmt.Errorf("nft collection token id should be greater than mint start")
		}
		for _, v := range nft_collections_map[transaction.CurrencyOrNftContractAddress]["MintUsers"].([]interface{}) {
			if v.(string) == transaction.From {
				return fmt.Errorf("user does not have minting rights")
			}
		}
	}
	return nil
}

func binarySearch(array []uint, to_search uint) bool {
	found := false
	low := 0
	high := len(array) - 1
	for low <= high {
		mid := (low + high) / 2
		if array[mid] == to_search {
			found = true
			break
		}
		if array[mid] < to_search {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return found
}
