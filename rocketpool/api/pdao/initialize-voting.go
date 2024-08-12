package pdao

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canNodeInitializeVoting(c *cli.Context) (*api.PDAOCanInitializeVotingResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
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
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.PDAOCanInitializeVotingResponse{}

	isInitialized, err := network.GetVotingInitialized(rp, nodeAccount.Address, nil)
	if isInitialized {
		return nil, fmt.Errorf("voting already initialized")
	}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	gasInfo, err := network.EstimateInitializeVotingGas(rp, opts)
	if err != nil {
		return nil, fmt.Errorf("Could not estimate the gas required to initialize voting: %w", err)
	}
	response.GasInfo = gasInfo

	return &response, nil
}

func nodeInitializeVoting(c *cli.Context) (*api.PDAOInitializeVotingResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
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
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.PDAOInitializeVotingResponse{}

	isInitialized, err := network.GetVotingInitialized(rp, nodeAccount.Address, nil)
	if isInitialized {
		return nil, fmt.Errorf("voting already initialized")
	}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	hash, err := network.InitializeVoting(rp, opts)
	if err != nil {
		return nil, fmt.Errorf("Error initializing voting: %w", err)
	}
	response.TxHash = hash

	return &response, nil

}
