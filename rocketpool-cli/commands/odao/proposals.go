package odao

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

var proposalStatesFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "states",
	Aliases: []string{"s"},
	Usage:   "Comma separated list of states to filter ('pending', 'active', 'succeeded', 'executed', 'cancelled', 'defeated', or 'expired')",
}

func filterProposalState(state string, stateFilter string) bool {
	// Easy out
	if stateFilter == "" {
		return false
	}

	// Check comma separated list for the state
	filterStates := strings.Split(stateFilter, ",")
	for _, fs := range filterStates {
		if fs == state {
			return false
		}
	}

	// Not found
	return true
}

func getProposals(c *cli.Context, stateFilter string) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get oracle DAO proposals
	allProposals, err := rp.Api.ODao.Proposals()
	if err != nil {
		return err
	}

	// Get oracle DAO members
	allMembers, err := rp.Api.ODao.Members()
	if err != nil {
		return err
	}

	// Get proposals by state
	stateProposals := map[string][]api.OracleDaoProposalDetails{}
	for _, proposal := range allProposals.Data.Proposals {
		stateName := proposal.State.String()
		if _, ok := stateProposals[stateName]; !ok {
			stateProposals[stateName] = []api.OracleDaoProposalDetails{}
		}
		stateProposals[stateName] = append(stateProposals[stateName], proposal)
	}

	// Proposal states print order
	proposalStates := []string{"Pending", "Active", "Succeeded", "Executed", "Cancelled", "Defeated", "Expired"}
	proposalStateInputs := []string{"pending", "active", "succeeded", "executed", "cancelled", "defeated", "expired"}

	// Print & return
	count := 0
	for i, stateName := range proposalStates {
		proposals, ok := stateProposals[stateName]
		if !ok {
			continue
		}

		// Check filter
		if filterProposalState(proposalStateInputs[i], stateFilter) {
			continue
		}

		// Proposal state count
		fmt.Printf("%d %s proposal(s):\n", len(proposals), stateName)
		fmt.Println("")

		// Proposals
		for _, proposal := range proposals {
			printed := false
			for _, member := range allMembers.Data.Members {
				if bytes.Equal(proposal.ProposerAddress.Bytes(), member.Address.Bytes()) {
					fmt.Printf("%d: %s - Proposed by: %s (%s)\n", proposal.ID, proposal.Message, member.ID, proposal.ProposerAddress)
					printed = true
					break
				}
			}
			if !printed {
				fmt.Printf("%d: %s - Proposed by: %s (no longer on the Oracle DAO)\n", proposal.ID, proposal.Message, proposal.ProposerAddress)
			}
		}

		count += len(proposals)

		fmt.Println()
	}
	if count == 0 {
		fmt.Println("There are no matching Oracle DAO proposals.")
	}
	return nil
}

func getProposal(c *cli.Context, id uint64) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get oracle DAO proposals
	allProposals, err := rp.Api.ODao.Proposals()
	if err != nil {
		return err
	}

	// Get oracle DAO members
	allMembers, err := rp.Api.ODao.Members()
	if err != nil {
		return err
	}

	// Find the proposal
	var proposal *api.OracleDaoProposalDetails

	for i, p := range allProposals.Data.Proposals {
		if p.ID == id {
			proposal = &allProposals.Data.Proposals[i]
			break
		}
	}

	// Find the proposer
	var memberID string
	for _, member := range allMembers.Data.Members {
		if bytes.Equal(proposal.ProposerAddress.Bytes(), member.Address.Bytes()) {
			memberID = member.ID
		}
	}

	if proposal == nil {
		fmt.Printf("Proposal with ID %d does not exist.\n", id)
		return nil
	}

	// Main details
	fmt.Printf("Proposal ID:          %d\n", proposal.ID)
	fmt.Printf("Message:              %s\n", proposal.Message)
	fmt.Printf("Payload:              %s\n", proposal.PayloadStr)
	fmt.Printf("Payload (bytes):      %s\n", hex.EncodeToString(proposal.Payload))
	fmt.Printf("Proposed by:          %s (%s)\n", memberID, proposal.ProposerAddress.Hex())
	fmt.Printf("Created at:           %s\n", utils.GetDateTimeStringOfTime(proposal.CreatedTime))

	// Start block - pending proposals
	if proposal.State == types.ProposalState_Pending {
		fmt.Printf("Starts at:            %s\n", utils.GetDateTimeStringOfTime(proposal.StartTime))
	}

	// End block - active proposals
	if proposal.State == types.ProposalState_Active {
		fmt.Printf("Ends at:              %s\n", utils.GetDateTimeStringOfTime(proposal.EndTime))
	}

	// Expiry block - succeeded proposals
	if proposal.State == types.ProposalState_Succeeded {
		fmt.Printf("Expires at:           %s\n", utils.GetDateTimeStringOfTime(proposal.ExpiryTime))
	}

	// Vote details
	fmt.Printf("Votes required:       %.2f\n", proposal.VotesRequired)
	fmt.Printf("Votes for:            %.2f\n", proposal.VotesFor)
	fmt.Printf("Votes against:        %.2f\n", proposal.VotesAgainst)
	if proposal.MemberVoted {
		if proposal.MemberSupported {
			fmt.Printf("Node has voted:       for\n")
		} else {
			fmt.Printf("Node has voted:       against\n")
		}
	} else {
		fmt.Printf("Node has voted:       no\n")
	}

	return nil
}
