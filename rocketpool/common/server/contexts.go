package server

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
)

// Wrapper for callbacks used by call runners that follow a common single-stage pattern:
// Create bindings, query the chain, and then do whatever else they want.
// Structs implementing this will handle the caller-specific functionality.
type ISingleStageCallContext[DataType any] interface {
	// Initialize the context with any bootstrapping, requirements checks, or bindings it needs to set up
	Initialize() error

	// Used to get any supplemental state required during initialization - anything in here will be fed into an rp.Query() multicall
	GetState(mc *batch.MultiCaller)

	// Prepare the response data in whatever way the context needs to do
	PrepareData(data *DataType, opts *bind.TransactOpts) error
}

// Interface for single-stage call context factories - these will be invoked during route handling to create the
// unique context for the route
type ISingleStageCallContextFactory[ContextType ISingleStageCallContext[DataType], DataType any] interface {
	// Create the context for the route
	Create(vars map[string]string) (ContextType, error)
}

// Wrapper for callbacks used by functions that will query all of the node's minipools - they follow this pattern:
// Create bindings, query the chain, return prematurely if some state isn't correct, query all of the minipools, and process them to
// populate a response.
// Structs implementing this will handle the caller-specific functionality.
type IMinipoolCallContext[DataType any] interface {
	// Initialize the context with any bootstrapping, requirements checks, or bindings it needs to set up
	Initialize() error

	// Used to get any supplemental state required during initialization - anything in here will be fed into an rp.Query() multicall
	GetState(node *node.Node, mc *batch.MultiCaller)

	// Check the initialized state after being queried to see if the response needs to be updated and the query can be ended prematurely
	// Return true if the function should continue, or false if it needs to end and just return the response as-is
	CheckState(node *node.Node, data *DataType) bool

	// Get whatever details of the given minipool are necessary; this will be passed into an rp.BatchQuery call, one run per minipool
	// belonging to the node
	GetMinipoolDetails(mc *batch.MultiCaller, mp minipool.Minipool, index int)

	// Prepare the response data using all of the provided artifacts
	PrepareResponse(addresses []common.Address, mps []minipool.Minipool, data *DataType, opts *bind.TransactOpts) error
}

// Interface for minipool call context factories - these will be invoked during route handling to create the
// unique context for the route
type IMinipoolCallContextFactory[ContextType IMinipoolCallContext[DataType], DataType any] interface {
	// Create the context for the route
	Create(vars map[string]string) (ContextType, error)
}
