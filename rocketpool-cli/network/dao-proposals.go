package network

import (
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/color"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func getActiveDAOProposals() error {
	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
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
	err = cliutils.PrintNetwork(cfg.GetNetworkInfo(), isNew)
	if err != nil {
		return err
	}

	// Get active DAO proposals
	proposalsResponse, err := rp.GetActiveDAOProposals()
	if err != nil {
		return err
	}

	// Voting status
	color.GreenPrintln("=== RocketDash Voting ===")
	blankAddress := common.Address{}
	if proposalsResponse.SignallingAddress == blankAddress {
		fmt.Printf("The node does not currently have an offchain signalling address set.\nTo learn more about offchain signalling, please visit %s.\n", signallingAddressLink)
	} else {
		fmt.Println("The node has a signalling address of", color.LightBlue(proposalsResponse.SignallingAddressFormatted), "which can represent it when voting on Rocket Pool governance proposals on RocketDash.")
	}

	if proposalsResponse.SnapshotResponse.Error != "" {
		fmt.Printf("Unable to fetch latest voting information from rocketdash.net: %s\n", proposalsResponse.SnapshotResponse.Error)
	} else {
		voteCount := proposalsResponse.SnapshotResponse.VoteCount()

		if len(proposalsResponse.SnapshotResponse.ActiveSnapshotProposals) == 0 {
			fmt.Print("Rocket Pool has no governance proposals being voted on.\n")
		} else {
			fmt.Printf("Rocket Pool has %d governance proposal(s) being voted on. You or your delegate have voted on %d of those.\n", len(proposalsResponse.SnapshotResponse.ActiveSnapshotProposals), voteCount)
		}

		for _, proposal := range proposalsResponse.SnapshotResponse.ActiveSnapshotProposals {
			fmt.Printf("\nTitle: %s\nVote: %s\n", proposal.Title, proposal.Link)
			currentTimestamp := time.Now().Unix()
			if currentTimestamp < proposal.Start {
				fmt.Printf("Start: %s (in %s)\n", cliutils.GetDateTimeString(uint64(proposal.Start)), time.Until(time.Unix(proposal.Start, 0)).Round(time.Second))
			} else {
				fmt.Printf("End: %s (in %s) \n", cliutils.GetDateTimeString(uint64(proposal.End)), time.Until(time.Unix(proposal.End, 0)).Round(time.Second))
				scoresBuilder := strings.Builder{}
				for i, score := range proposal.Scores {
					_, err = fmt.Fprintf(&scoresBuilder, "[%s = %.2f] ", proposal.Choices[i], score)
					if err != nil {
						return fmt.Errorf("error writing scores: %w", err)
					}
				}
				fmt.Printf("Scores: %s\n", scoresBuilder.String())
				quorumResult := ""
				if proposal.ScoresTotal >= proposal.Quorum {
					quorumResult += "✓"
				}
				fmt.Printf("Quorum: %.2f of %.2f needed %s\n", proposal.ScoresTotal, proposal.Quorum, quorumResult)
				voted := false
				for _, proposalVote := range proposalsResponse.SnapshotResponse.ProposalVotes {
					if proposalVote.Proposal.Id == proposal.Id {
						voter := "Your DELEGATE"
						if proposalVote.Voter == proposalsResponse.AccountAddress {
							voter = "YOU"
						}
						votedChoices := formatOffchainVote(proposalVote.Choice, proposal.Choices)

						color.GreenPrintf("%s voted [%s] on this proposal\n", voter, votedChoices)
						voted = true
					}
				}
				if !voted {
					color.YellowPrintln("You have NOT voted on this proposal yet")
				}
			}

		}
	}
	fmt.Println()

	// Onchain Voting Status
	color.GreenPrintln("=== Onchain Voting ===")

	switch proposalsResponse.OnchainVotingDelegate {
	case blankAddress:
		fmt.Println("The node doesn't have a delegate, which means it can vote directly on onchain proposals after it initializes voting.")
	case proposalsResponse.AccountAddress:
		fmt.Println("The node doesn't have a delegate, which means it can vote directly on onchain proposals. You can have another node represent you by running `rocketpool p svd <address>`.")
	default:
		fmt.Println("The node has a voting delegate of", color.LightBlue(proposalsResponse.OnchainVotingDelegateFormatted), "which can represent it when voting on Rocket Pool onchain governance proposals.")
	}
	fmt.Printf("The node's local voting power: %.10f\n", math.WeiToEth(proposalsResponse.VotingPower))

	if proposalsResponse.IsNodeRegistered {
		fmt.Printf("Total voting power delegated to the node: %.10f\n", math.WeiToEth(proposalsResponse.TotalDelegatedVp))
	} else {
		fmt.Println("The node must register using 'rocketpool node register' to be eligible to receive delegated voting power.")
	}

	fmt.Printf("Network total initialized voting power: %.4f\n", math.WeiToEth(proposalsResponse.SumVotingPower))
	fmt.Println()

	return nil
}
