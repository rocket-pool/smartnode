package client

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// This type wraps multiple ETH clients, providing natural fallback support if one of them fails.
type EthClientProxy struct {
    clientUrls []string
    clients []*ethclient.Client
    timeouts []time.Time
    reconnectDelay time.Duration
}


// This is a signature for a wrapped ethclient.Client function 
type clientFunction func(*ethclient.Client) (interface{}, error)


// Creates a new Eth1ClientProxy instance based on the main and backup client URLs
func NewEth1ClientProxy(reconnectDelay time.Duration, urls ...string) (*EthClientProxy) {

    clients := []*ethclient.Client{}
    timeouts := []time.Time{}

    // Clamp the delay
    if reconnectDelay < 0 {
        reconnectDelay = 0
    }

    // Try connecting to each client, but ignore errors - they'll be handled at runtime
    for _, url := range urls {
        client, err := ethclient.Dial(url)
        if err != nil {
            timeouts = append(timeouts, time.Now())
        } else {
            timeouts = append(timeouts, time.Time{})
        }
        clients = append(clients, client)
    }

    return &EthClientProxy{
        clientUrls: urls,
        clients: clients,
        timeouts: timeouts,
        reconnectDelay: reconnectDelay,
    }

}


/// ========================
/// ContractCaller Functions
/// ========================


// CodeAt returns the code of the given account. This is needed to differentiate
// between contract internal errors and the local chain being out of sync.
func (p *EthClientProxy) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
    result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
        return client.CodeAt(ctx, contract, blockNumber)
    })
    if err != nil {
        return nil, err
    }
    return result.([]byte), err
}


// CallContract executes an Ethereum contract call with the specified data as the
// input.
func (p *EthClientProxy) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
    result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
        return client.CallContract(ctx, call, blockNumber)
    })
    if err != nil {
        return nil, err
    }
    return result.([]byte), err
}


/// ============================
/// ContractTransactor Functions
/// ============================


// HeaderByNumber returns a block header from the current canonical chain. If number is
// nil, the latest known header is returned.
func (p *EthClientProxy) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
    result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
        return client.HeaderByNumber(ctx, number)
    })
    if err != nil {
        return nil, err
    }
    return result.(*types.Header), err
}


// PendingCodeAt returns the code of the given account in the pending state.
func (p *EthClientProxy) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
    result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
        return client.PendingCodeAt(ctx, account)
    })
    if err != nil {
        return nil, err
    }
    return result.([]byte), err
}


// PendingNonceAt retrieves the current pending nonce associated with an account.
func (p *EthClientProxy) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
    result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
        return client.PendingNonceAt(ctx, account)
    })
    if err != nil {
        return 0, err
    }
    return result.(uint64), err
}


// SuggestGasPrice retrieves the currently suggested gas price to allow a timely
// execution of a transaction.
func (p *EthClientProxy) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
    result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
        return client.SuggestGasPrice(ctx)
    })
    if err != nil {
        return nil, err
    }
    return result.(*big.Int), err
}


// SuggestGasTipCap retrieves the currently suggested 1559 priority fee to allow
// a timely execution of a transaction.
func (p *EthClientProxy) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
    result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
        return client.SuggestGasTipCap(ctx)
    })
    if err != nil {
        return nil, err
    }
    return result.(*big.Int), err
}


// EstimateGas tries to estimate the gas needed to execute a specific
// transaction based on the current pending state of the backend blockchain.
// There is no guarantee that this is the true gas limit requirement as other
// transactions may be added or removed by miners, but it should provide a basis
// for setting a reasonable default.
func (p *EthClientProxy) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
    result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
        return client.EstimateGas(ctx, call)
    })
    if err != nil {
        return 0, err
    }
    return result.(uint64), err
}


// SendTransaction injects the transaction into the pending pool for execution.
func (p *EthClientProxy) SendTransaction(ctx context.Context, tx *types.Transaction) error {
    _, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
        return nil, client.SendTransaction(ctx, tx)
    })
    return err
}


/// ==========================
/// ContractFilterer Functions
/// ==========================


// FilterLogs executes a log filter operation, blocking during execution and
// returning all the results in one batch.
//
// TODO(karalabe): Deprecate when the subscription one can return past data too.
func (p *EthClientProxy) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
    result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
        return client.FilterLogs(ctx, query)
    })
    if err != nil {
        return nil, err
    }
    return result.([]types.Log), err
}


