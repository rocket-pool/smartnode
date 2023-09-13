package minipool

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Wrapper for callbacks used by runMinipoolQuery; this implements the caller-specific functionality
type MinipoolCallHandler[responseType any] interface {
	// Used to create supplemental contract bindings (other than node.Node, which will already be created by the scaffolder);
	// this should create local variables that the caller keeps in scope throughout the life of runMinipoolQuery
	CreateBindings(rp *rocketpool.RocketPool) error

	// Used to get any supplemental state required during initialization - anything in here will be fed into an rp.Query() multicall
	GetState(node *node.Node, mc *batch.MultiCaller)

	// Check the initialized state after being queried to see if the response needs to be updated and the query can be ended prematurely
	// Return true if the function should continue, or false if it needs to end and just return the response as-is
	CheckState(node *node.Node, response *responseType) bool

	// Get whatever details of the given minipool are necessary; this will be passed into an rp.BatchQuery call, one run per minipool
	// belonging to the node
	GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.Minipool, index int)

	// Prepare the response object using all of the provided artifacts
	PrepareResponse(rp *rocketpool.RocketPool, bc beacon.Client, addresses []common.Address, mps []minipool.Minipool, response *responseType) error
}

// Create a scaffolded generic minipool query, with caller-specific functionality where applicable
func runMinipoolQuery[responseType any](c *cli.Context, h MinipoolCallHandler[responseType]) (*responseType, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, fmt.Errorf("error checking if node is registered: %w", err)
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, fmt.Errorf("error getting wallet: %w", err)
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, fmt.Errorf("error getting Rocket Pool binding: %w", err)
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, fmt.Errorf("error getting Beacon Node binding: %w", err)
	}
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}

	// Get the latest block for consistency
	latestBlock, err := rp.Client.BlockNumber(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting latest block number: %w", err)
	}
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(int64(latestBlock)),
	}

	// Response
	response := new(responseType)

	// Create the bindings
	node, err := node.NewNode(rp, nodeAccount.Address)
	if err != nil {
		return nil, fmt.Errorf("error creating node %s binding: %w", nodeAccount.Address.Hex(), err)
	}

	// Supplemental function-specific bindings
	err = h.CreateBindings(rp)
	if err != nil {
		return nil, err
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		node.GetMinipoolCount(mc)
		h.GetState(node, mc)
		return nil
	}, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Supplemental function-specific check to see if minipool processing should continue
	if !h.CheckState(node, response) {
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
		h.GetMinipoolDetails(mc, mps[i], i) // Supplemental function-specific minipool details
		return nil
	}, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool details: %w", err)
	}

	// Supplemental function-specific response construction
	err = h.PrepareResponse(rp, bc, addresses, mps, response)
	if err != nil {
		return nil, err
	}

	// Return
	return response, nil
}

// Get transaction info for an operation on all of the provided minipools, using the common minipool API (for version-agnostic functions)
func createBatchTxResponseForCommon(c *cli.Context, minipoolAddresses []common.Address, txCreator func(mpCommon *minipool.MinipoolCommon, opts *bind.TransactOpts) (*core.TransactionInfo, error), txName string) (*api.BatchTxResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.BatchTxResponse{}

	// Create minipools
	mps, err := minipool.CreateMinipoolsFromAddresses(rp, minipoolAddresses, false, nil)
	if err != nil {
		return nil, err
	}

	// Get the TXs
	txInfos := make([]*core.TransactionInfo, len(minipoolAddresses))
	for i, mp := range mps {
		mpCommon := mp.GetMinipoolCommon()
		txInfo, err := txCreator(mpCommon, opts)
		if err != nil {
			return nil, fmt.Errorf("error simulating %s transaction for minipool %s: %w", txName, mpCommon.Details.Address.Hex(), err)
		}
		txInfos[i] = txInfo
	}

	response.TxInfos = txInfos
	return &response, nil
}

// Get transaction info for an operation on all of the provided minipools, using the v3 minipool API (for Atlas-specific functions)
func createBatchTxResponseForV3(c *cli.Context, minipoolAddresses []common.Address, txCreator func(mpv3 *minipool.MinipoolV3, opts *bind.TransactOpts) (*core.TransactionInfo, error), txName string) (*api.BatchTxResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.BatchTxResponse{}

	// Create minipools
	mps, err := minipool.CreateMinipoolsFromAddresses(rp, minipoolAddresses, false, nil)
	if err != nil {
		return nil, err
	}

	// Get the TXs
	txInfos := make([]*core.TransactionInfo, len(minipoolAddresses))
	for i, mp := range mps {
		mpCommon := mp.GetMinipoolCommon()
		minipoolAddress := mpCommon.Details.Address
		mpv3, success := minipool.GetMinipoolAsV3(mp)
		if !success {
			return nil, fmt.Errorf("minipool %s is too old (current version: %d); please upgrade the delegate for it first", minipoolAddress.Hex(), mpCommon.Details.Version)
		}
		txInfo, err := txCreator(mpv3, opts)
		if err != nil {
			return nil, fmt.Errorf("error simulating %s transaction for minipool %s: %w", txName, minipoolAddress.Hex(), err)
		}
		txInfos[i] = txInfo
	}

	response.TxInfos = txInfos
	return &response, nil
}
