package pdao

import (
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/types"
	utilsStrings "github.com/rocket-pool/smartnode/bindings/utils/strings"

	"github.com/rocket-pool/smartnode/shared/math"
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
	return !slices.Contains(filterStates, state)
}

func getProposals(stateFilter string) error {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

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
			if len(proposal.Message) > 200 {
				proposal.Message = proposal.Message[:200]
			}
			proposal.Message = utilsStrings.Sanitize(proposal.Message)
			fmt.Printf("%d: %s - Proposed by: %s\n", proposal.ID, proposal.Message, proposal.ProposerAddress)
			if summary := formatMultiSettingsSummary(proposal.MultiSettings); summary != "" {
				fmt.Printf("    %s\n", summary)
			}
		}

		count += len(proposals)

		fmt.Println()
	}
	if count == 0 {
		fmt.Println("There are no matching Protocol DAO proposals.")
	}
	return nil

}

func getProposal(id uint64) error {
	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get protocol DAO proposals
	allProposals, err := rp.PDAOProposals()
	if err != nil {
		return err
	}

	// Get the voting delegate info
	votingDelegateInfo, err := rp.GetCurrentVotingDelegate()
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

	proposal.Message = utilsStrings.Sanitize(proposal.Message)

	// Main details
	fmt.Printf("Proposal ID:            %d\n", proposal.ID)
	fmt.Printf("Message:                %s\n", proposal.Message)
	printProposalPayload(proposal.ProtocolDaoProposalDetails)
	fmt.Printf("Proposed by:            %s\n", proposal.ProposerAddress.Hex())
	fmt.Printf("Created at:             %s, %s\n", proposal.CreatedTime.Format(time.RFC822), getTimeDifference(proposal.CreatedTime))
	fmt.Printf("State:                  %s\n", types.ProtocolDaoProposalStates[proposal.State])

	// Start block - pending proposals
	if proposal.State == types.ProtocolDaoProposalState_Pending {
		fmt.Printf("Voting start:           %s, %s\n", proposal.VotingStartTime.Format(time.RFC822), getTimeDifference(proposal.VotingStartTime))
	}
	if proposal.State == types.ProtocolDaoProposalState_Pending {
		fmt.Printf("Challenge window:       %s\n", proposal.ChallengeWindow)
	}

	// End block - active proposals
	if proposal.State == types.ProtocolDaoProposalState_ActivePhase1 {
		fmt.Printf("Phase 1 end:            %s, %s\n", proposal.Phase1EndTime.Format(time.RFC822), getTimeDifference(proposal.Phase1EndTime))
	}
	if proposal.State == types.ProtocolDaoProposalState_ActivePhase2 {
		fmt.Printf("Phase 2 end:            %s, %s\n", proposal.Phase2EndTime.Format(time.RFC822), getTimeDifference(proposal.Phase2EndTime))
	}

	// Expiry block - succeeded proposals
	if proposal.State == types.ProtocolDaoProposalState_Succeeded {
		fmt.Printf("Expires at:             %s, %s\n", proposal.ExpiryTime.Format(time.RFC822), getTimeDifference(proposal.ExpiryTime))
	}

	// Vote details
	votingPowerFor := math.RoundDown(math.WeiToEth(proposal.VotingPowerFor), 2)
	votingPowerRequired := math.RoundUp(math.WeiToEth(proposal.VotingPowerRequired), 2)
	votingPowerToVeto := math.RoundDown(math.WeiToEth(proposal.VotingPowerToVeto), 2)
	vetoQuorum := math.RoundUp(math.WeiToEth(proposal.VetoQuorum), 2)
	fmt.Printf("Voting power for:       %.2f / %.2f (%.2f%%)\n", votingPowerFor, votingPowerRequired, votingPowerFor/votingPowerRequired*100)
	fmt.Printf("Voting power against:   %.2f\n", math.RoundDown(math.WeiToEth(proposal.VotingPowerAgainst), 2))
	fmt.Printf("Against with veto:      %.2f / %2.f (%.2f%%)\n", votingPowerToVeto, vetoQuorum, votingPowerToVeto/vetoQuorum*100)
	fmt.Printf("Voting power abstained: %.2f\n", math.RoundDown(math.WeiToEth(proposal.VotingPowerAbstained), 2))
	if proposal.NodeVoteDirection != types.VoteDirection_NoVote {
		fmt.Printf("Node has voted:         %s\n", types.VoteDirections[proposal.NodeVoteDirection])
	} else {
		fmt.Printf("Node has voted:         no\n")
	}

	if votingDelegateInfo.VotingDelegate != votingDelegateInfo.AccountAddress && proposal.DelegateVoteDirection != types.VoteDirection_NoVote {
		fmt.Printf("Delegate has voted:     %s\n", types.VoteDirections[proposal.DelegateVoteDirection])
	} else if votingDelegateInfo.VotingDelegate != votingDelegateInfo.AccountAddress {
		fmt.Printf("Delegate has voted:     no\n")
	}
	if votingDelegateInfo.VotingDelegate != votingDelegateInfo.AccountAddress {
		fmt.Printf("Current Delegate:       %s", votingDelegateInfo.VotingDelegate.Hex())
	}

	return nil
}

func getTimeDifference(t time.Time) string {
	// Get the current time
	currentTime := time.Now()

	// Calculate the time difference
	timeDiff := currentTime.Sub(t)

	// Round timeDiff to the nearest whole second
	roundedDuration := time.Duration(int64(timeDiff.Seconds() + 0.5))
	timeDiff = time.Duration(roundedDuration) * time.Second

	// Absolute value
	absTimeDiff := time.Duration(math.Abs(float64(timeDiff)))

	var message string

	if timeDiff < 0 {
		message = fmt.Sprintf("in %s", absTimeDiff)
	} else {
		message = fmt.Sprintf("was %s ago", timeDiff)
	}

	return message
}

func printProposalPayload(proposal protocol.ProtocolDaoProposalDetails) {
	if len(proposal.MultiSettings) > 0 {
		fmt.Printf("Payload:                proposalSettingMulti (%d settings)\n", len(proposal.MultiSettings))
		printMultiSettings(proposal.MultiSettings, "                        ")
	} else {
		fmt.Printf("Payload:                %s\n", proposal.PayloadStr)
	}
	fmt.Printf("Payload (bytes):        %s\n", hex.EncodeToString(proposal.Payload))
}

func printMultiSettings(settings []protocol.DecodedProposalSetting, indent string) {
	for i, setting := range settings {
		fmt.Printf("%s%d. %s / %s = %s (%s)\n", indent, i+1, setting.Contract, setting.Path, setting.Value, setting.Type)
	}
}

func formatMultiSettingsSummary(settings []protocol.DecodedProposalSetting) string {
	if len(settings) == 0 {
		return ""
	}
	return protocol.FormatProposalSettingMulti(settings)
}

func proposalDisplayText(proposal api.PDAOProposalWithNodeVoteDirection) (message string, payload string) {
	message = proposal.Message
	if len(message) > 200 {
		message = message[:200]
	}
	message = utilsStrings.Sanitize(message)

	payload = proposal.PayloadStr
	if len(proposal.MultiSettings) > 0 {
		payload = formatMultiSettingsSummary(proposal.MultiSettings)
	}
	if len(payload) > 200 {
		payload = payload[:200] + "..."
	}
	return message, payload
}

func printSelectedMultiSettings(settings []protocol.DecodedProposalSetting) {
	if len(settings) == 0 {
		return
	}
	fmt.Printf("This proposal updates %d settings:\n", len(settings))
	printMultiSettings(settings, "  ")
	fmt.Println()
}
