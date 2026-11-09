package utils

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

type missingTransactionClient struct {
	rocketpool.ExecutionClient
	err error
}

func (c missingTransactionClient) TransactionByHash(context.Context, common.Hash) (*types.Transaction, bool, error) {
	return nil, false, c.err
}

func TestWaitForTransactionWrappedNotFound(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := missingTransactionClient{err: fmt.Errorf("lookup: %w", ethereum.NotFound)}
	_, err := WaitForTransactionWithContext(ctx, client, common.Hash{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation while waiting for a missing transaction, got %v", err)
	}
}
