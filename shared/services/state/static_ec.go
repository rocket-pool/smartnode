package state

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

// Compile-time check that StaticExecutionClient satisfies rocketpool.ExecutionClient
var _ rocketpool.ExecutionClient = (*StaticExecutionClient)(nil)

// StaticExecutionClient serves execution-layer reads from a pre-loaded
// NetworkState snapshot. Read methods that can be answered from the snapshot
// return synthetic values; anything that would require a live eth_call
// (contract reads, transaction submission, receipts, logs, …) returns
// ErrStaticMode so that callers fail loudly instead of silently stalling on
// a non-existent client.
type StaticExecutionClient struct {
	state   *NetworkState
	chainID *big.Int
}

// NewStaticExecutionClient wires the given NetworkState + chain ID into a
// static ExecutionClient implementation.
func NewStaticExecutionClient(ns *NetworkState, chainID *big.Int) *StaticExecutionClient {
	if chainID == nil {
		chainID = big.NewInt(0)
	}
	return &StaticExecutionClient{state: ns, chainID: new(big.Int).Set(chainID)}
}

// blockTime returns the timestamp of the snapshot's execution block, derived
// from the beacon config and slot number. This is the best approximation we
// have without a live EL header.
func (c *StaticExecutionClient) blockTime() time.Time {
	cfg := c.state.BeaconConfig
	ts := cfg.GenesisTime + c.state.BeaconSlotNumber*cfg.SecondsPerSlot
	return time.Unix(int64(ts), 0)
}

// CodeAt returns a single placeholder byte for any address so that bind.go's
// "no code at address" detection passes. Real bytecode reads cannot be
// answered from the snapshot.
func (c *StaticExecutionClient) CodeAt(_ context.Context, _ common.Address, _ *big.Int) ([]byte, error) {
	return []byte{0x00}, nil
}

func (c *StaticExecutionClient) CallContract(_ context.Context, _ ethereum.CallMsg, _ *big.Int) ([]byte, error) {
	return nil, ErrStaticMode
}

func (c *StaticExecutionClient) HeaderByHash(_ context.Context, _ common.Hash) (*types.Header, error) {
	return nil, ErrStaticMode
}

// HeaderByNumber synthesizes a minimal header for the snapshot's EL block.
// Any other block number returns ErrStaticMode since the snapshot only covers
// a single point in time.
func (c *StaticExecutionClient) HeaderByNumber(_ context.Context, number *big.Int) (*types.Header, error) {
	snapshotBlock := new(big.Int).SetUint64(c.state.ElBlockNumber)
	if number != nil && number.Cmp(snapshotBlock) != 0 {
		return nil, ErrStaticMode
	}
	return &types.Header{
		Number: snapshotBlock,
		Time:   uint64(c.blockTime().Unix()),
	}, nil
}

func (c *StaticExecutionClient) PendingCodeAt(_ context.Context, _ common.Address) ([]byte, error) {
	return nil, ErrStaticMode
}

func (c *StaticExecutionClient) PendingNonceAt(_ context.Context, _ common.Address) (uint64, error) {
	return 0, ErrStaticMode
}

func (c *StaticExecutionClient) SuggestGasPrice(_ context.Context) (*big.Int, error) {
	return nil, ErrStaticMode
}

func (c *StaticExecutionClient) SuggestGasTipCap(_ context.Context) (*big.Int, error) {
	return nil, ErrStaticMode
}

func (c *StaticExecutionClient) EstimateGas(_ context.Context, _ ethereum.CallMsg) (uint64, error) {
	return 0, ErrStaticMode
}

func (c *StaticExecutionClient) SendTransaction(_ context.Context, _ *types.Transaction) error {
	return ErrStaticMode
}

func (c *StaticExecutionClient) FilterLogs(_ context.Context, _ ethereum.FilterQuery) ([]types.Log, error) {
	return nil, ErrStaticMode
}

func (c *StaticExecutionClient) SubscribeFilterLogs(_ context.Context, _ ethereum.FilterQuery, _ chan<- types.Log) (ethereum.Subscription, error) {
	return nil, ErrStaticMode
}

func (c *StaticExecutionClient) TransactionReceipt(_ context.Context, _ common.Hash) (*types.Receipt, error) {
	return nil, ErrStaticMode
}

func (c *StaticExecutionClient) BlockNumber(_ context.Context) (uint64, error) {
	return c.state.ElBlockNumber, nil
}

func (c *StaticExecutionClient) BalanceAt(_ context.Context, _ common.Address, _ *big.Int) (*big.Int, error) {
	return nil, ErrStaticMode
}

func (c *StaticExecutionClient) TransactionByHash(_ context.Context, _ common.Hash) (*types.Transaction, bool, error) {
	return nil, false, ErrStaticMode
}

func (c *StaticExecutionClient) NonceAt(_ context.Context, _ common.Address, _ *big.Int) (uint64, error) {
	return 0, ErrStaticMode
}

// SyncProgress returns nil to signal that the client is fully synced. The
// snapshot is, by definition, a single fixed point, so "not syncing" is the
// correct answer.
func (c *StaticExecutionClient) SyncProgress(_ context.Context) (*ethereum.SyncProgress, error) {
	return nil, nil
}

func (c *StaticExecutionClient) LatestBlockTime(_ context.Context) (time.Time, error) {
	return c.blockTime(), nil
}

func (c *StaticExecutionClient) ChainID(_ context.Context) (*big.Int, error) {
	return new(big.Int).Set(c.chainID), nil
}

// NetworkID mirrors ChainID and exists so that the static client can be used
// in places (like checkEcStatus) where the underlying ethclient.Client's
// NetworkID method is invoked directly.
func (c *StaticExecutionClient) NetworkID(_ context.Context) (*big.Int, error) {
	return new(big.Int).Set(c.chainID), nil
}
