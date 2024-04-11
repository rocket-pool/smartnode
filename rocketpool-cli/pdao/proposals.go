package pdao

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
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

	// Get Protocol DAO proposals
	allProposals, err := rp.PDAOProposals()
	if err != nil {
		return err
	}

	// Get proposals by state
	stateProposals := map[string][]api.PDAOProposalWithNodeVoteDirection{}
	for _, proposal := range allProposals.Proposals {
		stateName := types.ProtocolDaoProposalStates[proposal.State]
		if _, ok := stateProposals[stateName]; !ok {
			stateProposals[stateName] = []api.PDAOProposalWithNodeVoteDirection{}
		}
		stateProposals[stateName] = append(stateProposals[stateName], proposal)
	}

	// Proposal states print order
	proposalStates := []string{"Pending", "Active (Phase 1)", "Active (Phase 2)", "Succeeded", "Executed", "Destroyed", "Vetoed", "Quorum not Met", "Defeated", "Expired"}
	proposalStateInputs := []string{"pending", "phase1", "phase2", "succeeded", "executed", "destroyed", "vetoed", "quorum-not-met", "defeated", "expired"}

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
			fmt.Printf("%d: %s - Proposed by: %s\n", proposal.ID, proposal.Message, proposal.ProposerAddress)
		}

		count += len(proposals)

		fmt.Println()
	}
	if count == 0 {
		fmt.Println("There are no matching Protocol DAO proposals.")
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

	// Get protocol DAO proposals
	allProposals, err := rp.PDAOProposals()
	if err != nil {
		return err
	}

	// Find the proposal
	var proposal *api.PDAOProposalWithNodeVoteDirection

	for i, p := range allProposals.Proposals {
		if p.ID == id {
			proposal = &allProposals.Proposals[i]
			break
		}
	}

	if proposal == nil {
		fmt.Printf("Proposal with ID %d does not exist.\n", id)
		return nil
	}

	// Main details
	fmt.Printf("Proposal ID:            %d\n", proposal.ID)
	fmt.Printf("Message:                %s\n", proposal.Message)
	fmt.Printf("Payload:                %s\n", proposal.PayloadStr)
	fmt.Printf("Payload (bytes):        %s\n", hex.EncodeToString(proposal.Payload))
	fmt.Printf("Proposed by:            %s\n", proposal.ProposerAddress.Hex())
	fmt.Printf("Created at:             %s\n", proposal.CreatedTime.Format(time.RFC822))
	fmt.Printf("State:                  %s\n", types.ProtocolDaoProposalStates[proposal.State])

	// Start block - pending proposals
	if proposal.State == types.ProtocolDaoProposalState_Pending {
		fmt.Printf("Voting start:           %s\n", proposal.VotingStartTime.Format(time.RFC822))
	}
	if proposal.State == types.ProtocolDaoProposalState_Pending {
		fmt.Printf("Challenge window:       %s\n", proposal.ChallengeWindow)
	}

	// End block - active proposals
	if proposal.State == types.ProtocolDaoProposalState_ActivePhase1 {
		fmt.Printf("Phase 1 end:            %s\n", proposal.Phase1EndTime.Format(time.RFC822))
	}
	if proposal.State == types.ProtocolDaoProposalState_ActivePhase2 {
		fmt.Printf("Phase 2 end:            %s\n", proposal.Phase2EndTime.Format(time.RFC822))
	}

	// Expiry block - succeeded proposals
	if proposal.State == types.ProtocolDaoProposalState_Succeeded {
		fmt.Printf("Expires at:             %s\n", proposal.ExpiryTime.Format(time.RFC822))
	}

	// Vote details
	fmt.Printf("Voting power required:  %d\n", proposal.VotingPowerRequired)
	fmt.Printf("Voting power for:       %d\n", proposal.VotingPowerFor)
	fmt.Printf("Voting power against:   %d\n", proposal.VotingPowerAgainst)
	fmt.Printf("Voting power abstained: %d\n", proposal.VotingPowerAbstained)
	fmt.Printf("Voting power against:   %d\n", proposal.VotingPowerToVeto)
	if proposal.NodeVoteDirection != types.VoteDirection_NoVote {
		fmt.Printf("Node has voted:         %s\n", types.VoteDirections[proposal.NodeVoteDirection])
	} else {
		fmt.Printf("Node has voted:         no\n")
	}

	return nil
}
