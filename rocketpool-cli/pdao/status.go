package pdao

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

const (
	colorBlue  string = "\033[36m"
	colorReset string = "\033[0m"
	colorGreen string = "\033[32m"
)

func getStatus(c *cli.Context) error {
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

	// Get PDAO status at the latest block
	response, err := rp.PDAOStatus()
	if err != nil {
		return err
	}

	status, err := rp.NodeStatus()
	if err != nil {
		return err
	}

	// Snapshot voting status
	fmt.Printf("%s=== Snapshot Voting ===%s\n", colorGreen, colorReset)
	blankAddress := common.Address{}
	if status.SnapshotVotingDelegate == blankAddress {
		fmt.Println("The node does not currently have a voting delegate set, which means it can only vote directly on Snapshot proposals (using a hardware wallet with the node mnemonic loaded).\nRun `rocketpool n sv <address>` to vote from a different wallet or have a delegate represent you. (See https://delegates.rocketpool.net for options)")
	} else {
		fmt.Printf("The node has a voting delegate of %s%s%s which can represent it when voting on Rocket Pool Snapshot governance proposals.\n", colorBlue, status.SnapshotVotingDelegateFormatted, colorReset)
	}

	if status.SnapshotResponse.Error != "" {
		fmt.Printf("Unable to fetch latest voting information from snapshot.org: %s\n", status.SnapshotResponse.Error)
	} else {
		voteCount := 0
		for _, activeProposal := range status.SnapshotResponse.ActiveSnapshotProposals {
			for _, votedProposal := range status.SnapshotResponse.ProposalVotes {
				if votedProposal.Proposal.Id == activeProposal.Id {
					voteCount++
					break
				}
			}
		}
		if len(status.SnapshotResponse.ActiveSnapshotProposals) == 0 {
			fmt.Print("Rocket Pool has no Snapshot governance proposals being voted on.\n")
		} else {
			fmt.Printf("Rocket Pool has %d Snapshot governance proposal(s) being voted on. You have voted on %d of those. See details using 'rocketpool network dao-proposals'.\n", len(status.SnapshotResponse.ActiveSnapshotProposals), voteCount)
		}
		fmt.Println("")
	}

	// Onchain Voting Status
	if status.IsHoustonDeployed {
		fmt.Printf("%s== Onchain Voting ==%s\n", colorGreen, colorReset)
		if status.IsVotingInitialized {
			fmt.Println("The node has been initialized for onchain voting.")

		} else {
			fmt.Println("The node has NOT been initialized for onchain voting. You need to run `rocketpool pdao initialize-voting` to participate in onchain votes.")
		}

		blankAddress := common.Address{}

		if status.OnchainVotingDelegate == blankAddress {
			fmt.Println("The node doesn't have a delegate, which means it can vote directly on onchain proposals after it initializes voting.")
		} else if status.OnchainVotingDelegate == status.AccountAddress {
			fmt.Println("The node doesn't have a delegate, which means it can vote directly on onchain proposals. You can have another node represent you by running `rocketpool p svd <address>`.")
		} else {
			fmt.Printf("The node has a voting delegate of %s%s%s which can represent it when voting on Rocket Pool onchain governance proposals.\n", colorBlue, status.OnchainVotingDelegateFormatted, colorReset)
		}
		if status.IsRPLLockingAllowed {
			fmt.Print("The node is allowed to lock RPL to create governance proposals/challenges.\n")
			if status.NodeRPLLocked.Cmp(big.NewInt(0)) != 0 {
				fmt.Printf("There are currently %.6f RPL locked.\n",
					math.RoundDown(eth.WeiToEth(status.NodeRPLLocked), 6))
			}

		} else {
			fmt.Print("The node is NOT allowed to lock RPL to create governance proposals/challenges.\n")
		}
		fmt.Printf("Your current voting power: %.10f\n", eth.WeiToEth(response.VotingPower))

	}

	return nil

}
