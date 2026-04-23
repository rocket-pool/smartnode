package utils

import (
	"context"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

// Wait for a transaction to get included, respecting the provided context for cancellation.
// The transaction lookup retries indefinitely (with 1-second pauses) until found or ctx is done.
func WaitForTransactionWithContext(ctx context.Context, client rocketpool.ExecutionClient, hash common.Hash) (*types.Receipt, error) {
	var tx *types.Transaction

	// Get the transaction from its hash, retrying until found or ctx is cancelled.
	for {
		var err error
		tx, _, err = client.TransactionByHash(ctx, hash)
		if err == nil {
			break
		}
		if err.Error() != "not found" {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}

	// Wait for transaction to be mined
	txReceipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return nil, err
	}

	// Check transaction status
	if txReceipt.Status == 0 {
		return txReceipt, errors.New("Transaction failed with status 0")
	}

	return txReceipt, nil
}

// Wait for a transaction to get mined
func WaitForTransaction(client rocketpool.ExecutionClient, hash common.Hash) (*types.Receipt, error) {
	return WaitForTransactionWithContext(context.Background(), client, hash)
}
