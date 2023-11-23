package security

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/rocket-pool/rocketpool-go/dao"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
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

	// Get security council proposals
	proposals, err := rp.SecurityProposals()
	if err != nil {
		return err
	}

	// Get security council members
	allMembers, err := rp.SecurityMembers()
	if err != nil {
		return err
	}

	// Get votable proposals
	votableProposals := []dao.ProposalDetails{}
	for _, proposal := range proposals.Proposals {
		if proposal.State == types.Active && !proposal.MemberVoted {
			votableProposals = append(votableProposals, proposal)
		}
	}

	// Check for votable proposals
	if len(votableProposals) == 0 {
		fmt.Println("No proposals can be voted on.")
		return nil
	}

	// Get selected proposal
	var selectedProposal dao.ProposalDetails
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
		var memberID string
		options := make([]string, len(votableProposals))
		for pi, proposal := range votableProposals {
			for _, member := range allMembers.Members {
				if bytes.Equal(proposal.ProposerAddress.Bytes(), member.Address.Bytes()) {
					memberID = member.ID
				}
			}
			options[pi] = fmt.Sprintf(
				"proposal %d (message: '%s', payload: %s, end time: %s, votes required: %.2f, votes for: %.2f, votes against: %.2f, proposed by: %s (%s))",
				proposal.ID,
				proposal.Message,
				proposal.PayloadStr,
				cliutils.GetDateTimeString(proposal.EndTime),
				proposal.VotesRequired,
				proposal.VotesFor,
				proposal.VotesAgainst,
				memberID,
				proposal.ProposerAddress)
		}
		selected, _ := cliutils.Select("Please select a proposal to vote on:", options)
		selectedProposal = votableProposals[selected]

	}

	// Get support status
	var support bool
	var supportLabel string
	if c.String("support") != "" {

		// Parse support status
		var err error
		support, err = cliutils.ValidateBool("support", c.String("support"))
		if err != nil {
			return err
		}

	} else {

		// Prompt for support status
		support = cliutils.Confirm("Would you like to vote in support of the proposal?")

	}
	if support {
		supportLabel = "in support of"
	} else {
		supportLabel = "against"
	}

	// Check if proposal can be voted on
	canVote, err := rp.SecurityCanVoteOnProposal(selectedProposal.ID)
	if err != nil {
		return err
	}
	if !canVote.CanVote {
		fmt.Println("Cannot vote on proposal:")
		if canVote.JoinedAfterCreated {
			fmt.Println("You cannot vote on proposals created before you joined the security council.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canVote.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to vote %s proposal %d? Your vote cannot be changed later.", supportLabel, selectedProposal.ID))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Vote on proposal
	response, err := rp.SecurityVoteOnProposal(selectedProposal.ID, support)
	if err != nil {
		return err
	}

	fmt.Printf("Submitting vote...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully voted %s proposal %d.\n", supportLabel, selectedProposal.ID)
	return nil

}
