package network

import (
	"github.com/rocket-pool/smartnode/rocketpool/api/pdao"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func getActiveDAOProposals(c *cli.Context) (*api.NetworkDAOProposalsResponse, error) {

	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	s, err := services.GetSnapshotDelegation(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	response := api.NetworkDAOProposalsResponse{}
	response.AccountAddress = nodeAccount.Address

	// Return nothing if Snapshot isn't available on this network
	if s == nil {
		return &response, nil
	}

	// Get snapshot proposals
	snapshotResponse, err := pdao.GetSnapshotProposals(cfg.Smartnode.GetSnapshotApiDomain(), cfg.Smartnode.GetSnapshotID(), "active")
	if err != nil {
		return nil, err
	}

	// Get delegate address
	idHash := cfg.Smartnode.GetVotingSnapshotID()
	response.VotingDelegate, err = s.Delegation(nil, nodeAccount.Address, idHash)
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
	return &response, nil
}
