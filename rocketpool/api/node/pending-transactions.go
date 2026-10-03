package node

import (
	"context"
	"fmt"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

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
		poolContent := fetchNodeMempool(ctx, ec, nodeAddress)

		for nonce := latestNonce; nonce < pendingNonce; nonce++ {
			item := api.PendingTxDetails{
				Nonce:   nonce,
				IsStuck: true,
			}
			populatePendingTxDetails(&item, poolContent.findTxByNonce(nonce))
			resp.PendingTransactions = append(resp.PendingTransactions, item)
		}
	}

	return &resp, nil
}

func pendingTransactionsHandler(ctx snroute.Context) {
	resp, err := nodePendingTransactions(ctx.Command())
	response.WriteResponse(ctx.Writer, resp, err)
}
