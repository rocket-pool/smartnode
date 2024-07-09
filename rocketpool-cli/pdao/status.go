package pdao

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/rocketpool-go/utils/strings"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

const (
	colorBlue             string = "\033[36m"
	colorReset            string = "\033[0m"
	colorGreen            string = "\033[32m"
	signallingAddressLink string = "https://docs.rocketpool.net/guides/houston/participate#setting-your-snapshot-signalling-address"
	challengeLink         string = "https://docs.rocketpool.net/guides/houston/pdao#challenge-process"
)

func getStatus(c *cli.Context) error {

	// Get RP client
	rp := rocketpool.NewClientFromCtx(c)
	defer rp.Close()

	// Get wallet status
	walletStatus, err := rp.WalletStatus()
	if err != nil {
		return err
	}

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

	// rp.PDAOStatus() will fail with an error, but we can short-circuit it here.
	if !walletStatus.WalletInitialized {
		return errors.New("The node wallet is not initialized.")
	}

	// Get PDAO status at the latest block
	response, err := rp.PDAOStatus()
	if err != nil {
		return err
	}

	// Get Protocol DAO proposals
	allProposals, err := rp.PDAOProposals()
	if err != nil {
		return err
	}

	// Get protocol DAO proposals
	claimableBondsResponse, err := rp.PDAOGetClaimableBonds()
	if err != nil {
		return fmt.Errorf("error checking for claimable bonds: %w", err)
	}
	claimableBonds := claimableBondsResponse.ClaimableBonds

	// Signalling Status
	fmt.Printf("%s=== Signalling on Snapshot ===%s\n", colorGreen, colorReset)
	blankAddress := common.Address{}
	if response.SignallingAddress == blankAddress {
		fmt.Printf("The node does not currently have a snapshot signalling address set.\nTo learn more about snapshot signalling, please visit %s.\n", signallingAddressLink)
	} else {
		fmt.Printf("The node can vote directly or override their delegate of %s%s%s which can represent it when voting on Rocket Pool Snapshot governance proposals.\n", colorBlue, response.SignallingAddressFormatted, colorReset)
	}

	if response.SnapshotResponse.Error != "" {
		fmt.Printf("Unable to fetch latest voting information from snapshot.org: %s\n", response.SnapshotResponse.Error)
	} else {
		voteCount := 0
		for _, activeProposal := range response.SnapshotResponse.ActiveSnapshotProposals {
			for _, votedProposal := range response.SnapshotResponse.ProposalVotes {
				if votedProposal.Proposal.Id == activeProposal.Id {
					voteCount++
					break
				}
			}
		}
		if len(response.SnapshotResponse.ActiveSnapshotProposals) == 0 {
			fmt.Print("Rocket Pool has no Snapshot governance proposals being voted on.\n")
		} else {
			fmt.Printf("Rocket Pool has %d Snapshot governance proposal(s) being voted on. You have voted on %d of those. See details using 'rocketpool network dao-proposals'.\n", len(response.SnapshotResponse.ActiveSnapshotProposals), voteCount)
		}
		fmt.Println("")
	}

	// Onchain Voting Status
	fmt.Printf("%s=== Onchain Voting ===%s\n", colorGreen, colorReset)
	if response.IsVotingInitialized {
		fmt.Printf("The node %s%s%s has been initialized for onchain voting.\n", colorBlue, response.AccountAddressFormatted, colorReset)
	} else {
		fmt.Printf("The node %s%s%s has NOT been initialized for onchain voting. You need to run `rocketpool pdao initialize-voting` to participate in onchain votes.\n", colorBlue, response.AccountAddressFormatted, colorReset)
	}

	if response.OnchainVotingDelegate == blankAddress {
		fmt.Println("The node doesn't have a delegate, which means it can vote directly on onchain proposals after it initializes voting.")
	} else if response.OnchainVotingDelegate == response.AccountAddress {
		fmt.Println("The node doesn't have a delegate, which means it can vote directly on onchain proposals. You can have another node represent you by running `rocketpool p svd <address>`.")
	} else {
		fmt.Printf("The node has a signalling address of %s%s%s which can represent it when voting on Rocket Pool onchain governance proposals.\n", colorBlue, response.OnchainVotingDelegateFormatted, colorReset)
	}
	fmt.Printf("The node's local voting power: %.10f\n", eth.WeiToEth(response.VotingPower))

	if response.IsNodeRegistered {
		fmt.Printf("Total voting power delegated to the node: %.10f\n", eth.WeiToEth(response.TotalDelegatedVp))
	} else {
		fmt.Print("The node must register using 'rocketpool node register' to be eligible to receive delegated voting power.\n")
	}

	fmt.Printf("Network total initialized voting power: %.10f\n", eth.WeiToEth(response.SumVotingPower))
	fmt.Println("")

	// Claimable Bonds Status:
	fmt.Printf("%s=== Claimable RPL Bonds ===%s\n", colorGreen, colorReset)
	if response.IsRPLLockingAllowed {
		fmt.Print("The node is allowed to lock RPL to create governance proposals/challenges.\n")
		if response.NodeRPLLocked.Cmp(big.NewInt(0)) != 0 {
			fmt.Printf("The node currently has %.6f RPL locked.\n",
				math.RoundDown(eth.WeiToEth(response.NodeRPLLocked), 6))
		}

	} else {
		fmt.Print("The node is NOT allowed to lock RPL to create governance proposals/challenges. Use 'rocketpool node allow-rpl-locking, to allow RPL locking.\n")
	}
	if len(claimableBonds) == 0 {
		fmt.Println("You do not have any unlockable bonds or claimable rewards.")
	} else {
		fmt.Println("The node has unlockable bonds or claimable rewards available. Use 'rocketpool pdao claim-bonds' to view and claim.")
	}
	fmt.Println("")

	// Check if PDAO proposal checking duty is enabled
	fmt.Printf("%s=== PDAO Proposal Checking Duty ===%s\n", colorGreen, colorReset)
	// Make sure the user opted into this duty
	if response.VerifyEnabled {
		fmt.Println("The node has PDAO proposal checking duties enabled. It will periodically check for proposals to challenge.")
	} else {
		fmt.Printf("The node does not have PDAO proposal checking duties enabled (See %s to learn more about this duty).", challengeLink)
		fmt.Println("")
	}
	fmt.Println("")

	// Claimable Bonds Status:
	fmt.Printf("%s=== Pending, Active and Succeeded Proposals ===%s\n", colorGreen, colorReset)
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
	proposalStates := []string{"Pending", "Active (Phase 1)", "Active (Phase 2)", "Succeeded"}
	proposalStateInputs := []string{"pending", "phase1", "phase2", "succeeded"}

	// Print & return
	count := 0
	succeededExists := false
	for i, stateName := range proposalStates {
		proposals, ok := stateProposals[stateName]
		if !ok {
			continue
		}

		// Check filter
		if filterProposalState(proposalStateInputs[i], "") {
			continue
		}

		// Print message for Succeeded Proposals
		if stateName == "Succeeded" {
			succeededExists = true
			fmt.Printf("%sThe following proposal(s) have succeeded and are waiting to be executed. Use `rocketpool pdao proposals execute` to execute.%s\n\n", colorBlue, colorReset)
		}

		// Proposal state count
		fmt.Printf("%d %s proposal(s):\n", len(proposals), stateName)
		fmt.Println("")

		// Proposals
		for _, proposal := range proposals {
			if len(proposal.Message) > 200 {
				proposal.Message = proposal.Message[:200]
			}
			proposal.Message = strings.Sanitize(proposal.Message)
			fmt.Printf("%d: %s - Proposed by: %s\n", proposal.ID, proposal.Message, proposal.ProposerAddress)
		}

		count += len(proposals)

		fmt.Println()
	}
	if count == 0 {
		fmt.Println("There are no onchain proposals open for voting.")
	}
	if !succeededExists {
		fmt.Println("There are no proposals waiting to be executed.")
	}
	return nil

}
