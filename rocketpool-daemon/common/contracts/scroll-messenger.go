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
	scrollMessengerAbiString string = `[{"inputs": [],"name": "rateStale","outputs": [{"internalType": "bool","name": "","type": "bool"}],"stateMutability": "view","type": "function"},{"inputs": [{"internalType": "uint256","name": "_l2GasLimit","type": "uint256"}],"name": "submitRate","outputs": [],"stateMutability": "payable","type": "function"}]`
)

// ABI cache
var scrollMessengerAbi abi.ABI
var scrollOnce sync.Once

// ===============
// === Structs ===
// ===============

// Binding for the Scroll Messenger
type ScrollMessenger struct {
	contract *eth.Contract
	txMgr    *eth.TransactionManager
}

// ====================
// === Constructors ===
// ====================

// Creates a new Scroll Messenger contract binding
func NewScrollMessenger(address common.Address, client eth.IExecutionClient, txMgr *eth.TransactionManager) (*ScrollMessenger, error) {
	// Parse the ABI
	var err error
	scrollOnce.Do(func() {
		var parsedAbi abi.ABI
		parsedAbi, err = abi.JSON(strings.NewReader(scrollMessengerAbiString))
		if err == nil {
			scrollMessengerAbi = parsedAbi
		}
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing scroll messenger ABI: %w", err)
	}

	// Create the contract
	contract := &eth.Contract{
		ContractImpl: bind.NewBoundContract(address, scrollMessengerAbi, client, client, client),
		Address:      address,
		ABI:          &scrollMessengerAbi,
	}

	return &ScrollMessenger{
		contract: contract,
	}, nil
}

// =============
// === Calls ===
// =============

// Check if the RPL rate is stale and needs to be updated
func (c *ScrollMessenger) IsRateStale(mc *batch.MultiCaller, out *bool) {
	eth.AddCallToMulticaller(mc, c.contract, out, "rateStale")
}

// ====================
// === Transactions ===
// ====================

// Send the latest RPL rate to the L2
func (c *ScrollMessenger) SubmitRate(l2GasLimit *big.Int, opts *bind.TransactOpts) (*eth.TransactionInfo, error) {
	return c.txMgr.CreateTransactionInfo(c.contract, "submitRate", opts, l2GasLimit)
}