// SubscribeFilterLogs creates a background log filtering operation, returning
// a subscription immediately, which can be used to stream the found events.
func (p *EthClientProxy) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
    result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
        return client.SubscribeFilterLogs(ctx, query, ch)
    })
    if err != nil {
        return nil, err
    }
    return result.(ethereum.Subscription), err
}


/// =======================
/// DeployBackend Functions
/// =======================


// TransactionReceipt returns the receipt of a transaction by transaction hash.
// Note that the receipt is not available for pending transactions.
func (p *EthClientProxy) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
    result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
        return client.TransactionReceipt(ctx, txHash)
    })
    if err != nil {
        return nil, err
    }
    return result.(*types.Receipt), err
}


/// ================
/// Client functions
/// ================


// BlockNumber returns the most recent block number
func (p *EthClientProxy) BlockNumber(ctx context.Context) (uint64, error) {
    result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
        return client.BlockNumber(ctx)
    })
    if err != nil {
        return 0, err
    }
    return result.(uint64), err
}


// BalanceAt returns the wei balance of the given account.
// The block number can be nil, in which case the balance is taken from the latest known block.
func (p *EthClientProxy) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
    result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
        return client.BalanceAt(ctx, account, blockNumber)
    })
    if err != nil {
        return nil, err
    }
    return result.(*big.Int), err
}


// TransactionByHash returns the transaction with the given hash.
func (p *EthClientProxy) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
    result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
        tx, isPending, err := client.TransactionByHash(ctx, hash)
        result := []interface{} { tx, isPending }
        return result, err
    })
    if err != nil {
        return nil, false, err
    }

    // TODO: Can we just use the named return values inside the closer to skip this?
    resultArray := result.([]interface{})
    tx = resultArray[0].(*types.Transaction)
    isPending = resultArray[1].(bool)
    return tx, isPending, err
}


// NonceAt returns the account nonce of the given account.
// The block number can be nil, in which case the nonce is taken from the latest known block.
func (p *EthClientProxy) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
    result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
        return client.NonceAt(ctx, account, blockNumber)
    })
    if err != nil {
        return 0, err
    }
    return result.(uint64), err
}


// SyncProgress retrieves the current progress of the sync algorithm. If there's
// no sync currently running, it returns nil.
func (p *EthClientProxy) SyncProgress(ctx context.Context) (*ethereum.SyncProgress, error) {
    result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
        return client.SyncProgress(ctx)
    })
    if err != nil {
        return nil, err
    }
    return result.(*ethereum.SyncProgress), err
}


/// ==================
/// Internal functions
/// ==================


// Attempts to run a function progressively through each client until one succeeds or they all fail.
func (p *EthClientProxy) runFunction(function clientFunction) (interface{}, error) {

    // A cumulative error string as each client gets tried
    errorString := ""

    for i := 0; i < len(p.clients); i++ {
        client, clientErr := p.getClient(i)
        if client != nil {

            // This client is available, try running the function
            result, err := function(client)
            if err != nil {

                // If it's disconnected, log it and try the next client
                errorString += fmt.Sprintf("\nError with client %d: %s", i, err.Error())
                if isDisconnected(err) {
                    p.clients[i] = nil
                    p.timeouts[i] = time.Now()

                // If it's a different error, just return it
                } else {
                    return nil, fmt.Errorf(errorString)
                }

            // If there's no error, return the result
            } else {
                return result, nil
            }

        // Note a client failure and try the next one
        } else {
            errorString += fmt.Sprintf("\nError with client %d: %s", i, clientErr.Error())
        }
    }
    
    // If none of the clients worked, return the aggregated error string
    errorString += "\nNone of the clients were available."
    return nil, fmt.Errorf(errorString)

}


// Returns true if the error was a connection failure and a backup client is available
func isDisconnected(err error) (bool) {
    return strings.Contains(err.Error(), "dial tcp")
}


// Get the client at the given index, trying a reconnect if it's disconnected
func (p *EthClientProxy) getClient(index int) (*ethclient.Client, error) {
    // Try connecting to the client if it's dead
    var err error
    if p.clients[index] == nil {

        // Check if enough time has passed
        if time.Since(p.timeouts[index]) > p.reconnectDelay {
            p.clients[index], err = ethclient.Dial(p.clientUrls[index])
        
            // If the connection failed, reset the timer
            if err != nil {
                p.timeouts[index] = time.Now()
            }

        } else {
            err = fmt.Errorf("Connection failure, waiting %s to reconnect", p.reconnectDelay)
        }
    }

    // Return the client regardless of its state
    return p.clients[index], err
}

