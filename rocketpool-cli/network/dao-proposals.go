package network

import (
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

func getActiveDAOProposals(c *cli.Context) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check and assign the EC status
	err = cliutils.CheckClientStatus(rp)
	if err != nil {
		return err
	}

	// Print what network we're on
	err = cliutils.PrintNetwork(rp)
	if err != nil {
		return err
	}

	// Get node status
	status, err := rp.NodeStatus()
	if err != nil {
		return err
	}

	// // Get the config
	// cfg, _, err := rp.LoadConfig()
	// if err != nil {
	// 	return fmt.Errorf("Error loading configuration: %w", err)
	// }

	// Voting status
	fmt.Printf("%s=== DAO Voting ===%s\n", colorGreen, colorReset)
	blankAddress := common.Address{}
	if status.VotingDelegate == blankAddress {
		fmt.Println("The node does not currently have a voting delegate set, and will not be able to vote on Rocket Pool governance proposals.")
	} else {
		fmt.Printf("The node has a voting delegate of %s%s%s which can represent it when voting on Rocket Pool governance proposals.\n", colorBlue, status.VotingDelegate.Hex(), colorReset)
	}

	voteCount := 0
	for _, activeProposal := range status.ActiveSnapshotProposals {
		for _, votedProposal := range status.ProposalVotes {
			if votedProposal.Proposal.Id == activeProposal.Id {
				voteCount++
				break
			}
		}
	}
	if len(status.ActiveSnapshotProposals) == 0 {
		fmt.Print("Rocket Pool has no governance proposals being voted.\n")
	} else {
		fmt.Printf("Rocket Pool has %d governance proposal(s) being voted. You have voted on %d of those. See details using 'rocketpool network dao-proposals'.\n", len(status.ActiveSnapshotProposals), voteCount)
	}

	for _, proposal := range status.ActiveSnapshotProposals {
		fmt.Printf("\nTitle: %s\n", proposal.Title)
		currentTimestamp := time.Now().Unix()
		if currentTimestamp < proposal.Start {
			fmt.Printf("Start: %s (in %s)\n", cliutils.GetDateTimeString(uint64(proposal.Start)), time.Until(time.Unix(proposal.Start, 0)).Round(time.Second))
		} else {
			fmt.Printf("End: %s (in %s) \n", cliutils.GetDateTimeString(uint64(proposal.End)), time.Until(time.Unix(proposal.End, 0)).Round(time.Second))
			scoresBuilder := strings.Builder{}
			for i, score := range proposal.Scores {
				scoresBuilder.WriteString(fmt.Sprintf("[%s = %.2f] ", proposal.Choices[i], score))
			}
			fmt.Printf("Scores: %s\n", scoresBuilder.String())
			quorumResult := ""
			if proposal.ScoresTotal > float64(proposal.Quorum) {
				quorumResult += "✓"
			}
			fmt.Printf("Quorum: %.2f of %d needed %s\n", proposal.ScoresTotal, proposal.Quorum, quorumResult)
			voted := false
			for _, proposalVote := range status.ProposalVotes {
				if proposalVote.Proposal.Id == proposal.Id {
					voter := "Your DELEGATE"
					if proposalVote.Voter == status.AccountAddress {
						voter = "YOU"
					}
					fmt.Printf("%s%s voted [%s] on this proposal\n%s", colorGreen, voter, proposal.Choices[proposalVote.Choice-1], colorReset)
					voted = true
				}
			}
			if !voted {
				fmt.Printf("%sYou have NOT voted on this proposal yet\n%s", colorYellow, colorReset)
			}
		}
	}
	fmt.Println("")
	return nil
}
