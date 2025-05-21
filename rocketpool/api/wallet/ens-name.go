package wallet

import (
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
	ens "github.com/wealdtech/go-ens/v3"
)

const (
	GasLimitMultiplier float64 = 1.5
	MaxGasLimit        uint64  = 30000000
)

// Set a name to the node wallet's ENS reverse record.
func setEnsName(c *cli.Context, name string, onlyEstimateGas bool) (*api.SetEnsNameResponse, error) {
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	account, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Name validation
	if name == "" {
		return nil, fmt.Errorf("name cannot be blank")
	}

	// The ENS name must resolve to the wallet address
	resolvedAddress, err := ens.Resolve(rp.Client, name)
	if err != nil {
		return nil, fmt.Errorf("error resolving '%s' to an address: %w", name, err)
	}

	if resolvedAddress != account.Address {
		return nil, fmt.Errorf("error: %s currently resolves to the address %s instead of the node wallet address %s", name, resolvedAddress.Hex(), account.Address.Hex())
	}

	// Check if the name is already in use
	resolvedName, err := ens.ReverseResolve(rp.Client, account.Address)
	if err != nil && err.Error() != "not a resolver" {
		// Handle errors unrelated to the address not being an ENS resolver
		return nil, fmt.Errorf("error reverse resolving %s to an ENS name: %w", account.Address.Hex(), err)
	} else if resolvedName == name {
		return nil, fmt.Errorf("error: the ENS record already points to the name '%s'", name)
	}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// If onlyEstimateGas is set, then don't send the tx, only simulates and returns the gas estimate
	opts.NoSend = onlyEstimateGas

	registrar, err := ens.NewReverseRegistrar(rp.Client)
	if err != nil {
		return nil, fmt.Errorf("error creating reverse registrar binding: %w", err)
	}
	tx, err := registrar.SetName(opts, name)
	if err != nil {
		return nil, fmt.Errorf("error setting ENS name: %w", err)
	}
	response := api.SetEnsNameResponse{
		Address: account.Address,
		EnsName: name,
		TxHash:  tx.Hash(),
		GasInfo: rocketpool.GasInfo{
			EstGasLimit:  tx.Gas(),
			SafeGasLimit: uint64(float64(tx.Gas()) * GasLimitMultiplier),
		},
	}

	if response.GasInfo.EstGasLimit > MaxGasLimit {
		return nil, fmt.Errorf("estimated gas of %d is greater than the max gas limit of %d", response.GasInfo.EstGasLimit, MaxGasLimit)
	}
	if response.GasInfo.SafeGasLimit > MaxGasLimit {
		response.GasInfo.SafeGasLimit = MaxGasLimit
	}

	return &response, nil
}
