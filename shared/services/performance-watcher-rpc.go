package services

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	watcher "github.com/rocket-pool/smartnode/shared/services/performancewatcher"
)

// One budget covers live, backfill, registry refresh, and API revalidation.
type watcherBudget chan struct{}

func (b watcherBudget) acquire(ctx context.Context) (func(), error) {
	select {
	case b <- struct{}{}:
		return func() { <-b }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type watcherRequester struct {
	watcher.Requester
	budget watcherBudget
}

func (r watcherRequester) WatcherRequest(ctx context.Context, path string) ([]byte, error) {
	release, err := r.budget.acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer release()
	return r.Requester.WatcherRequest(ctx, path)
}

type watcherRPCClient struct {
	rocketpool.ExecutionClient
	fallback rocketpool.ExecutionClient
	budget   watcherBudget
	chainID  uint64
	verified *sync.Map
}

func watcherRead[T any](ctx context.Context, e watcherRPCClient, call func(context.Context, rocketpool.ExecutionClient) (T, error)) (T, error) {
	checked := func(ctx context.Context, client rocketpool.ExecutionClient) (T, error) {
		if e.chainID != 0 && e.verified != nil {
			if _, ok := e.verified.Load(client); !ok {
				chain, err := client.ChainID(ctx)
				if err != nil {
					var zero T
					return zero, err
				}
				if !chain.IsUint64() || chain.Uint64() != e.chainID {
					var zero T
					return zero, fmt.Errorf("configured execution client has a different chain ID")
				}
				e.verified.Store(client, true)
			}
		}
		return call(ctx, client)
	}
	first, cancel := context.WithTimeout(ctx, 15*time.Second)
	value, err := checked(first, e.ExecutionClient)
	cancel()
	if err != nil && e.fallback != nil && ctx.Err() == nil {
		return checked(ctx, e.fallback)
	}
	return value, err
}

func (e watcherRPCClient) scope(ctx context.Context) (context.Context, func(), error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	release, err := e.budget.acquire(ctx)
	if err != nil {
		cancel()
		return ctx, nil, err
	}
	return ctx, func() { release(); cancel() }, nil
}
func (e watcherRPCClient) CallContract(ctx context.Context, c ethereum.CallMsg, n *big.Int) ([]byte, error) {
	ctx, done, err := e.scope(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return watcherRead(ctx, e, func(ctx context.Context, client rocketpool.ExecutionClient) ([]byte, error) {
		return client.CallContract(ctx, c, n)
	})
}
func (e watcherRPCClient) CodeAt(ctx context.Context, a common.Address, n *big.Int) ([]byte, error) {
	ctx, done, err := e.scope(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return watcherRead(ctx, e, func(ctx context.Context, client rocketpool.ExecutionClient) ([]byte, error) {
		return client.CodeAt(ctx, a, n)
	})
}
func (e watcherRPCClient) HeaderByNumber(ctx context.Context, n *big.Int) (*ethtypes.Header, error) {
	ctx, done, err := e.scope(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return watcherRead(ctx, e, func(ctx context.Context, client rocketpool.ExecutionClient) (*ethtypes.Header, error) {
		return client.HeaderByNumber(ctx, n)
	})
}
func (e watcherRPCClient) SyncProgress(ctx context.Context) (*ethereum.SyncProgress, error) {
	ctx, done, err := e.scope(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return watcherRead(ctx, e, func(ctx context.Context, client rocketpool.ExecutionClient) (*ethereum.SyncProgress, error) {
		progress, err := client.SyncProgress(ctx)
		if err == nil && progress != nil {
			err = watcher.ErrSyncing
		}
		return progress, err
	})
}
func (e watcherRPCClient) ChainID(ctx context.Context) (*big.Int, error) {
	ctx, done, err := e.scope(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return watcherRead(ctx, e, func(ctx context.Context, client rocketpool.ExecutionClient) (*big.Int, error) {
		return client.ChainID(ctx)
	})
}
