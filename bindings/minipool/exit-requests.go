package minipool

import (
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

// A pending request for a minipool validator to exit the beacon chain
type MinipoolExitRequest struct {
	ValidatorIndex   uint64 // Beacon chain validator index
	RequestTimestamp uint64 // Unix seconds when the exit was requested
}

// Get the list of pending minipool exit requests
// TODO: stub — the contract view is not available yet; replace with the real
// contract call once it is deployed. Returns a fixed example response for now.
func GetMinipoolExitRequests(rp *rocketpool.RocketPool, opts *bind.CallOpts) ([]MinipoolExitRequest, error) {
	return []MinipoolExitRequest{
		{
			ValidatorIndex:   1000,
			RequestTimestamp: uint64(time.Now().Add(-48 * time.Hour).Unix()),
		},
		{
			ValidatorIndex:   1001,
			RequestTimestamp: uint64(time.Now().Add(-1 * time.Hour).Unix()),
		},
	}, nil
}
