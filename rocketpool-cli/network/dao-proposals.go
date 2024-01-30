package network

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

func getActiveDAOProposals(c *cli.Context) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get the config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading configuration: %w", err)
	}

	// Print what network we're on
	err = cliutils.PrintNetwork(cfg.GetNetwork(), isNew)
	if err != nil {
		return err
	}

	// Get active DAO proposals
	snapshotProposalsResponse, err := rp.GetActiveDAOProposals()
	if err != nil {
		return err
	}

	currentVotingDelegate, err := rp.GetCurrentVotingDelegate()
	if err != nil {
		return err
	}

	// Voting status
	fmt.Printf("%s=== DAO Snapshot Voting ===%s\n", colorGreen, colorReset)
	blankAddress := common.Address{}
	if snapshotProposalsResponse.VotingDelegate == blankAddress {
		fmt.Println("The node does not currently have a voting delegate set, and will not be able to vote on Rocket Pool Snapshot governance proposals.")
	} else {
		fmt.Printf("The node has a voting delegate of %s%s%s which can represent it when voting on Rocket Pool Snapshot governance proposals.\n", colorBlue, snapshotProposalsResponse.VotingDelegate.Hex(), colorReset)
	}

	voteCount := 0
	for _, activeProposal := range snapshotProposalsResponse.ActiveSnapshotProposals {
		for _, votedProposal := range snapshotProposalsResponse.ProposalVotes {
			if votedProposal.Proposal.Id == activeProposal.Id {
				voteCount++
				break
			}
		}
	}
	if len(snapshotProposalsResponse.ActiveSnapshotProposals) == 0 {
		fmt.Print("Rocket Pool has no governance proposals being voted on.\n")
	} else {
		fmt.Printf("Rocket Pool has %d governance proposal(s) being voted on. You have voted on %d of those.\n", len(snapshotProposalsResponse.ActiveSnapshotProposals), voteCount)
	}

	for _, proposal := range snapshotProposalsResponse.ActiveSnapshotProposals {
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
			if proposal.ScoresTotal > proposal.Quorum {
				quorumResult += "✓"
			}
			fmt.Printf("Quorum: %.2f of %.2f needed %s\n", proposal.ScoresTotal, proposal.Quorum, quorumResult)
			voted := false
			for _, proposalVote := range snapshotProposalsResponse.ProposalVotes {
				if proposalVote.Proposal.Id == proposal.Id {
					voter := "Your DELEGATE"
					if proposalVote.Voter == snapshotProposalsResponse.AccountAddress {
						voter = "YOU"
					}
					votedChoices := ""
					switch proposalVote.Choice.(type) {
					case float64:
						choiceFloat := proposalVote.Choice.(float64)
						choice := int(choiceFloat) - 1
						if choice < len(proposal.Choices) && choice >= 0 {
							votedChoices = proposal.Choices[choice]
						} else {
							votedChoices = fmt.Sprintf("Unknown (%d is out of bounds)", choice)
						}

					case []interface{}:
						choicesArray := proposalVote.Choice.([]interface{})
						choices := []string{}
						for i := 0; i < len(choicesArray); i++ {
							choice := int(choicesArray[i].(float64))
							if choice < len(proposal.Choices) && choice >= 0 {
								choices = append(choices, proposal.Choices[choice])
							} else {
								choices = append(choices, fmt.Sprintf("Unknown (%d is out of bounds)", choice))
							}
						}
						votedChoices = strings.Join(choices, ", ")

					case map[string]interface{}:
						choiceMap := proposalVote.Choice.(map[string]interface{})
						choices := []string{}
						for choice, weight := range choiceMap {
							// choice here is 1-based
							choiceInt, _ := strconv.Atoi(choice)
							// here it is zero based, hence the -1
							choices = append(choices, fmt.Sprintf("%s: %.2f", proposal.Choices[choiceInt-1], weight))
						}
						votedChoices = strings.Join(choices, ", ")

					default:
						votedChoices = fmt.Sprintf("%v", proposalVote.Choice)
					}

					fmt.Printf("%s%s voted [%s] on this proposal\n%s", colorGreen, voter, votedChoices, colorReset)
					voted = true
				}
			}
			if !voted {
				fmt.Printf("%sYou have NOT voted on this proposal yet\n%s", colorYellow, colorReset)
			}
		}

	}
	// On-chain Voting status
	fmt.Println()
	fmt.Printf("%s=== DAO On-chain Voting ===%s\n", colorGreen, colorReset)
	if currentVotingDelegate.VotingDelegate == blankAddress {
		fmt.Println("The node does not currently have a voting delegate set, and will not be able to vote on Rocket Pool on-chain governance proposals.")
	} else {
		fmt.Printf("The node has a voting delegate of %s%s%s which can represent it when voting on Rocket Pool on-chain governance proposals.\n", colorBlue, currentVotingDelegate.VotingDelegate.Hex(), colorReset)
	}
	fmt.Println("")
	return nil
}
