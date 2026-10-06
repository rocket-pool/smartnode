package pdao

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"
	"github.com/wealdtech/go-ens/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/network"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/proposals"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

func getStatus(c *cli.Command) (*api.PDAOStatusResponse, error) {

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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}
	reg, err := services.GetRocketSignerRegistry(c)
	if err != nil {
		return nil, err
	}
	if reg == nil {
		return nil, fmt.Errorf("Error getting the signer registry on network [%v].", cfg.Smartnode.Network.Value.(cfgtypes.Network))
	}

	// Response
	response := api.PDAOStatusResponse{}
	response.NodeRPLLocked = big.NewInt(0)

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	response.AccountAddress = nodeAccount.Address
	response.AccountAddressFormatted = formatResolvedAddress(c, response.AccountAddress)

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

	// Check whether RPL locking is allowed for the node
	wg.Go(func() error {
		var err error
		response.IsRPLLockingAllowed, err = node.GetRPLLockedAllowed(rp, nodeAccount.Address, nil)
		return err
	})

	// Get the node's locked RPL
	wg.Go(func() error {
		var err error
		response.NodeRPLLocked, err = node.GetNodeLockedRPL(rp, nodeAccount.Address, nil)
		return err
	})

	// Check if Node is registered
	wg.Go(func() error {
		var err error
		response.IsNodeRegistered, err = node.GetNodeExists(rp, nodeAccount.Address, nil)
		return err
	})

	// Get RocketDash proposals and votes, but treat errors as non-fatal
	if reg != nil {
		wg.Go(func() error {
			var err error
			r := &response.SnapshotResponse
			if cfg.Smartnode.GetRocketSignerRegistryAddress() != "" {
				response.SignallingAddress, err = reg.NodeToSigner(&bind.CallOpts{}, nodeAccount.Address)
				if err != nil {
					r.Error = err.Error()
					return nil
				}
				blankAddress := common.Address{}
				if response.SignallingAddress != blankAddress {
					response.SignallingAddressFormatted = formatResolvedAddress(c, response.SignallingAddress)
				}
			}
			*r, err = GetOffchainVotingStatus(cfg, rp, nodeAccount.Address, "active")
			if err != nil {
				r.Error = err.Error()
			}
			return nil
		})
	}

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Cast to uint32
	response.BlockNumber = uint32(blockNumber)

	// Get the proposal artifacts
	propMgr, err := proposals.NewProposalManager(nil, cfg, rp, bc)
	if err != nil {
		return nil, err
	}

	// Get the delegated voting power
	response.TotalDelegatedVp, _, _, err = propMgr.GetArtifactsForVoting(response.BlockNumber, nodeAccount.Address)
	if err != nil {
		return nil, err
	}

	// Get the local tree
	votingTree, err := propMgr.GetNetworkTree(response.BlockNumber, nil)
	if err != nil {
		return nil, err
	}
	response.SumVotingPower = votingTree.Nodes[0].Sum

	// Get voting power
	response.VotingPower, err = network.GetVotingPower(rp, nodeAccount.Address, response.BlockNumber, nil)
	if err != nil {
		return nil, err
	}

	// Update & return response
	return &response, nil
}

func formatResolvedAddress(c *cli.Command, address common.Address) string {
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

func statusHandler(ctx snroute.Context) {
	resp, err := getStatus(ctx.Command())
	response.WriteResponse(ctx.Writer, resp, err)
}
