package services

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// This is a proxy for multiple ETH clients, providing natural fallback support if one of them fails.
type ExecutionClientManager struct {
	primaryEcUrl    string
	fallbackEcUrl   string
	primaryEc       *ethclient.Client
	fallbackEc      *ethclient.Client
	logger          log.ColorLogger
	primaryReady    bool
	fallbackReady   bool
	ignoreSyncCheck bool
}

// This is a signature for a wrapped ethclient.Client function
type ecFunction func(*ethclient.Client) (interface{}, error)

// Creates a new ExecutionClientManager instance based on the Rocket Pool config
func NewExecutionClientManager(cfg *config.RocketPoolConfig) (*ExecutionClientManager, error) {

	var primaryEcUrl string
	var fallbackEcUrl string

	// Get the primary EC url
	if cfg.IsNativeMode {
		primaryEcUrl = cfg.Native.EcHttpUrl.Value.(string)
	} else if cfg.ExecutionClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
		primaryEcUrl = fmt.Sprintf("http://%s:%d", config.Eth1ContainerName, cfg.ExecutionCommon.HttpPort.Value)
	} else {
		primaryEcUrl = cfg.ExternalExecution.HttpUrl.Value.(string)
	}

	// Get the fallback EC url, if applicable
	if cfg.UseFallbackClients.Value == true {
		if cfg.IsNativeMode {
			fallbackEcUrl = cfg.FallbackNormal.EcHttpUrl.Value.(string)
		} else {
			cc, _ := cfg.GetSelectedConsensusClient()
			switch cc {
			case cfgtypes.ConsensusClient_Prysm:
				fallbackEcUrl = cfg.FallbackPrysm.EcHttpUrl.Value.(string)
			default:
				fallbackEcUrl = cfg.FallbackNormal.EcHttpUrl.Value.(string)
			}
		}
	}

	primaryEc, err := ethclient.Dial(primaryEcUrl)
	if err != nil {
		return nil, fmt.Errorf("error connecting to primary EC at [%s]: %w", primaryEcUrl, err)
	}

	var fallbackEc *ethclient.Client
	if fallbackEcUrl != "" {
		fallbackEc, err = ethclient.Dial(fallbackEcUrl)
		if err != nil {
			return nil, fmt.Errorf("error connecting to fallback EC at [%s]: %w", fallbackEcUrl, err)
		}
	}

	return &ExecutionClientManager{
		primaryEcUrl:  primaryEcUrl,
		fallbackEcUrl: fallbackEcUrl,
		primaryEc:     primaryEc,
		fallbackEc:    fallbackEc,
		logger:        log.NewColorLogger(color.FgYellow),
		primaryReady:  true,
		fallbackReady: fallbackEc != nil,
	}, nil

}

/// ========================
/// ContractCaller Functions
/// ========================

// CodeAt returns the code of the given account. This is needed to differentiate
// between contract internal errors and the local chain being out of sync.
func (p *ExecutionClientManager) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
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
func (p *ExecutionClientManager) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
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

// HeaderByHash returns the block header with the given hash.
func (p *ExecutionClientManager) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
		return client.HeaderByHash(ctx, hash)
	})
	if err != nil {
		return nil, err
	}
	return result.(*types.Header), err
}

// HeaderByNumber returns a block header from the current canonical chain. If number is
// nil, the latest known header is returned.
func (p *ExecutionClientManager) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
		return client.HeaderByNumber(ctx, number)
	})
	if err != nil {
		return nil, err
	}
	return result.(*types.Header), err
}

// PendingCodeAt returns the code of the given account in the pending state.
func (p *ExecutionClientManager) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
		return client.PendingCodeAt(ctx, account)
	})
	if err != nil {
		return nil, err
	}
	return result.([]byte), err
}

// PendingNonceAt retrieves the current pending nonce associated with an account.
func (p *ExecutionClientManager) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
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
func (p *ExecutionClientManager) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
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
func (p *ExecutionClientManager) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
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
func (p *ExecutionClientManager) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
		return client.EstimateGas(ctx, call)
	})
	if err != nil {
		return 0, err
	}
	return result.(uint64), err
}

// SendTransaction injects the transaction into the pending pool for execution.
func (p *ExecutionClientManager) SendTransaction(ctx context.Context, tx *types.Transaction) error {
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
func (p *ExecutionClientManager) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
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
func (p *ExecutionClientManager) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
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
func (p *ExecutionClientManager) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
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
func (p *ExecutionClientManager) BlockNumber(ctx context.Context) (uint64, error) {
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
func (p *ExecutionClientManager) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
		return client.BalanceAt(ctx, account, blockNumber)
	})
	if err != nil {
		return nil, err
	}
	return result.(*big.Int), err
}

