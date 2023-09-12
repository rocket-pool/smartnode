package faucet

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/contracts"
	"github.com/urfave/cli"
)

// Wrapper for callbacks used by runFaucetCall; this implements the caller-specific functionality
type FaucetCallHandler[responseType any] interface {
	// Used to create supplemental contract bindings
	CreateBindings(rp *rocketpool.RocketPool) error

	// Used to get any supplemental state required during initialization - anything in here will be fed into an rp.Query() multicall
	GetState(f *contracts.RplFaucet, nodeAddress common.Address, mc *batch.MultiCaller)

	// Prepare the response object using all of the provided artifacts
	PrepareResponse(rp *rocketpool.RocketPool, f *contracts.RplFaucet, nodeAccount accounts.Account, opts *bind.TransactOpts, response *responseType) error
}

// Create a scaffolded generic call handler, with caller-specific functionality where applicable
func runFaucetCall[responseType any](c *cli.Context, q FaucetCallHandler[responseType]) (*responseType, error) {
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
	f, err := services.GetRplFaucet(c)
	if err != nil {
		return nil, err
	}
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, fmt.Errorf("error getting node account transactor: %w", err)
	}

	// Response
	response := new(responseType)

	// Supplemental function-specific bindings
	err = q.CreateBindings(rp)
	if err != nil {
		return nil, err
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		q.GetState(f, nodeAccount.Address, mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Supplemental function-specific response construction
	err = q.PrepareResponse(rp, f, nodeAccount, opts, response)
	if err != nil {
		return nil, err
	}

	// Return
	return response, nil
}
