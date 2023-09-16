package server

import (
	batch "github.com/rocket-pool/batch-query"
)

// Wrapper for callbacks used by call runners that follow a common single-stage pattern:
// Create bindings, query the chain, and then do whatever else they want.
// Structs implementing this will handle the caller-specific functionality.
type ISingleStageCallContext[DataType any, CommonContextType any] interface {
	// Used to create supplemental contract bindings
	CreateBindings(ctx *CommonContextType) error

	// Used to get any supplemental state required during initialization - anything in here will be fed into an rp.Query() multicall
	GetState(mc *batch.MultiCaller)

	// Prepare the response data using all of the provided artifacts
	PrepareData(data *DataType) error
}
