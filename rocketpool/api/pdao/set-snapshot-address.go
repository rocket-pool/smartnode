package pdao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	apiutils "github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"

	"github.com/urfave/cli"
)

func canSetSnapshotAddress(c *cli.Context, snapshotAddress common.Address, signature string) (*api.PDAOCanSetSnapshotAddressResponse, error) {

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
	response := api.PDAOCanSetSnapshotAddressResponse{}

	votingInitialized, err := network.GetVotingInitialized(rp, nodeAccount.Address, nil)
	if !votingInitialized {
		return nil, fmt.Errorf("Voting must be initialized to set a snapshot address. Use 'rocketpool pdao initialize-voting' to initialize voting first.")
	}
	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Parse signature into vrs components, v to uint8 and v,s to *[32]byte
	_v, _r, _s, err := apiutils.ParseEIP712(signature)
	if err != nil {
		fmt.Println("Error parsing signature", err)
	}

	// Gas info
	gasInfo, err := network.EstimateSetSnapshotAddress(rp, snapshotAddress, _v, _r, _s, opts)
	if err != nil {
		return nil, fmt.Errorf("Could not estimate the gas required set snapshot address: %w", err)
	}
	response.GasInfo = gasInfo

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
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.PDAOSetSnapshotAddressResponse{}

	votingInitialized, err := network.GetVotingInitialized(rp, nodeAccount.Address, nil)
	if !votingInitialized {
		return nil, fmt.Errorf("Voting must be initialized to set a snapshot address. Use 'rocketpool pdao initialize-voting' to initialize voting first.")
	}

	// Parse signature into vrs components, v to uint8 and v,s to *[32]byte
	_v, _r, _s, err := apiutils.ParseEIP712(signature)
	if err != nil {
		fmt.Println("Error parsing signature", err)
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

	// Todo:
	// Network call set-snapshot-address on RocketSignerRegistry
	hash, err := network.SetSnapshotAddress(rp, snapshotAddress, _v, _r, _s, opts)
	if err != nil {
		return nil, fmt.Errorf("Error setting snapshot address: %w", err)
	}
	response.TxHash = hash

	return &response, nil
}
