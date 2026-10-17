package api

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestPendingTxDetailsMethods(t *testing.T) {
	empty := PendingTxDetails{Nonce: 1}
	if empty.HasHash() {
		t.Error("expected empty.HasHash() to be false")
	}
	if empty.IsEnriched() {
		t.Error("expected empty.IsEnriched() to be false")
	}

	h := common.HexToHash("0x123")
	withHash := PendingTxDetails{Nonce: 2, Hash: &h}
	if !withHash.HasHash() {
		t.Error("expected withHash.HasHash() to be true")
	}
	if !withHash.IsEnriched() {
		t.Error("expected withHash.IsEnriched() to be true")
	}

	withFee := PendingTxDetails{Nonce: 3, MaxFeePerGas: big.NewInt(100)}
	if withFee.HasHash() {
		t.Error("expected withFee.HasHash() to be false")
	}
	if !withFee.IsEnriched() {
		t.Error("expected withFee.IsEnriched() to be true")
	}
}
