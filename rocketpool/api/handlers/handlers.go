package handlers

import (
	batch "github.com/rocket-pool/batch-query"
)

// Wrapper for callbacks used by call runners that follow a common single-stage pattern:
// Create bindings, query the chain, and then do whatever else they want.
// Structs implementing this will handle the caller-specific functionality.
type ISingleStageCallHandler[DataType any, ContextType any, ImplType any] interface {
	// Used to create supplemental contract bindings
	CreateBindings(ctx *ContextType) error

	// Used to get any supplemental state required during initialization - anything in here will be fed into an rp.Query() multicall
	GetState(ctx *ContextType, mc *batch.MultiCaller)

	// Prepare the response data using all of the provided artifacts
	PrepareData(ctx *ContextType, data *DataType) error

	// Enforce that implementing types must be a pointer (all functions must have pointer)
	*ImplType
}
