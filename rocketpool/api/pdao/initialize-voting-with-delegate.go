package pdao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/bindings/network"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canNodeInitializeVotingWithDelegate(c *cli.Context, delegateAddress common.Address) (*api.PDAOCanInitializeVotingResponse, error) {

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
	if err != nil {
		return nil, err
	}
	if isInitialized {
		return nil, fmt.Errorf("voting already initialized")
	}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	gasInfo, err := network.EstimateInitializeVotingWithDelegateGas(rp, delegateAddress, opts)
	if err != nil {
		return nil, fmt.Errorf("Could not estimate the gas required to claim RPL: %w", err)
	}
	response.GasInfo = gasInfo

	return &response, nil
}

func nodeInitializeVotingWithDelegate(c *cli.Context, delegateAddress common.Address) (*api.PDAOInitializeVotingResponse, error) {

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
	if err != nil {
		return nil, err
	}
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

	hash, err := network.InitializeVotingWithDelegate(rp, delegateAddress, opts)
	if err != nil {
		return nil, fmt.Errorf("Error initializing voting: %w", err)
	}
	response.TxHash = hash

	return &response, nil

}
