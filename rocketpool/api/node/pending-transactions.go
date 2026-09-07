package node

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type txPoolContentFromResponse struct {
	Pending map[string]map[string]interface{} `json:"pending"`
	Queued  map[string]map[string]interface{} `json:"queued"`
}

func nodePendingTransactions(c *cli.Command) (*api.NodePendingTransactionsResponse, error) {
	// Require node wallet
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}

	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	nodeAddress := nodeAccount.Address

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1. Get latest mined nonce and pending mempool nonce
	latestNonce, err := ec.NonceAt(ctx, nodeAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting latest on-chain nonce: %w", err)
	}

	pendingNonce, err := ec.PendingNonceAt(ctx, nodeAddress)
	if err != nil {
		return nil, fmt.Errorf("error getting pending nonce: %w", err)
	}

	resp := api.NodePendingTransactionsResponse{
		NodeAddress:         nodeAddress,
		LatestNonce:         latestNonce,
		PendingNonce:        pendingNonce,
		PendingTransactions: make([]api.PendingTxDetails, 0),
	}

	if pendingNonce > latestNonce {
		resp.PendingCount = pendingNonce - latestNonce

		// 2. Best-effort mempool query via txpool_contentFrom
		var poolContent txPoolContentFromResponse
		_ = ec.RawCallContext(ctx, &poolContent, "txpool_contentFrom", nodeAddress.Hex())

		for nonce := latestNonce; nonce < pendingNonce; nonce++ {
			nonceStr := strconv.FormatUint(nonce, 10)
			item := api.PendingTxDetails{
				Nonce:   nonce,
				IsStuck: true,
			}

			// Check if enriched details exist in txpool output
			if txMap, ok := poolContent.Pending[nonceStr]; ok {
				enrichPendingTxDetails(&item, txMap)
			} else if txMap, ok := poolContent.Queued[nonceStr]; ok {
				enrichPendingTxDetails(&item, txMap)
			}

			resp.PendingTransactions = append(resp.PendingTransactions, item)
		}
	}

	return &resp, nil
}

func enrichPendingTxDetails(item *api.PendingTxDetails, txMap map[string]interface{}) {
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

func parseHexUint64(v interface{}) uint64 {
	if s, ok := v.(string); ok {
		s = strings.TrimPrefix(s, "0x")
		n, _ := strconv.ParseUint(s, 16, 64)
		return n
	}
	return 0
}

func parseHexBigInt(v interface{}) *big.Int {
	if s, ok := v.(string); ok {
		s = strings.TrimPrefix(s, "0x")
		b, _ := new(big.Int).SetString(s, 16)
		return b
	}
	return nil
}

func pendingTransactionsHandler(ctx snroute.Context) {
	resp, err := nodePendingTransactions(ctx.Command())
	response.WriteResponse(ctx.Writer, resp, err)
}
