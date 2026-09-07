package node

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func TestParseHexHelpers(t *testing.T) {
	// parseHexUint64 tests
	if got := parseHexUint64("0x1a"); got != 26 {
		t.Errorf("expected 26, got %d", got)
	}
	if got := parseHexUint64("0x5208"); got != 21000 {
		t.Errorf("expected 21000, got %d", got)
	}
	if got := parseHexUint64("0x0"); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
	if got := parseHexUint64(123); got != 0 {
		t.Errorf("expected 0 for non-string, got %d", got)
	}

	// parseHexBigInt tests
	val := parseHexBigInt("0xde0b6b3a7640000") // 1 ETH in wei
	expectedVal, _ := new(big.Int).SetString("1000000000000000000", 10)
	if val == nil || val.Cmp(expectedVal) != 0 {
		t.Errorf("expected 1 ETH (10^18 wei), got %v", val)
	}

	zeroVal := parseHexBigInt("0x0")
	if zeroVal == nil || zeroVal.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected 0, got %v", zeroVal)
	}

	if got := parseHexBigInt(123); got != nil {
		t.Errorf("expected nil for non-string, got %v", got)
	}
}

func TestEnrichPendingTxDetails(t *testing.T) {
	item := api.PendingTxDetails{
		Nonce:   10,
		IsStuck: true,
	}

	expectedHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	expectedTo := "0x1111111111111111111111111111111111111111"
	txMap := map[string]interface{}{
		"hash":                  expectedHash,
		"to":                    expectedTo,
		"value":                 "0xde0b6b3a7640000", // 1 ETH
		"gas":                   "0x5208",            // 21000
		"maxFeePerGas":          "0x4a817c800",       // 20 Gwei
		"maxPriorityFeePerGas":  "0x77359400",        // 2 Gwei
	}

	enrichPendingTxDetails(&item, txMap)

	if item.Hash == nil || item.Hash.Hex() != common.HexToHash(expectedHash).Hex() {
		t.Errorf("hash mismatch: expected %s, got %v", expectedHash, item.Hash)
	}
	if item.To == nil || item.To.Hex() != common.HexToAddress(expectedTo).Hex() {
		t.Errorf("to address mismatch: expected %s, got %v", expectedTo, item.To)
	}
	if item.GasLimit != 21000 {
		t.Errorf("expected gas limit 21000, got %d", item.GasLimit)
	}
	if item.MaxFeePerGas == nil || item.MaxFeePerGas.Cmp(big.NewInt(20000000000)) != 0 {
		t.Errorf("maxFeePerGas mismatch: expected 20 Gwei, got %v", item.MaxFeePerGas)
	}
	if item.MaxPriorityFee == nil || item.MaxPriorityFee.Cmp(big.NewInt(2000000000)) != 0 {
		t.Errorf("maxPriorityFee mismatch: expected 2 Gwei, got %v", item.MaxPriorityFee)
	}
}
