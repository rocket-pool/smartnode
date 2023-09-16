package server

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	sharedtypes "github.com/rocket-pool/smartnode/shared/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

const (
	minipoolBatchSize int = 100
)

// Run a route registered with the common single-stage querying pattern
func runSingleStageRoute[DataType any](ctx ISingleStageCallContext[DataType], serviceProvider *services.ServiceProvider) (*api.ApiResponse[DataType], error) {
	// Get the services
	w := serviceProvider.GetWallet()
	rp := serviceProvider.GetRocketPool()

	// Initialize the context with any bootstrapping, requirements checks, or bindings it needs to set up
	err := ctx.Initialize()
	if err != nil {
		return nil, err
	}

	// Get the context-specific contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		ctx.GetState(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Get the transact opts if this node is ready for transaction
	var opts *bind.TransactOpts
	walletStatus := w.GetStatus()
	if walletStatus == sharedtypes.WalletStatus_Ready {
		var err error
		opts, err = w.GetTransactor()
		if err != nil {
			return nil, fmt.Errorf("error getting node account transactor: %w", err)
		}
	}

	// Create the response and data
	data := new(DataType)
	response := &api.ApiResponse[DataType]{
		WalletStatus: walletStatus,
		Data:         data,
	}

	// Prep the data with the context-specific behavior
	err = ctx.PrepareData(data, opts)
	if err != nil {
		return nil, err
	}

	// Return
	return response, nil
}

// Create a scaffolded generic minipool query, with caller-specific functionality where applicable
func runMinipoolRoute[DataType any](ctx IMinipoolCallContext[DataType], serviceProvider *services.ServiceProvider) (*api.ApiResponse[DataType], error) {
	// Common requirements
	err := serviceProvider.RequireNodeRegistered()
	if err != nil {
		return nil, err
	}

	// Get the services
	w := serviceProvider.GetWallet()
	rp := serviceProvider.GetRocketPool()
	nodeAddress, _ := serviceProvider.GetWallet().GetAddress()
	walletStatus := w.GetStatus()

	// Get the latest block for consistency
	latestBlock, err := rp.Client.BlockNumber(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting latest block number: %w", err)
	}
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(int64(latestBlock)),
	}

	// Create the bindings
	node, err := node.NewNode(rp, nodeAddress)
	if err != nil {
		return nil, fmt.Errorf("error creating node %s binding: %w", nodeAddress.Hex(), err)
	}

	// Supplemental function-specific bindings
	err = ctx.Initialize()
	if err != nil {
		return nil, err
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		node.GetMinipoolCount(mc)
		ctx.GetState(node, mc)
		return nil
	}, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Create the response and data
	data := new(DataType)
	response := &api.ApiResponse[DataType]{
		WalletStatus: walletStatus,
		Data:         data,
	}

	// Supplemental function-specific check to see if minipool processing should continue
	if !ctx.CheckState(node, data) {
		return response, nil
	}

	// Get the minipool addresses for this node
	addresses, err := node.GetMinipoolAddresses(node.Details.MinipoolCount.Formatted(), opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}

	// Create each minipool binding
	mps, err := minipool.CreateMinipoolsFromAddresses(rp, addresses, false, opts)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool bindings: %w", err)
	}

	// Get the relevant details
	err = rp.BatchQuery(len(addresses), minipoolBatchSize, func(mc *batch.MultiCaller, i int) error {
		ctx.GetMinipoolDetails(mc, mps[i], i) // Supplemental function-specific minipool details
		return nil
	}, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool details: %w", err)
	}

	// Get the transact opts if this node is ready for transaction
	var txOpts *bind.TransactOpts
	if walletStatus == sharedtypes.WalletStatus_Ready {
		var err error
		txOpts, err = w.GetTransactor()
		if err != nil {
			return nil, fmt.Errorf("error getting node account transactor: %w", err)
		}
	}

	// Supplemental function-specific response construction
	err = ctx.PrepareResponse(addresses, mps, data, txOpts)
	if err != nil {
		return nil, err
	}

	// Return
	return response, nil
}
