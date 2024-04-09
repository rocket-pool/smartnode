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
	rplTwapPoolAbiString string = `[{"inputs":[{"internalType":"uint32[]","name":"secondsAgos","type":"uint32[]"}],"name":"observe","outputs":[{"internalType":"int56[]","name":"tickCumulatives","type":"int56[]"},{"internalType":"uint160[]","name":"secondsPerLiquidityCumulativeX128s","type":"uint160[]"}],"stateMutability":"view","type":"function"}]`
)

// ABI cache
var rplTwapPoolAbi abi.ABI
var rplTwapPoolOnce sync.Once

// ===============
// === Structs ===
// ===============

// Binding for the zkSync Era Messenger
type RplTwapPool struct {
	contract *eth.Contract
}

// Response from the Observe function
type PoolObserveResponse struct {
	TickCumulatives                    []*big.Int `abi:"tickCumulatives"`
	SecondsPerLiquidityCumulativeX128s []*big.Int `abi:"secondsPerLiquidityCumulativeX128s"`
}

// ====================
// === Constructors ===
// ====================

// Creates a new RPL TWAP Pool contract binding
func NewRplTwapPool(address common.Address, client eth.IExecutionClient) (*RplTwapPool, error) {
	// Parse the ABI
	var err error
	rplTwapPoolOnce.Do(func() {
		var parsedAbi abi.ABI
		parsedAbi, err = abi.JSON(strings.NewReader(rplTwapPoolAbiString))
		if err == nil {
			rplTwapPoolAbi = parsedAbi
		}
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing RPL TWAP pool ABI: %w", err)
	}

	// Create the contract
	contract := &eth.Contract{
		ContractImpl: bind.NewBoundContract(address, rplTwapPoolAbi, client, client, client),
		Address:      address,
		ABI:          &rplTwapPoolAbi,
	}

	return &RplTwapPool{
		contract: contract,
	}, nil
}

// =============
// === Calls ===
// =============

// Get the TWAP RPL price from the pool
func (c *RplTwapPool) Observe(mc *batch.MultiCaller, out *PoolObserveResponse, secondsAgos []uint32) {
	eth.AddCallToMulticaller(mc, c.contract, out, "observe", secondsAgos)
}
