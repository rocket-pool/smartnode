package megapool

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

const prestakeWatcherABI = `[
	{
		"name": "getPrestakedValidatorIndices",
		"type": "function",
		"stateMutability": "view",
		"inputs": [
			{"name": "_offset", "type": "uint256"},
			{"name": "_limit", "type": "uint256"}
		],
		"outputs": [
			{"name": "indices", "type": "uint256[]"},
			{"name": "nextIndex", "type": "uint256"},
			{"name": "failedIndices", "type": "uint256[]"}
		]
	},
	{
		"name": "getValidatorPrestakeInfo",
		"type": "function",
		"stateMutability": "view",
		"inputs": [
			{"name": "_index", "type": "uint256"}
		],
		"outputs": [
			{"name": "pubkey", "type": "bytes"},
			{"name": "validatorInfo", "type": "tuple", "components": [
				{"name": "lastAssignmentTime", "type": "uint32"},
				{"name": "lastRequestedValue", "type": "uint32"},
				{"name": "lastRequestedBond", "type": "uint32"},
				{"name": "depositValue", "type": "uint32"},
				{"name": "staked", "type": "bool"},
				{"name": "exited", "type": "bool"},
				{"name": "inQueue", "type": "bool"},
				{"name": "inPrestake", "type": "bool"},
				{"name": "expressUsed", "type": "bool"},
				{"name": "dissolved", "type": "bool"},
				{"name": "exiting", "type": "bool"},
				{"name": "locked", "type": "bool"},
				{"name": "exitBalance", "type": "uint64"},
				{"name": "lockedTime", "type": "uint64"}
			]},
			{"name": "megapool", "type": "address"},
			{"name": "validatorId", "type": "uint32"},
			{"name": "signature", "type": "bytes"},
			{"name": "challengeDeadline", "type": "uint256"}
		]
	}
]`

func TestUnpackPrestakedValidatorIndices(t *testing.T) {
	parsed := mustPrestakeABI(t)
	indices := []*big.Int{big.NewInt(4), big.NewInt(9)}
	next := big.NewInt(12)
	failed := []*big.Int{big.NewInt(7)}
	packed, err := parsed.Methods["getPrestakedValidatorIndices"].Outputs.Pack(indices, next, failed)
	if err != nil {
		t.Fatal(err)
	}
	page, err := unpackPrestakedValidatorIndices(&parsed, packed)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Indices) != 2 || page.Indices[0].Cmp(indices[0]) != 0 || page.Indices[1].Cmp(indices[1]) != 0 {
		t.Fatalf("indices = %v", page.Indices)
	}
	if page.NextIndex.Cmp(next) != 0 {
		t.Fatalf("nextIndex = %s", page.NextIndex)
	}
	if len(page.FailedIndices) != 1 || page.FailedIndices[0].Cmp(failed[0]) != 0 {
		t.Fatalf("failedIndices = %v", page.FailedIndices)
	}
}

func TestUnpackValidatorPrestakeInfo(t *testing.T) {
	parsed := mustPrestakeABI(t)
	pubkey := make([]byte, 48)
	pubkey[0] = 0xab
	signature := make([]byte, 96)
	signature[0] = 0xcd
	megapoolAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
	deadline := big.NewInt(43200)
	info := ValidatorInfo{
		LastAssignmentTime: 10,
		LastRequestedValue: 32000,
		LastRequestedBond:  4000,
		DepositValue:       1000,
		InPrestake:         true,
		ExitBalance:        32,
	}
	packed, err := parsed.Methods["getValidatorPrestakeInfo"].Outputs.Pack(
		pubkey, info, megapoolAddress, uint32(3), signature, deadline,
	)
	if err != nil {
		t.Fatal(err)
	}
	got, err := unpackValidatorPrestakeInfo(&parsed, packed)
	if err != nil {
		t.Fatal(err)
	}
	if got.Pubkey[0] != 0xab || got.Signature[0] != 0xcd {
		t.Fatalf("pubkey/signature not preserved: %x %x", got.Pubkey, got.Signature)
	}
	if got.MegapoolAddress != megapoolAddress || got.ValidatorId != 3 {
		t.Fatalf("identity = %s %d", got.MegapoolAddress.Hex(), got.ValidatorId)
	}
	if !got.ValidatorInfo.InPrestake || got.ValidatorInfo.LastAssignmentTime != 10 || got.ValidatorInfo.LastRequestedBond != 4000 {
		t.Fatalf("validator info = %+v", got.ValidatorInfo)
	}
	if got.ChallengeDeadline.Cmp(deadline) != 0 {
		t.Fatalf("deadline = %s", got.ChallengeDeadline)
	}
}

func mustPrestakeABI(t *testing.T) abi.ABI {
	t.Helper()
	parsed, err := abi.JSON(strings.NewReader(prestakeWatcherABI))
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
