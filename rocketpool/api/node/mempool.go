package node

import (
	"context"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// txPoolContentResponse represents the payload returned by txpool_contentFrom
type txPoolContentResponse struct {
	Pending map[string]map[string]interface{} `json:"pending"`
	Queued  map[string]map[string]interface{} `json:"queued"`
}

// fetchNodeMempool performs a best-effort txpool_contentFrom query against the execution client.
func fetchNodeMempool(ctx context.Context, ec *services.ExecutionClientManager, nodeAddress common.Address) *txPoolContentResponse {
	var poolContent txPoolContentResponse
	_ = ec.RawCallContext(ctx, &poolContent, "txpool_contentFrom", nodeAddress.Hex())
	return &poolContent
}

// findTxByNonce looks for a transaction at the given nonce in pending or queued sets.
func (p *txPoolContentResponse) findTxByNonce(nonce uint64) map[string]interface{} {
	if p == nil {
		return nil
	}
	nonceStr := strconv.FormatUint(nonce, 10)
	if txMap, ok := p.Pending[nonceStr]; ok {
		return txMap
	}
	if txMap, ok := p.Queued[nonceStr]; ok {
		return txMap
	}
	return nil
}

// getTxGasFees extracts the max priority fee and max fee (or gasPrice) for the given nonce if known.
func (p *txPoolContentResponse) getTxGasFees(nonce uint64) (oldTip *big.Int, oldFee *big.Int) {
	txMap := p.findTxByNonce(nonce)
	if txMap == nil {
		return nil, nil
	}
	if tip := parseHexBigInt(txMap["maxPriorityFeePerGas"]); tip != nil {
		oldTip = tip
	}
	if fee := parseHexBigInt(txMap["maxFeePerGas"]); fee != nil {
		oldFee = fee
	} else if gp := parseHexBigInt(txMap["gasPrice"]); gp != nil {
		oldFee = gp
	}
	return oldTip, oldFee
}

// populatePendingTxDetails fills in item fields from txMap if available.
func populatePendingTxDetails(item *api.PendingTxDetails, txMap map[string]interface{}) {
	if txMap == nil {
		return
	}
	if hStr, ok := txMap["hash"].(string); ok && hStr != "" {
		h := common.HexToHash(hStr)
		item.Hash = &h
	}
	if toStr, ok := txMap["to"].(string); ok && toStr != "" {
		to := common.HexToAddress(toStr)
		item.To = &to
	}
	if val := parseHexBigInt(txMap["value"]); val != nil {
		item.Value = val
	}
	if gas := parseHexUint64(txMap["gas"]); gas > 0 {
		item.GasLimit = gas
	}
	if maxFee := parseHexBigInt(txMap["maxFeePerGas"]); maxFee != nil {
		item.MaxFeePerGas = maxFee
	} else if gasPrice := parseHexBigInt(txMap["gasPrice"]); gasPrice != nil {
		item.MaxFeePerGas = gasPrice
	}
	if maxPrio := parseHexBigInt(txMap["maxPriorityFeePerGas"]); maxPrio != nil {
		item.MaxPriorityFee = maxPrio
	}
}

// parseHexUint64 parses a hex string like "0x5208" or scalar to uint64.
func parseHexUint64(v interface{}) uint64 {
	if s, ok := v.(string); ok {
		s = strings.TrimPrefix(s, "0x")
		n, _ := strconv.ParseUint(s, 16, 64)
		return n
	}
	return 0
}

// parseHexBigInt parses a hex string like "0x123" to *big.Int.
func parseHexBigInt(v interface{}) *big.Int {
	if s, ok := v.(string); ok {
		s = strings.TrimPrefix(s, "0x")
		b, _ := new(big.Int).SetString(s, 16)
		return b
	}
	return nil
}
