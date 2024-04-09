package contracts

import (
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
)

const (
	arbitrumMessengerAbiString string = `[{"inputs":[],"name":"rateStale","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"_maxSubmissionCost","type":"uint256"},{"internalType":"uint256","name":"_gasLimit","type":"uint256"},{"internalType":"uint256","name":"_gasPriceBid","type":"uint256"}],"name":"submitRate","outputs":[],"stateMutability":"payable","type":"function"}]`
)

// ABI cache
var arbitrumMessengerAbi abi.ABI
var arbitrumOnce sync.Once

// ===============
// === Structs ===
// ===============

// Binding for the Arbitrum Messenger
type ArbitrumMessenger struct {
	contract *eth.Contract
	txMgr    *eth.TransactionManager
}

// ====================
// === Constructors ===
// ====================

// Creates a new Arbitrum Messenger contract binding
func NewArbitrumMessenger(address common.Address, client eth.IExecutionClient, txMgr *eth.TransactionManager) (*ArbitrumMessenger, error) {
	// Parse the ABI
	var err error
	arbitrumOnce.Do(func() {
		var parsedAbi abi.ABI
		parsedAbi, err = abi.JSON(strings.NewReader(arbitrumMessengerAbiString))
		if err == nil {
			arbitrumMessengerAbi = parsedAbi
		}
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing Arbitrum messenger ABI: %w", err)
	}

	// Create the contract
	contract := &eth.Contract{
		ContractImpl: bind.NewBoundContract(address, arbitrumMessengerAbi, client, client, client),
		Address:      address,
		ABI:          &arbitrumMessengerAbi,
	}

	return &ArbitrumMessenger{
		contract: contract,
		txMgr:    txMgr,
	}, nil
}

// =============
// === Calls ===
// =============

// Check if the RPL rate is stale and needs to be updated
func (c *ArbitrumMessenger) IsRateStale(mc *batch.MultiCaller, out *bool) {
	eth.AddCallToMulticaller(mc, c.contract, out, "rateStale")
}

// ====================
// === Transactions ===
// ====================

// Send the latest RPL rate to the L2
func (c *ArbitrumMessenger) SubmitRate(maxSubmissionCost *big.Int, gasLimit *big.Int, gasPriceBid *big.Int, opts *bind.TransactOpts) (*eth.TransactionInfo, error) {
	return c.txMgr.CreateTransactionInfo(c.contract, "submitRate", opts, maxSubmissionCost, gasLimit, gasPriceBid)
}
