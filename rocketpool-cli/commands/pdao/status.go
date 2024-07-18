package pdao

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/eth"
	utilsMath "github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

const (
	colorBlue  string = terminal.ColorBlue
	colorReset string = terminal.ColorReset
	colorGreen string = terminal.ColorGreen

	signallingAddressLink string = "https://docs.rocketpool.net/guides/houston/participate#setting-your-snapshot-signalling-address"
)

func getStatus(c *cli.Context) error {

	// Get RP client
	rp, err := client.NewClientFromCtx(c)
	if err != nil {
		return err
	}

	// Get PDAO status at the latest block
	response, err := rp.Api.PDao.GetStatus()
	if err != nil {
		return err
	}

	// Get Protocol DAO proposals
	allProposals, err := rp.Api.PDao.Proposals()
	if err != nil {
		return err
	}

	// Get protocol DAO proposals
	claimableBondsResponse, err := rp.Api.PDao.GetClaimableBonds()
	if err != nil {
		return fmt.Errorf("error checking for claimable bonds: %w", err)
	}
	claimableBonds := claimableBondsResponse.Data.ClaimableBonds

	// Signalling Status
	fmt.Printf("%s=== Signalling on Snapshot ===%s\n", colorGreen, colorReset)
	blankAddress := common.Address{}
	if response.Data.SignallingAddress == blankAddress {
		fmt.Println("The node does not currently have a snapshot signalling address set.")
		fmt.Printf("To learn more about snapshot signalling, please visit %s.\n", signallingAddressLink)
	} else {
		fmt.Printf("The node has a signalling address of %s%s%s which can represent it when voting on Rocket Pool Snapshot governance proposals.\n", colorBlue, response.Data.SignallingAddressFormatted, colorReset)
	}
	if response.Data.SnapshotResponse.Error != "" {
		fmt.Printf("Unable to fetch latest voting information from snapshot.org: %s\n", response.Data.SnapshotResponse.Error)
	} else {
		voteCount := response.Data.SnapshotResponse.VoteCount()
		if len(response.Data.SnapshotResponse.ActiveSnapshotProposals) == 0 {
			fmt.Println("Rocket Pool has no Snapshot governance proposals being voted on.")
		} else {
			fmt.Printf("Rocket Pool has %d Snapshot governance proposal(s) being voted on. You have voted on %d of those. See details using 'rocketpool network dao-proposals'.\n", len(response.Data.SnapshotResponse.ActiveSnapshotProposals), voteCount)
		}
	}
	fmt.Println()

	// Onchain Voting Status
	fmt.Printf("%s=== Onchain Voting ===%s\n", colorGreen, colorReset)
	if response.Data.IsVotingInitialized {
		fmt.Printf("The node %s%s%s has been initialized for onchain voting.\n", colorBlue, response.Data.AccountAddressFormatted, colorReset)
	} else {
		fmt.Printf("The node %s%s%s has NOT been initialized for onchain voting. You need to run `rocketpool pdao initialize-voting` to participate in onchain votes.\n", colorBlue, response.Data.AccountAddressFormatted, colorReset)
	}
	if response.Data.OnchainVotingDelegate == blankAddress {
		fmt.Println("The node doesn't have a delegate, which means it can vote directly on onchain proposals after it initializes voting.")
	} else if response.Data.OnchainVotingDelegate == response.Data.AccountAddress {
		fmt.Println("The node doesn't have a delegate, which means it can vote directly on onchain proposals. You can have another node represent you by running `rocketpool p svd <address>`.")
	} else {
		fmt.Printf("The node has a voting delegate of %s%s%s which can represent it when voting on Rocket Pool onchain governance proposals.\n", colorBlue, response.Data.OnchainVotingDelegateFormatted, colorReset)
	}
	fmt.Printf("The node's local voting power: %.10f\n", eth.WeiToEth(response.Data.VotingPower))
	if response.Data.IsNodeRegistered {
		fmt.Printf("Total voting power delegated to the node: %.10f\n", eth.WeiToEth(response.Data.TotalDelegatedVp))
	} else {
		fmt.Println("The node must register using 'rocketpool node register' to be eligible to receive delegated voting power")
	}
	fmt.Printf("Network total initialized voting power: %.10f\n", eth.WeiToEth(response.Data.SumVotingPower))
	fmt.Println()

	// Claimable Bonds Status:
	fmt.Printf("%s=== Claimable RPL Bonds ===%s\n", colorGreen, colorReset)
	if response.Data.IsRPLLockingAllowed {
		fmt.Println("The node is allowed to lock RPL to create governance proposals/challenges.")
		if response.Data.NodeRPLLocked.Cmp(big.NewInt(0)) != 0 {
			fmt.Printf("The node currently has %.6f RPL locked.\n", utilsMath.RoundDown(eth.WeiToEth(response.Data.NodeRPLLocked), 6))
		}
	} else {
		fmt.Println("The node is NOT allowed to lock RPL to create governance proposals/challenges. Use 'rocketpool node allow-rpl-locking, to allow RPL locking.")
	}
	if len(claimableBonds) == 0 {
		fmt.Println("The node does not have any unlockable bonds or claimable rewards.")
	} else {
		fmt.Println("The node has unlockable bonds or claimable rewards available. Use 'rocketpool pdao claim-bonds' to view and claim.")
	}
	fmt.Println()

	// Check if PDAO proposal checking duty is enabled
	fmt.Printf("%s=== PDAO Proposal Checking Duty ===%s\n", colorGreen, colorReset)
	// Make sure the user opted into this duty
	if response.Data.VerifyEnabled {
		fmt.Println("The node has PDAO proposal checking duties enabled. It will periodically check for proposals to challenge.")
	} else {
		fmt.Println("The node does not have PDAO proposal checking duties enabled (See https://docs.rocketpool.net/guides/houston/pdao#challenge-process to learn more about this duty).")
	}
	fmt.Println()

	// Claimable Bonds Status:
	fmt.Printf("%s=== Pending, Active and Succeeded Proposals ===%s\n", colorGreen, colorReset)
	// Get proposals by state
	stateProposals := map[string][]api.ProtocolDaoProposalDetails{}
	for _, proposal := range allProposals.Data.Proposals {
		stateName := types.ProtocolDaoProposalStates[proposal.State]
		if _, ok := stateProposals[stateName]; !ok {
			stateProposals[stateName] = []api.ProtocolDaoProposalDetails{}
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
			fmt.Printf("%sThe following proposal(s) have succeeded and are waiting to be executed. Use `rocketpool pdao proposals execute` to execute.%s\n", colorBlue, colorReset)
		}

		// Proposal state count
		fmt.Printf("%d %s proposal(s):\n", len(proposals), stateName)
		fmt.Println()

		// Proposals
		for _, proposal := range proposals {
			proposal.Message = utils.TruncateAndSanitize(proposal.Message, 200)
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
