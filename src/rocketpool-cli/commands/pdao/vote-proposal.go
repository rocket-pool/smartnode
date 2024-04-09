package pdao

import (
	"fmt"
	"time"

	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	cliutils "github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
	"github.com/rocket-pool/smartnode/v2/shared/utils"
)

func voteOnProposal(c *cli.Context) error {
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
		if (proposal.State == rptypes.ProtocolDaoProposalState_ActivePhase1 || proposal.State == rptypes.ProtocolDaoProposalState_ActivePhase2) && proposal.NodeVoteDirection == rptypes.VoteDirection_NoVote {
			votableProposals = append(votableProposals, proposal)
		}
	}

	// Check for votable proposals
	if len(votableProposals) == 0 {
		fmt.Println("No proposals can be voted on.")
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
			return fmt.Errorf("proposal %d can not be voted on", selectedId)
		}
	} else {
		// Prompt for proposal selection
		options := make([]string, len(votableProposals))
		endTime := ""
		for pi, proposal := range votableProposals {
			if proposal.State == rptypes.ProtocolDaoProposalState_ActivePhase1 {
				endTime = fmt.Sprintf("phase 1 end: %s", proposal.Phase1EndTime.Format(time.RFC822))
			} else {
				endTime = fmt.Sprintf("phase 2 end: %s", proposal.Phase2EndTime.Format(time.RFC822))
			}
			options[pi] = fmt.Sprintf(
				"proposal %d (message: '%s', payload: %s, %s, vp required: %.2f, for: %.2f, against: %.2f, abstained: %.2f, veto: %.2f, proposed by: %s)",
				proposal.ID,
				proposal.Message,
				proposal.PayloadStr,
				endTime,
				eth.WeiToEth(proposal.VotingPowerRequired),
				eth.WeiToEth(proposal.VotingPowerFor),
				eth.WeiToEth(proposal.VotingPowerAgainst),
				eth.WeiToEth(proposal.VotingPowerAbstained),
				eth.WeiToEth(proposal.VotingPowerToVeto),
				proposal.ProposerAddress)
		}
		selected, _ := cliutils.Select("Please select a proposal to vote on:", options)
		selectedProposal = votableProposals[selected]
	}

	// Get support status
	var voteDirection rptypes.VoteDirection
	var voteDirectionLabel string
	if c.String(voteDirectionFlag.Name) != "" {
		// Parse vote dirrection
		var err error
		voteDirection, err = utils.ValidateVoteDirection("vote-direction", c.String(voteDirectionFlag.Name))
		if err != nil {
			return err
		}
		voteDirectionLabel = rptypes.VoteDirections[voteDirection]
	} else {
		// Prompt for vote direction
		options := []string{
			"Abstain",
			"In Favor",
			"Against",
			"Veto",
		}
		var selected int
		selected, voteDirectionLabel = cliutils.Select("How would you like to vote on the proposal?", options)
		voteDirection = rptypes.VoteDirection(selected + 1)
	}

	// Build the TX
	actionString := ""
	actionPast := ""
	var response *types.ApiResponse[api.ProtocolDaoVoteOnProposalData]
	if selectedProposal.State == rptypes.ProtocolDaoProposalState_ActivePhase1 {
		// Check if proposal can be voted on
		actionString = "vote"
		actionPast = "voted"
		response, err = rp.Api.PDao.VoteOnProposal(selectedProposal.ID, voteDirection)
	} else {
		// Check if proposal can be overriden on
		actionString = "override your delegate's vote"
		actionPast = "overrode delegate with a vote for"
		response, err = rp.Api.PDao.OverrideVoteOnProposal(selectedProposal.ID, voteDirection)
	}
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanVote {
		fmt.Printf("Cannot %s on proposal:\n", actionString)
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
	fmt.Printf("You currently have %s voting power.\n\n", response.Data.VotingPower.String())

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf("Are you sure you want to %s '%s' on proposal %d? Your vote cannot be changed later.", actionString, voteDirectionLabel, selectedProposal.ID),
		"voting on proposal",
		"Submitting vote...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("Successfully %s '%s' for proposal %d.\n", actionPast, voteDirectionLabel, selectedProposal.ID)
	return nil
}
