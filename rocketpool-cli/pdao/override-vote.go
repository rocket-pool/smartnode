package pdao

import (
	"fmt"
	"time"

	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

func overrideVote(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get Protocol DAO proposals
	proposals, err := rp.Api.PDao.Proposals()
	if err != nil {
		return err
	}

	// Get votable proposals
	votableProposals := []api.ProtocolDaoProposalDetails{}
	for _, proposal := range proposals.Data.Proposals {
		if proposal.State == types.ProtocolDaoProposalState_ActivePhase2 && proposal.NodeVoteDirection == types.VoteDirection_NoVote {
			votableProposals = append(votableProposals, proposal)
		}
	}

	// Check for votable proposals
	if len(votableProposals) == 0 {
		fmt.Println("No proposal votes can be overridden.")
		return nil
	}

	// Get selected proposal
	var selectedProposal api.ProtocolDaoProposalDetails
	if c.Uint64(proposalFlag.Name) != 0 {
		// Get matching proposal
		selectedId := c.Uint64(proposalFlag.Name)
		found := false
		for _, proposal := range votableProposals {
			if proposal.ID == selectedId {
				selectedProposal = proposal
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Proposal %d can not be overridden.", selectedId)
		}
	} else {
		// Prompt for proposal selection
		options := make([]string, len(votableProposals))
		for pi, proposal := range votableProposals {
			options[pi] = fmt.Sprintf(
				"proposal %d (message: '%s', payload: %s, phase 2 end: %d, vp required: %.2f, for: %.2f, against: %.2f, abstained: %.2f, veto: %.2f, proposed by: %s)",
				proposal.ID,
				proposal.Message,
				proposal.PayloadStr,
				proposal.Phase2EndTime.Format(time.RFC822),
				eth.WeiToEth(proposal.VotingPowerRequired),
				eth.WeiToEth(proposal.VotingPowerFor),
				eth.WeiToEth(proposal.VotingPowerAgainst),
				eth.WeiToEth(proposal.VotingPowerAbstained),
				eth.WeiToEth(proposal.VotingPowerToVeto),
				proposal.ProposerAddress)
		}
		selected, _ := utils.Select("Please select a proposal to override:", options)
		selectedProposal = votableProposals[selected]
	}

	// Get support status
	var voteDirection types.VoteDirection
	var voteDirectionLabel string
	if c.String(voteDirectionFlag.Name) != "" {
		// Parse vote dirrection
		var err error
		voteDirection, err = input.ValidateVoteDirection("vote-direction", c.String(voteDirectionFlag.Name))
		if err != nil {
			return err
		}
		voteDirectionLabel = types.VoteDirections[voteDirection]
	} else {
		// Prompt for vote direction
		options := []string{
			"Abstain",
			"In Favor",
			"Against",
			"Veto",
		}
		var selected int
		selected, voteDirectionLabel = utils.Select("How would you like to vote on the proposal?", options)
		voteDirection = types.VoteDirection(selected + 1)
	}

	// Build the TX
	response, err := rp.Api.PDao.OverrideVoteOnProposal(selectedProposal.ID, voteDirection)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanVote {
		fmt.Println("Cannot override vote on proposal:")
		if response.Data.InsufficientPower {
			fmt.Println("You do not have any voting power.")
		}
		if response.Data.AlreadyVoted {
			fmt.Println("You already voted on this proposal.")
		}
		if response.Data.DoesNotExist {
			fmt.Println("The proposal does not exist.")
		}
		if response.Data.InvalidState {
			fmt.Println("The proposal is not in a voteable state.")
		}
		return nil
	}

	// Print voting power
	fmt.Printf("You currently have %.2f voting power.\n\n", eth.WeiToEth(response.Data.VotingPower))

	// Run the TX
	err = tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf("Are you sure you want to override your delegate's vote with your own vote for '%s' on proposal %d? Your vote cannot be changed later.", voteDirectionLabel, selectedProposal.ID),
		"overriding vote",
		"Overriding vote...",
	)

	// Log & return
	fmt.Printf("Successfully overrode delegate with a vote for '%s' on proposal %d.\n", voteDirectionLabel, selectedProposal.ID)
	return nil
}
