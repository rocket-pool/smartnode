package node

import (
	"bytes"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

func TestTruncateHex(t *testing.T) {
	fullHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	truncated := truncateHex(fullHash, 6, 4)
	if truncated != "0x1234...cdef" {
		t.Fatalf("expected 0x1234...cdef, got %s", truncated)
	}

	short := "0x1234"
	if truncateHex(short, 6, 4) != short {
		t.Fatalf("expected %s to not be truncated", short)
	}
}

func TestFormatPendingTxRow(t *testing.T) {
	// Test sparse tx
	sparseTx := api.PendingTxDetails{
		Nonce:   5,
		IsStuck: false,
	}
	nonce, hashStr, toStr, valEth, maxFeeGwei, prioFeeGwei, status := formatPendingTxRow(sparseTx)
	if nonce != 5 {
		t.Errorf("expected nonce 5, got %d", nonce)
	}
	if hashStr != "<unknown>" {
		t.Errorf("expected <unknown>, got %s", hashStr)
	}
	if toStr != "<unknown>" {
		t.Errorf("expected <unknown>, got %s", toStr)
	}
	if valEth != "0.0000" {
		t.Errorf("expected 0.0000, got %s", valEth)
	}
	if maxFeeGwei != "-" {
		t.Errorf("expected '-', got %s", maxFeeGwei)
	}
	if prioFeeGwei != "-" {
		t.Errorf("expected '-', got %s", prioFeeGwei)
	}
	if status != "Pending" {
		t.Errorf("expected 'Pending', got %s", status)
	}

	// Test fully enriched tx
	h := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	to := common.HexToAddress("0xabcdef1234567890abcdef1234567890abcdef12")
	enrichedTx := api.PendingTxDetails{
		Nonce:          12,
		Hash:           &h,
		To:             &to,
		Value:          big.NewInt(1500000000000000000), // 1.5 ETH
		MaxFeePerGas:   big.NewInt(30000000000),         // 30 Gwei
		MaxPriorityFee: big.NewInt(2000000000),          // 2 Gwei
		IsStuck:        true,
	}
	nonce, hashStr, toStr, valEth, maxFeeGwei, prioFeeGwei, status = formatPendingTxRow(enrichedTx)
	if nonce != 12 {
		t.Errorf("expected nonce 12, got %d", nonce)
	}
	if hashStr != "0x1234...cdef" {
		t.Errorf("expected 0x1234...cdef, got %s", hashStr)
	}
	if toStr != truncateHex(to.Hex(), 6, 4) {
		t.Errorf("expected %s, got %s", truncateHex(to.Hex(), 6, 4), toStr)
	}
	if valEth != "1.5000" {
		t.Errorf("expected 1.5000, got %s", valEth)
	}
	if maxFeeGwei != "30.00" {
		t.Errorf("expected 30.00, got %s", maxFeeGwei)
	}
	if prioFeeGwei != "2.00" {
		t.Errorf("expected 2.00, got %s", prioFeeGwei)
	}
	if status != "Stuck" {
		t.Errorf("expected 'Stuck', got %s", status)
	}
}

func TestRenderPendingTransactionsTable(t *testing.T) {
	buf := new(bytes.Buffer)
	txs := []api.PendingTxDetails{
		{Nonce: 1, IsStuck: true},
		{Nonce: 2, IsStuck: true},
	}

	renderPendingTransactionsTable(buf, txs)
	output := buf.String()

	if !strings.Contains(output, "NONCE") || !strings.Contains(output, "HASH") {
		t.Errorf("table missing header: %s", output)
	}
	if !strings.Contains(output, "1") || !strings.Contains(output, "2") {
		t.Errorf("table missing rows: %s", output)
	}
}
