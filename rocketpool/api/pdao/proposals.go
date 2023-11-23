package pdao

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

// Settings
const (
	ProposalDetailsBatchSize = 10
)

func getProposals(c *cli.Context) (*api.PDAOProposalsResponse, error) {

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

	// Response
	response := api.PDAOProposalsResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get proposals
	proposals, err := protocol.GetProposals(rp, nil)
	if err != nil {
		return nil, err
	}
	augmentedProps, err := getProposalsWithNodeVoteDirection(rp, nodeAccount.Address, proposals)
	if err != nil {
		return nil, err
	}

	response.Proposals = augmentedProps

	// Return response
	return &response, nil

}

func getProposalsWithNodeVoteDirection(rp *rocketpool.RocketPool, nodeAddress common.Address, props []protocol.ProtocolDaoProposalDetails) ([]api.PDAOProposalWithNodeVoteDirection, error) {
	// Load node votes in batches
	proposalCount := uint64(len(props))
	details := make([]api.PDAOProposalWithNodeVoteDirection, proposalCount)
	for bsi := uint64(0); bsi < proposalCount; bsi += ProposalDetailsBatchSize {
		// Get batch start & end index
		psi := bsi
		pei := bsi + ProposalDetailsBatchSize
		if pei > proposalCount {
			pei = proposalCount
		}

		// Load details
		var wg errgroup.Group
		for pi := psi; pi < pei; pi++ {
			pi := pi
			wg.Go(func() error {
				propID := props[pi].ID
				voteDir, err := protocol.GetAddressVoteDirection(rp, propID, nodeAddress, nil)
				if err == nil {
					details[pi].NodeVoteDirection = voteDir
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return nil, err
		}
	}

	return details, nil
}

func getProposal(c *cli.Context, id uint64) (*api.PDAOProposalResponse, error) {

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

	// Response
	response := api.PDAOProposalResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get proposal
	proposal, err := protocol.GetProposalDetails(rp, id, nil)
	if err != nil {
		return nil, err
	}

	// Get the node vote direction
	voteDir, err := protocol.GetAddressVoteDirection(rp, id, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Make the augmented proposal
	augmentedProp := api.PDAOProposalWithNodeVoteDirection{
		ProtocolDaoProposalDetails: proposal,
		NodeVoteDirection:          voteDir,
	}
	response.Proposal = augmentedProp

	// Return response
	return &response, nil

}
