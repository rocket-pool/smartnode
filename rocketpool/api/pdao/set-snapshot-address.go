package pdao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
)

func canSetSnapshotAddress(c *cli.Context, snapshotAddress common.Address, signature string) (*api.PDAOCanSetSnapshotAddressResponse, error) {

	// // Get services
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

	// Response
	response := api.PDAOCanSetSnapshotAddressResponse{}

	// TODO:
	// check if the node can set a snapshot delegate
	// conditions required:
	// 		- on chain voting must initialized
	// 		- gas estimation must pass

	// Gas info

	// Update response

	return &response, nil
}

func setSnapshotAddress(c *cli.Context, snapshotAddress common.Address, signature string) (*api.PDAOSetSnapshotAddressResponse, error) {

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

	//response
	response := api.PDAOSetSnapshotAddressResponse{}

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

	// Todo:
	// Network call set-snapshot-address on RocketSignerRegistry
	// network.SetSnapshotAddress in the rocketpool-go lib

	// Update response with txhash

	return &response, nil
}
