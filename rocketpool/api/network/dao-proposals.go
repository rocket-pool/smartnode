package network

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/pdao"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/proposals"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
	"github.com/wealdtech/go-ens/v3"
	"golang.org/x/sync/errgroup"
)

func getActiveDAOProposals(c *cli.Context) (*api.NetworkDAOProposalsResponse, error) {

	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	response := api.NetworkDAOProposalsResponse{}

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

	// Check if Voting is initialized and add to response
	wg.Go(func() error {
		var err error
		response.IsVotingInitialized, err = network.GetVotingInitialized(rp, nodeAccount.Address, nil)
		return err
	})

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

	// Check if Node is registered
	wg.Go(func() error {
		var err error
		response.IsNodeRegistered, err = node.GetNodeExists(rp, nodeAccount.Address, nil)
		return err
	})

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

	// Get the delegated voting power if voting is initialized
	if response.IsVotingInitialized {
		totalDelegatedVP, _, _, err := propMgr.GetArtifactsForVoting(response.BlockNumber, nodeAccount.Address)
		if err != nil {
			return nil, err
		}
		response.TotalDelegatedVp = totalDelegatedVP
	} else {
		response.TotalDelegatedVp = nil
	}

	// Get the local tree
	votingTree, err := propMgr.GetNetworkTree(response.BlockNumber, nil)
	if err != nil {
		return nil, err
	}
	response.SumVotingPower = votingTree.Nodes[0].Sum

	// Get snapshot proposals
	snapshotResponse, err := pdao.GetSnapshotProposals(cfg.Smartnode.GetSnapshotApiDomain(), cfg.Smartnode.GetSnapshotID(), "active")
	if err != nil {
		return nil, err
	}

	// Get voted proposals
	votedProposals, err := pdao.GetSnapshotVotedProposals(cfg.Smartnode.GetSnapshotApiDomain(), cfg.Smartnode.GetSnapshotID(), nodeAccount.Address, response.VotingDelegate)
	if err != nil {
		return nil, err
	}
	response.ProposalVotes = votedProposals.Data.Votes
	response.ActiveSnapshotProposals = snapshotResponse.Data.Proposals

	// Get voting power
	response.VotingPower, err = network.GetVotingPower(rp, nodeAccount.Address, response.BlockNumber, nil)
	if err != nil {
		return nil, err
	}

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
