package network

import (
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	cliutils "github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	sharedtypes "github.com/rocket-pool/smartnode/v2/shared/types"
	"github.com/urfave/cli/v2"
)

func getActiveDAOProposals(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get the config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Print what network we're on
	err = cliutils.PrintNetwork(cfg.Network.Value, isNew)
	if err != nil {
		return err
	}

	// Get active DAO proposals
	snapshotProposalsResponse, err := rp.Api.Network.GetActiveDaoProposals()
	if err != nil {
		return err
	}

	currentVotingDelegate, err := rp.Api.PDao.GetCurrentVotingDelegate()
	if err != nil {
		return err
	}

	// Voting status
	fmt.Printf("%s=== DAO Snapshot Voting ===%s\n", terminal.ColorGreen, terminal.ColorReset)
	blankAddress := common.Address{}
	if snapshotProposalsResponse.Data.VotingDelegate == blankAddress {
		fmt.Println("The node does not currently have a voting delegate set, and will not be able to vote on Rocket Pool Snapshot governance proposals.")
	} else {
		fmt.Printf("The node has a voting delegate of %s%s%s which can represent it when voting on Rocket Pool Snapshot governance proposals.\n", terminal.ColorBlue, snapshotProposalsResponse.Data.VotingDelegate.Hex(), terminal.ColorReset)
	}

	voteCount := 0
	for _, activeProposal := range snapshotProposalsResponse.Data.ActiveSnapshotProposals {
		if len(activeProposal.DelegateVotes) > 0 || len(activeProposal.UserVotes) > 0 {
			voteCount++
			break
		}
	}
	if len(snapshotProposalsResponse.Data.ActiveSnapshotProposals) == 0 {
		fmt.Print("Rocket Pool has no governance proposals being voted on.\n")
	} else {
		fmt.Printf("Rocket Pool has %d governance proposal(s) being voted on. You have voted on %d of those.\n", len(snapshotProposalsResponse.Data.ActiveSnapshotProposals), voteCount)
	}

	for _, proposal := range snapshotProposalsResponse.Data.ActiveSnapshotProposals {
		fmt.Printf("\nTitle: %s\n", proposal.Title)
		currentTimestamp := time.Now()
		if currentTimestamp.Before(proposal.Start) {
			fmt.Printf("Start: %s (in %s)\n", cliutils.GetDateTimeStringOfTime(proposal.Start), time.Until(proposal.Start).Round(time.Second))
		} else {
			fmt.Printf("End: %s (in %s) \n", cliutils.GetDateTimeStringOfTime(proposal.End), time.Until(proposal.End).Round(time.Second))

			// Scores
			var totalScores float64
			scoresBuilder := strings.Builder{}
			for i, score := range proposal.Scores {
				totalScores += score
				scoresBuilder.WriteString(fmt.Sprintf("[%s = %.2f] ", proposal.Choices[i], score))
			}
			fmt.Printf("Scores: %s\n", scoresBuilder.String())
			quorumResult := ""
			if totalScores > proposal.Quorum {
				quorumResult += "✓"
			}
			fmt.Printf("Quorum: %.2f of %.2f needed %s\n", totalScores, proposal.Quorum, quorumResult)

			// Vote status
			var voted bool
			delegateVotes := getVoteString(proposal.DelegateVotes, proposal)
			nodeVotes := getVoteString(proposal.UserVotes, proposal)

			if delegateVotes != "" {
				fmt.Printf("%sYour DELEGATE voted [%s] on this proposal\n%s", terminal.ColorGreen, delegateVotes, terminal.ColorReset)
				voted = true
			}
			if nodeVotes != "" {
				fmt.Printf("%sYOU voted [%s] on this proposal\n%s", terminal.ColorGreen, delegateVotes, terminal.ColorReset)
				voted = true
			}
			if !voted {
				fmt.Printf("%sYou have NOT voted on this proposal yet\n%s", terminal.ColorYellow, terminal.ColorReset)
			}
		}
	}

	// On-chain Voting status
	fmt.Println()
	fmt.Printf("%s=== DAO On-chain Voting ===%s\n", terminal.ColorGreen, terminal.ColorReset)
	if currentVotingDelegate.Data.VotingDelegate == blankAddress {
		fmt.Println("The node does not currently have a voting delegate set, and will not be able to vote on Rocket Pool on-chain governance proposals.")
	} else {
		fmt.Printf("The node has a voting delegate of %s%s%s which can represent it when voting on Rocket Pool on-chain governance proposals.\n", terminal.ColorBlue, currentVotingDelegate.Data.VotingDelegate.Hex(), terminal.ColorReset)
	}

	fmt.Println("")
	return nil
}

func getVoteString(votes []int, proposal *sharedtypes.SnapshotProposal) string {
	if len(votes) == 0 {
		return ""
	}

	choices := []string{}
	for _, vote := range votes {
		var choice string
		if vote == -1 {
			choice = "<deserialization error>"
		} else {
			choice = proposal.Choices[vote]
		}
		choices = append(choices, choice)
	}
	return strings.Join(choices, ", ")
}
