package pdao

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"
	"github.com/wealdtech/go-ens/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getStatus(c *cli.Context) (*api.PDAOStatusResponse, error) {

	// Get services
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.PDAOStatusResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Sync
	var wg errgroup.Group
	var blockNumber uint64

	// Get the node onchain voting delegate
	wg.Go(func() error {
		var err error
		response.OnchainVotingDelegate, err = network.GetCurrentVotingDelegate(rp, nodeAccount.Address, nil)
		if err == nil {
			response.OnchainVotingDelegateFormatted = formatResolvedAddress(c, response.OnchainVotingDelegate)
		}
		return err
	})

	// Get latest block number
	wg.Go(func() error {
		_blockNumber, err := ec.BlockNumber(context.Background())
		if err != nil {
			return fmt.Errorf("Error getting block number: %w", err)
		}
		blockNumber = _blockNumber
		return nil
	})

	// Check if node is opted into pdao proposal checking duty
	wg.Go(func() error {
		var err error
		response.VerifyEnabled = cfg.Smartnode.VerifyProposals.Value.(bool)
		if err != nil {
			return fmt.Errorf("Error loading configuration: %w", err)
		}
		return nil
	})

	// Check if houston deployed and add to response
	wg.Go(func() error {
		var err error
		response.IsHoustonDeployed, err = state.IsHoustonDeployed(rp, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Cast to uint32
	response.BlockNumber = uint32(blockNumber)

	// Check voting power
	response.VotingPower, err = network.GetVotingPower(rp, nodeAccount.Address, response.BlockNumber, nil)
	if err != nil {
		return nil, err
	}

	// Update & return response
	return &response, nil
}

func formatResolvedAddress(c *cli.Context, address common.Address) string {
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return address.Hex()
	}

	name, err := ens.ReverseResolve(rp.Client, address)
	if err != nil {
		return address.Hex()
	}
	return fmt.Sprintf("%s (%s)", name, address.Hex())
}
