package pdao

import (
	"fmt"
	"strconv"
	"time"

	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func voteOnProposal(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check for Houston
	houston, err := rp.IsHoustonDeployed()
	if err != nil {
		return fmt.Errorf("error checking if Houston has been deployed: %w", err)
	}
	if !houston.IsHoustonDeployed {
		fmt.Println("This command cannot be used until Houston has been deployed.")
		return nil
	}

	// Get oracle DAO proposals
	proposals, err := rp.PDAOProposals()
	if err != nil {
		return err
	}

	// Get votable proposals
	votableProposals := []api.PDAOProposalWithNodeVoteDirection{}
	for _, proposal := range proposals.Proposals {
		if (proposal.State == types.ProtocolDaoProposalState_ActivePhase1 || proposal.State == types.ProtocolDaoProposalState_ActivePhase2) && proposal.NodeVoteDirection == types.VoteDirection_NoVote {
			votableProposals = append(votableProposals, proposal)
		}
	}

	// Check for votable proposals
	if len(votableProposals) == 0 {
		fmt.Println("No proposals can be voted on.")
		return nil
	}

	// Get selected proposal
	var selectedProposal api.PDAOProposalWithNodeVoteDirection
	if c.String("proposal") != "" {

		// Get selected proposal ID
		selectedId, err := strconv.ParseUint(c.String("proposal"), 10, 64)
		if err != nil {
			return fmt.Errorf("Invalid proposal ID '%s': %w", c.String("proposal"), err)
		}

		// Get matching proposal
		found := false
		for _, proposal := range votableProposals {
			if proposal.ID == selectedId {
				selectedProposal = proposal
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Proposal %d can not be voted on.", selectedId)
		}

	} else {

		// Prompt for proposal selection
		options := make([]string, len(votableProposals))
		endTime := ""
		for pi, proposal := range votableProposals {
			if proposal.State == types.ProtocolDaoProposalState_ActivePhase1 {
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
	var voteDirection types.VoteDirection
	var voteDirectionLabel string
	if c.String("vote-direction") != "" {
		// Parse vote dirrection
		var err error
		voteDirection, err = cliutils.ValidateVoteDirection("vote-direction", c.String("vote-direction"))
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
		selected, voteDirectionLabel = cliutils.Select("How would you like to vote on the proposal?", options)
		voteDirection = types.VoteDirection(selected + 1)
	}
	canVote := api.CanVoteOnPDAOProposalResponse{}
	actionString := ""
	actionPast := ""

	if selectedProposal.State == types.ProtocolDaoProposalState_ActivePhase1 {
		// Check if proposal can be voted on
		actionString = "vote"
		actionPast = "voted"
		canVote, err = rp.PDAOCanVoteProposal(selectedProposal.ID, voteDirection)
		if err != nil {
			return err
		}
	} else {
		// Check if proposal can be overriden on
		actionString = "override your delegate's vote"
		actionPast = "overrode delegate with a vote for"
		canVote, err = rp.PDAOCanOverrideVote(selectedProposal.ID, voteDirection)
		if err != nil {
			return err
		}
	}

	if !canVote.CanVote {
		fmt.Printf("Cannot %s on proposal:\n", actionString)
		if canVote.InsufficientPower {
			fmt.Println("You didn't have voting power at the proposal snapshot.")
		}
		return nil
	}

	// Print the voting power
	fmt.Printf("\n\nYour voting power on this proposal: %.10f\n\n", eth.WeiToEth(canVote.VotingPower))

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canVote.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to %s '%s' on proposal %d? Your vote cannot be changed later.", actionString, voteDirectionLabel, selectedProposal.ID))) {
		fmt.Println("Cancelled.")
		return nil
	}

	response := api.VoteOnPDAOProposalResponse{}
	if selectedProposal.State == types.ProtocolDaoProposalState_ActivePhase1 {
		// Vote on proposal
		response, err = rp.PDAOVoteProposal(selectedProposal.ID, voteDirection)
		if err != nil {
			return err
		}
		fmt.Printf("Submitting vote...\n")
	} else {
		// Override vote on proposal
		response, err = rp.PDAOOverrideVote(selectedProposal.ID, voteDirection)
		if err != nil {
			return err
		}
		fmt.Printf("Overriding vote...\n")
	}

	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully %s '%s' for proposal %d.\n", actionPast, voteDirectionLabel, selectedProposal.ID)
	return nil

}