// TransactionByHash returns the transaction with the given hash.
func (p *ExecutionClientManager) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	result, err := p.runFunction(func(client *ethclient.Client) (interface{}, error) {
		tx, isPending, err := client.TransactionByHash(ctx, hash)
		result := []interface{}{tx, isPending}
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
func (p *ExecutionClientManager) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
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
func (p *ExecutionClientManager) SyncProgress(ctx context.Context) (*ethereum.SyncProgress, error) {
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

func (p *ExecutionClientManager) CheckStatus(cfg *config.RocketPoolConfig) *api.ClientManagerStatus {

	status := &api.ClientManagerStatus{
		FallbackEnabled: p.fallbackEc != nil,
	}

	// Ignore the sync check and just use the predefined settings if requested
	if p.ignoreSyncCheck {
		status.PrimaryClientStatus.IsWorking = p.primaryReady
		status.PrimaryClientStatus.IsSynced = p.primaryReady
		if status.FallbackEnabled {
			status.FallbackClientStatus.IsWorking = p.fallbackReady
			status.FallbackClientStatus.IsSynced = p.fallbackReady
		}
		return status
	}

	// Get the primary EC status
	status.PrimaryClientStatus = checkEcStatus(p.primaryEc)

	// Flag if primary client is ready
	p.primaryReady = (status.PrimaryClientStatus.IsWorking && status.PrimaryClientStatus.IsSynced)

	// Get the fallback EC status if applicable
	if status.FallbackEnabled {
		status.FallbackClientStatus = checkEcStatus(p.fallbackEc)
		// Check if fallback is using the expected network
		expectedChainID := cfg.Smartnode.GetChainID()
		if status.FallbackClientStatus.Error == "" && status.FallbackClientStatus.NetworkId != expectedChainID {
			p.fallbackReady = false
			colorReset := "\033[0m"
			colorYellow := "\033[33m"
			status.FallbackClientStatus.Error = fmt.Sprintf("The fallback client is using a different chain [%s%s%s, Chain ID %d] than what your node is configured for [%s, Chain ID %d]", colorYellow, getNetworkNameFromId(status.FallbackClientStatus.NetworkId), colorReset, status.FallbackClientStatus.NetworkId, getNetworkNameFromId(expectedChainID), expectedChainID)
			return status
		}
	}

	p.fallbackReady = (status.FallbackEnabled && status.FallbackClientStatus.IsWorking && status.FallbackClientStatus.IsSynced)

	return status
}

func getNetworkNameFromId(networkId uint) string {
	switch networkId {
	case 1:
		return "Ethereum Mainnet"
	case 5:
		return "Goerli Testnet"
	case 17000:
		return "Holesky Testnet"
	default:
		return "Unknown Network"
	}

}

// Check the client status
func checkEcStatus(client *ethclient.Client) api.ClientStatus {

	status := api.ClientStatus{}

	// Get the NetworkId
	networkId, err := client.NetworkID(context.Background())
	if err != nil {
		status.Error = fmt.Sprintf("Sync progress check failed with [%s]", err.Error())
		status.IsSynced = false
		status.IsWorking = false
		return status
	}

	if networkId != nil {
		status.NetworkId = uint(networkId.Uint64())
	}

	// Get the fallback's sync progress
	progress, err := client.SyncProgress(context.Background())
	if err != nil {
		status.Error = fmt.Sprintf("Sync progress check failed with [%s]", err.Error())
		status.IsSynced = false
		status.IsWorking = false
		return status
	}

	// Make sure it's up to date
	if progress == nil {

		isUpToDate, blockTime, err := IsSyncWithinThreshold(client)
		if err != nil {
			status.Error = fmt.Sprintf("Error checking if client's sync progress is up to date: [%s]", err.Error())
			status.IsSynced = false
			status.IsWorking = false
			return status
		}

		status.IsWorking = true
		if !isUpToDate {
			status.Error = fmt.Sprintf("Client claims to have finished syncing, but its last block was from %s ago. It likely doesn't have enough peers", time.Since(blockTime))
			status.IsSynced = false
			status.SyncProgress = 0
			return status
		}

		// It's synced and it works!
		status.IsSynced = true
		status.SyncProgress = 1
		return status

	}

	// It's not synced yet, print the progress
	status.IsWorking = true
	status.IsSynced = false

	status.SyncProgress = float64(progress.CurrentBlock) / float64(progress.HighestBlock)
	if status.SyncProgress > 1 {
		status.SyncProgress = 1
	}
	if math.IsNaN(status.SyncProgress) {
		status.SyncProgress = 0
	}

	return status

}

// Attempts to run a function progressively through each client until one succeeds or they all fail.
func (p *ExecutionClientManager) runFunction(function ecFunction) (interface{}, error) {

	// Check if we can use the primary
	if p.primaryReady {
		// Try to run the function on the primary
		result, err := function(p.primaryEc)
		if err != nil {
			if p.isDisconnected(err) {
				// If it's disconnected, log it and try the fallback
				p.logger.Printlnf("WARNING: Primary Execution client disconnected (%s), using fallback...", err.Error())
				p.primaryReady = false
				return p.runFunction(function)
			}

			// If it's a different error, just return it
			return nil, err
		}

		// If there's no error, return the result
		return result, nil
	}

	if p.fallbackReady {
		// Try to run the function on the fallback
		result, err := function(p.fallbackEc)
		if err != nil {
			if p.isDisconnected(err) {
				// If it's disconnected, log it and try the fallback
				p.logger.Printlnf("WARNING: Fallback Execution client disconnected (%s)", err.Error())
				p.fallbackReady = false
				return nil, fmt.Errorf("all Execution clients failed")
			}

			// If it's a different error, just return it
			return nil, err
		}

		// If there's no error, return the result
		return result, nil
	}

	return nil, fmt.Errorf("no Execution clients were ready")
}

// Returns true if the error was a connection failure and a backup client is available
func (p *ExecutionClientManager) isDisconnected(err error) bool {
	return strings.Contains(err.Error(), "dial tcp")
}
