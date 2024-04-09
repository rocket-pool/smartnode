package minipool

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
)

// Wrapper for callbacks used by functions that will query all of the node's minipools - they follow this pattern:
// Create bindings, query the chain, return prematurely if some state isn't correct, query all of the minipools, and process them to
// populate a response.
// Structs implementing this will handle the caller-specific functionality.
type IMinipoolCallContext[DataType any] interface {
	// Initialize the context with any bootstrapping, requirements checks, or bindings it needs to set up
	Initialize() (types.ResponseStatus, error)

	// Used to get any supplemental state required during initialization - anything in here will be fed into an rp.Query() multicall
	GetState(node *node.Node, mc *batch.MultiCaller)

	// Check the initialized state after being queried to see if the response needs to be updated and the query can be ended prematurely
	// Return true if the function should continue, or false if it needs to end and just return the response as-is
	CheckState(node *node.Node, data *DataType) bool

	// Get whatever details of the given minipool are necessary; this will be passed into an rp.BatchQuery call, one run per minipool
	// belonging to the node
	GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.IMinipool, index int)

	// Prepare the response data using all of the provided artifacts
	PrepareData(addresses []common.Address, mps []minipool.IMinipool, data *DataType) (types.ResponseStatus, error)
}

// Interface for minipool call context factories - these will be invoked during route handling to create the
// unique context for the route
type IMinipoolCallContextFactory[ContextType IMinipoolCallContext[DataType], DataType any] interface {
	// Create the context for the route
	Create(args url.Values) (ContextType, error)
}

// Registers a new route with the router, which will invoke the provided factory to create and execute the context
// for the route when it's called; use this for complex calls that will iterate over and query each minipool in the node
func RegisterMinipoolRoute[ContextType IMinipoolCallContext[DataType], DataType any](
	router *mux.Router,
	functionName string,
	factory IMinipoolCallContextFactory[ContextType, DataType],
	ctx context.Context,
	logger *log.Logger,
	serviceProvider *services.ServiceProvider,
) {
	router.HandleFunc(fmt.Sprintf("/%s", functionName), func(w http.ResponseWriter, r *http.Request) {
		// Log
		args := r.URL.Query()
		logger.Info("New request", slog.String(log.MethodKey, r.Method), slog.String(log.PathKey, r.URL.Path))
		logger.Debug("Request params:", slog.String(log.QueryKey, r.URL.RawQuery))

		// Check the method
		if r.Method != http.MethodGet {
			server.HandleInvalidMethod(logger.Logger, w)
			return
		}

		// Create the handler and deal with any input validation errors
		mpContext, err := factory.Create(args)
		if err != nil {
			server.HandleInputError(logger.Logger, w, err)
			return
		}

		// Run the context's processing routine
		status, response, err := runMinipoolRoute[DataType](ctx, mpContext, serviceProvider)
		server.HandleResponse(logger.Logger, w, status, response, err)
	})
}

// Create a scaffolded generic minipool query, with caller-specific functionality where applicable
func runMinipoolRoute[DataType any](ctx context.Context, mpContext IMinipoolCallContext[DataType], serviceProvider *services.ServiceProvider) (types.ResponseStatus, *types.ApiResponse[DataType], error) {
	// Get the services
	w := serviceProvider.GetWallet()
	q := serviceProvider.GetQueryManager()
	rp := serviceProvider.GetRocketPool()
	nodeAddress, _ := w.GetAddress()

	// Common requirements
	status, err := serviceProvider.RequireNodeRegistered(ctx)
	if err != nil {
		return status, nil, err
	}

	// Get the latest block for consistency
	latestBlock, err := rp.Client.BlockNumber(ctx)
	if err != nil {
		return types.ResponseStatus_Error, nil, fmt.Errorf("error getting latest block number: %w", err)
	}
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(int64(latestBlock)),
	}

	// Create the bindings
	node, err := node.NewNode(rp, nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, nil, fmt.Errorf("error creating node %s binding: %w", nodeAddress.Hex(), err)
	}

	// Supplemental function-specific bindings
	status, err = mpContext.Initialize()
	if err != nil {
		return status, nil, err
	}

	// Get contract state
	err = q.Query(func(mc *batch.MultiCaller) error {
		node.MinipoolCount.AddToQuery(mc)
		mpContext.GetState(node, mc)
		return nil
	}, opts)
	if err != nil {
		return types.ResponseStatus_Error, nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Create the response and data
	data := new(DataType)
	response := &types.ApiResponse[DataType]{
		Data: data,
	}

	// Supplemental function-specific check to see if minipool processing should continue
	if !mpContext.CheckState(node, data) {
		return types.ResponseStatus_Success, response, nil
	}

	// Get the minipool addresses for this node
	addresses, err := node.GetMinipoolAddresses(node.MinipoolCount.Formatted(), opts)
	if err != nil {
		return types.ResponseStatus_Error, nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}

	// Create each minipool binding
	mpMgr, err := minipool.NewMinipoolManager(rp)
	if err != nil {
		return types.ResponseStatus_Error, nil, fmt.Errorf("error creating minipool manager binding: %w", err)
	}
	mps, err := mpMgr.CreateMinipoolsFromAddresses(addresses, false, opts)
	if err != nil {
		return types.ResponseStatus_Error, nil, fmt.Errorf("error creating minipool bindings: %w", err)
	}

	// Get the relevant details
	err = rp.BatchQuery(len(addresses), minipoolBatchSize, func(mc *batch.MultiCaller, i int) error {
		mpContext.GetMinipoolDetails(mc, mps[i], i) // Supplemental function-specific minipool details
		return nil
	}, opts)
	if err != nil {
		return types.ResponseStatus_Error, nil, fmt.Errorf("error getting minipool details: %w", err)
	}

	// Supplemental function-specific response construction
	status, err = mpContext.PrepareData(addresses, mps, data)
	return status, response, err
}
