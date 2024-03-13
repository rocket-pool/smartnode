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
	scrollFeeEstimatorAbiString string = `[{"inputs": [{"internalType": "uint256","name": "_l2GasLimit","type": "uint256"}],"name": "estimateCrossDomainMessageFee","outputs": [ {"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type": "function"}]`
)

// ABI cache
var scrollFeeEstimatorAbi abi.ABI
var scrollFeeEstimatorOnce sync.Once

// ===============
// === Structs ===
// ===============

// Binding for the Scroll Fee Estimator
type ScrollFeeEstimator struct {
	contract *eth.Contract
}

// ====================
// === Constructors ===
// ====================

// Creates a new Scroll Fee Estimator contract binding
func NewScrollFeeEstimator(address common.Address, client eth.IExecutionClient) (*ScrollFeeEstimator, error) {
	// Parse the ABI
	var err error
	scrollFeeEstimatorOnce.Do(func() {
		var parsedAbi abi.ABI
		parsedAbi, err = abi.JSON(strings.NewReader(scrollFeeEstimatorAbiString))
		if err == nil {
			scrollFeeEstimatorAbi = parsedAbi
		}
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing scroll fee estimator ABI: %w", err)
	}

	// Create the contract
	contract := &eth.Contract{
		ContractImpl: bind.NewBoundContract(address, scrollFeeEstimatorAbi, client, client, client),
		Address:      address,
		ABI:          &scrollFeeEstimatorAbi,
	}

	return &ScrollFeeEstimator{
		contract: contract,
	}, nil
}

// =============
// === Calls ===
// =============

// Estimate the fee for sending a message to the Scroll L2
func (c *ScrollFeeEstimator) EstimateCrossDomainMessageFee(mc *batch.MultiCaller, out **big.Int, l2GasLimit *big.Int) {
	eth.AddCallToMulticaller(mc, c.contract, out, "estimateCrossDomainMessageFee", l2GasLimit)
}
