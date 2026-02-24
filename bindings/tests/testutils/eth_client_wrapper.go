package testutils

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthClientWrapper wraps *ethclient.Client to implement the ExecutionClient interface
type EthClientWrapper struct {
	*ethclient.Client
}

// NewEthClientWrapper creates a new wrapper around an ethclient.Client
func NewEthClientWrapper(client *ethclient.Client) *EthClientWrapper {
	return &EthClientWrapper{Client: client}
}

// LatestBlockTime returns the timestamp of the latest block
func (w *EthClientWrapper) LatestBlockTime(ctx context.Context) (time.Time, error) {
	header, err := w.Client.HeaderByNumber(ctx, nil)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(int64(header.Time), 0), nil
}

// CodeAt returns the code of the given account
func (w *EthClientWrapper) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	return w.Client.CodeAt(ctx, contract, blockNumber)
}

// CallContract executes an Ethereum contract call
func (w *EthClientWrapper) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return w.Client.CallContract(ctx, call, blockNumber)
}

// HeaderByHash returns the block header with the given hash
func (w *EthClientWrapper) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return w.Client.HeaderByHash(ctx, hash)
}

// HeaderByNumber returns a block header from the current canonical chain
func (w *EthClientWrapper) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return w.Client.HeaderByNumber(ctx, number)
}

// PendingCodeAt returns the code of the given account in the pending state
func (w *EthClientWrapper) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	return w.Client.PendingCodeAt(ctx, account)
}

// PendingNonceAt retrieves the current pending nonce associated with an account
func (w *EthClientWrapper) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return w.Client.PendingNonceAt(ctx, account)
}

// SuggestGasPrice retrieves the currently suggested gas price
func (w *EthClientWrapper) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return w.Client.SuggestGasPrice(ctx)
}

// SuggestGasTipCap retrieves the currently suggested 1559 priority fee
func (w *EthClientWrapper) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return w.Client.SuggestGasTipCap(ctx)
}

// EstimateGas tries to estimate the gas needed to execute a specific transaction
func (w *EthClientWrapper) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	return w.Client.EstimateGas(ctx, call)
}

// SendTransaction injects the transaction into the pending pool for execution
func (w *EthClientWrapper) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return w.Client.SendTransaction(ctx, tx)
}

// FilterLogs executes a log filter operation
func (w *EthClientWrapper) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	return w.Client.FilterLogs(ctx, query)
}

// SubscribeFilterLogs creates a background log filtering operation
func (w *EthClientWrapper) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	return w.Client.SubscribeFilterLogs(ctx, query, ch)
}

// TransactionReceipt returns the receipt of a transaction by transaction hash
func (w *EthClientWrapper) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return w.Client.TransactionReceipt(ctx, txHash)
}

// BlockNumber returns the most recent block number
func (w *EthClientWrapper) BlockNumber(ctx context.Context) (uint64, error) {
	return w.Client.BlockNumber(ctx)
}

// BalanceAt returns the wei balance of the given account
func (w *EthClientWrapper) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	return w.Client.BalanceAt(ctx, account, blockNumber)
}

// TransactionByHash returns the transaction with the given hash
func (w *EthClientWrapper) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	return w.Client.TransactionByHash(ctx, hash)
}

// NonceAt returns the account nonce of the given account
func (w *EthClientWrapper) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	return w.Client.NonceAt(ctx, account, blockNumber)
}

// SyncProgress retrieves the current progress of the sync algorithm
func (w *EthClientWrapper) SyncProgress(ctx context.Context) (*ethereum.SyncProgress, error) {
	return w.Client.SyncProgress(ctx)
}

// ChainID retrieves the current chain ID
func (w *EthClientWrapper) ChainID(ctx context.Context) (*big.Int, error) {
	return w.Client.ChainID(ctx)
}
