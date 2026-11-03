package node

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

func TestParseHexUint64(t *testing.T) {
	if val := parseHexUint64("0x5208"); val != 21000 {
		t.Fatalf("expected 21000, got %d", val)
	}
	if val := parseHexUint64(nil); val != 0 {
		t.Fatalf("expected 0 for nil, got %d", val)
	}
	if val := parseHexUint64("invalid"); val != 0 {
		t.Fatalf("expected 0 for invalid, got %d", val)
	}
	if val := parseHexUint64("a"); val != 10 {
		t.Fatalf("expected 10, got %d", val)
	}
}

func TestParseHexBigInt(t *testing.T) {
	b := parseHexBigInt("0x10")
	if b == nil || b.Cmp(big.NewInt(16)) != 0 {
		t.Fatalf("expected 16, got %v", b)
	}

	if parseHexBigInt(nil) != nil {
		t.Fatalf("expected nil for nil input")
	}
	if parseHexBigInt("invalid_hex_string") != nil {
		t.Fatalf("expected nil for invalid string")
	}
}

func TestTxPoolContentResponse_FindTxByNonce(t *testing.T) {
	var nilResp *txPoolContentResponse
	if nilResp.findTxByNonce(5) != nil {
		t.Fatalf("expected nil for nil receiver")
	}

	pool := &txPoolContentResponse{
		Pending: map[string]map[string]interface{}{
			"10": {"hash": "0x1111111111111111111111111111111111111111111111111111111111111111"},
		},
		Queued: map[string]map[string]interface{}{
			"11": {"hash": "0x2222222222222222222222222222222222222222222222222222222222222222"},
		},
	}

	if pool.findTxByNonce(10) == nil {
		t.Fatalf("expected tx at nonce 10")
	}
	if pool.findTxByNonce(11) == nil {
		t.Fatalf("expected tx at nonce 11")
	}
	if pool.findTxByNonce(12) != nil {
		t.Fatalf("expected nil at nonce 12")
	}
}

func TestTxPoolContentResponse_GetTxGasFees(t *testing.T) {
	pool := &txPoolContentResponse{
		Pending: map[string]map[string]interface{}{
			"10": {
				"maxPriorityFeePerGas": "0x77359400", // 2 Gwei
				"maxFeePerGas":         "0xba43b7400", // 50 Gwei
			},
			"11": {
				"gasPrice": "0x4a817c800", // 20 Gwei legacy
			},
		},
		Queued: map[string]map[string]interface{}{},
	}

	tip, fee := pool.getTxGasFees(10)
	if tip == nil || tip.Cmp(big.NewInt(2000000000)) != 0 {
		t.Fatalf("expected tip 2000000000, got %v", tip)
	}
	if fee == nil || fee.Cmp(big.NewInt(50000000000)) != 0 {
		t.Fatalf("expected fee 50000000000, got %v", fee)
	}

	tipLegacy, feeLegacy := pool.getTxGasFees(11)
	if tipLegacy != nil {
		t.Fatalf("expected nil tip for legacy tx, got %v", tipLegacy)
	}
	if feeLegacy == nil || feeLegacy.Cmp(big.NewInt(20000000000)) != 0 {
		t.Fatalf("expected fee 20000000000, got %v", feeLegacy)
	}

	tipMissing, feeMissing := pool.getTxGasFees(999)
	if tipMissing != nil || feeMissing != nil {
		t.Fatalf("expected nil for missing nonce")
	}
}

func TestPopulatePendingTxDetails(t *testing.T) {
	item := &api.PendingTxDetails{Nonce: 42}
	populatePendingTxDetails(item, nil)
	if item.Hash != nil {
		t.Fatalf("expected nil hash when txMap is nil")
	}

	targetTo := common.HexToAddress("0x1111111111111111111111111111111111111111")
	txMap := map[string]interface{}{
		"hash":                 "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"to":                   targetTo.Hex(),
		"value":                "0xde0b6b3a7640000", // 1 ETH
		"gas":                  "0x5208",            // 21000
		"maxFeePerGas":         "0x2540be400",       // 10 Gwei
		"maxPriorityFeePerGas": "0x77359400",        // 2 Gwei
	}

	populatePendingTxDetails(item, txMap)

	if item.Hash == nil || item.Hash.Hex() != "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("unexpected hash: %v", item.Hash)
	}
	if item.To == nil || *item.To != targetTo {
		t.Fatalf("unexpected to: %v", item.To)
	}
	if item.Value == nil || item.Value.Cmp(big.NewInt(1000000000000000000)) != 0 {
		t.Fatalf("unexpected value: %v", item.Value)
	}
	if item.GasLimit != 21000 {
		t.Fatalf("expected gas 21000, got %d", item.GasLimit)
	}
	if item.MaxFeePerGas == nil || item.MaxFeePerGas.Cmp(big.NewInt(10000000000)) != 0 {
		t.Fatalf("unexpected max fee: %v", item.MaxFeePerGas)
	}
	if item.MaxPriorityFee == nil || item.MaxPriorityFee.Cmp(big.NewInt(2000000000)) != 0 {
		t.Fatalf("unexpected max priority fee: %v", item.MaxPriorityFee)
	}
	if !item.IsEnriched() {
		t.Fatalf("expected item.IsEnriched() to be true")
	}
}
