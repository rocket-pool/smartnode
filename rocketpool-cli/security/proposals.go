package security

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/rocket-pool/rocketpool-go/dao"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

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
	allProposals, err := rp.SecurityProposals()
	if err != nil {
		return err
	}

	// Get security council members
	allMembers, err := rp.SecurityMembers()
	if err != nil {
		return err
	}

	// Get proposals by state
	stateProposals := map[string][]dao.ProposalDetails{}
	for _, proposal := range allProposals.Proposals {
		stateName := proposal.State.String()
		if _, ok := stateProposals[stateName]; !ok {
			stateProposals[stateName] = []dao.ProposalDetails{}
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
			for _, member := range allMembers.Members {
				if bytes.Equal(proposal.ProposerAddress.Bytes(), member.Address.Bytes()) {
					fmt.Printf("%d: %s - Proposed by: %s (%s)\n", proposal.ID, proposal.Message, member.ID, proposal.ProposerAddress)
					printed = true
				}
			}
			if !printed {
				fmt.Printf("%d: %s - Proposed by: %s (no longer on the Security Council)\n", proposal.ID, proposal.Message, proposal.ProposerAddress)
			}
		}

		count += len(proposals)

		fmt.Println()
	}
	if count == 0 {
		fmt.Println("There are no matching Security Council proposals.")
	}
	return nil

}

func getProposal(c *cli.Context, id uint64) error {
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
	allProposals, err := rp.SecurityProposals()
	if err != nil {
		return err
	}

	// Get security council members
	allMembers, err := rp.SecurityMembers()
	if err != nil {
		return err
	}

	// Find the proposal
	var proposal *dao.ProposalDetails

	for i, p := range allProposals.Proposals {
		if p.ID == id {
			proposal = &allProposals.Proposals[i]
			break
		}
	}

	// Find the proposer
	var memberID string
	for _, member := range allMembers.Members {
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
	fmt.Printf("Created at:           %s\n", cliutils.GetDateTimeString(proposal.CreatedTime))

	// Start block - pending proposals
	if proposal.State == types.Pending {
		fmt.Printf("Starts at:            %s\n", cliutils.GetDateTimeString(proposal.StartTime))
	}

	// End block - active proposals
	if proposal.State == types.Active {
		fmt.Printf("Ends at:              %s\n", cliutils.GetDateTimeString(proposal.EndTime))
	}

	// Expiry block - succeeded proposals
	if proposal.State == types.Succeeded {
		fmt.Printf("Expires at:           %s\n", cliutils.GetDateTimeString(proposal.ExpiryTime))
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
