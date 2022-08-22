package node

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

const (
	colorReset  string = "\033[0m"
	colorRed    string = "\033[31m"
	colorGreen  string = "\033[32m"
	colorYellow string = "\033[33m"
)

func getStatus(c *cli.Context) error {

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

	// Account address & balances
	fmt.Printf("%s=== Account and Balances ===%s\n", colorGreen, colorReset)
	fmt.Printf(
		"The node %s%s%s has a balance of %.6f ETH and %.6f RPL.\n",
		colorBlue,
		status.AccountAddress.Hex(),
		colorReset,
		math.RoundDown(eth.WeiToEth(status.AccountBalances.ETH), 6),
		math.RoundDown(eth.WeiToEth(status.AccountBalances.RPL), 6))
	if status.AccountBalances.FixedSupplyRPL.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("The node has a balance of %.6f old RPL which can be swapped for new RPL.\n", math.RoundDown(eth.WeiToEth(status.AccountBalances.FixedSupplyRPL), 6))
	}

	// Registered node details
	if status.Registered {

		// Node status
		fmt.Printf("The node is registered with Rocket Pool with a timezone location of %s.\n", status.TimezoneLocation)
		if status.Trusted {
			fmt.Println("The node is a member of the oracle DAO - it can vote on DAO proposals and perform watchtower duties.")
		}
		fmt.Println("")

		// Penalties
		fmt.Printf("%s=== Penalty Status ===%s\n", colorGreen, colorReset)
		if len(status.PenalizedMinipools) > 0 {
			strikeMinipools := []common.Address{}
			infractionMinipools := []common.Address{}
			for mp, count := range status.PenalizedMinipools {
				if count < 3 {
					strikeMinipools = append(strikeMinipools, mp)
				} else {
					infractionMinipools = append(infractionMinipools, mp)
				}
			}

			if len(strikeMinipools) > 0 {
				sort.Slice(strikeMinipools, func(i, j int) bool { // Sort them lexicographically
					return strikeMinipools[i].Hex() < strikeMinipools[j].Hex()
				})
				fmt.Printf("%sWARNING: The following minipools have been given strikes for cheating with an invalid fee recipient:\n", colorYellow)
				for _, mp := range strikeMinipools {
					fmt.Printf("\t%s: %d strikes\n", mp.Hex(), status.PenalizedMinipools[mp])
				}
				fmt.Println(colorReset)
				fmt.Println()
			}

			if len(infractionMinipools) > 0 {
				sort.Slice(infractionMinipools, func(i, j int) bool { // Sort them lexicographically
					return infractionMinipools[i].Hex() < infractionMinipools[j].Hex()
				})
				fmt.Printf("%sWARNING: The following minipools have been given infractions for cheating with an invalid fee recipient:\n", colorRed)
				for _, mp := range infractionMinipools {
					fmt.Printf("\t%s: %d infractions\n", mp.Hex(), status.PenalizedMinipools[mp]-2)
				}
				fmt.Println(colorReset)
				fmt.Println()
			}
		} else {
			fmt.Println("The node does not have any penalties for cheating with an invalid fee recipient.")
			fmt.Println()
		}

		// Voting status
		fmt.Printf("%s=== DAO Voting ===%s\n", colorGreen, colorReset)
		blankAddress := common.Address{}
		if status.VotingDelegate == blankAddress {
			fmt.Println("The node does not currently have a voting delegate set, and will not be able to vote on Rocket Pool governance proposals.")
		} else {
			fmt.Printf("The node has a voting delegate of %s%s%s which can represent it when voting on Rocket Pool governance proposals.\n", colorBlue, status.VotingDelegate.Hex(), colorReset)
		}

		switch len(status.ActiveSnapshotProposals) {
		case 0:
			fmt.Print("Rocket Pool has no active proposals on Snapshot.\n")
		case 1:
			fmt.Print("Rocket Pool has 1 active proposal on Snapshot:\n")
		default:
			fmt.Printf("Rocket Pool has %d active proposals on Snapshot:\n", len(status.ActiveSnapshotProposals))
		}

		for _, proposal := range status.ActiveSnapshotProposals {
			fmt.Printf("\nTitle: %s\n", proposal.Title)
			currentTimestamp := time.Now().Unix()
			if currentTimestamp < proposal.Start {
				fmt.Printf("Start: %s (in %s)\n", cliutils.GetDateTimeString(uint64(proposal.Start)), cliutils.GetDateTimeDiffString(uint64(proposal.Start), uint64(currentTimestamp)))
			} else {
				fmt.Printf("End: %s (in %s) \n", cliutils.GetDateTimeString(uint64(proposal.End)), cliutils.GetDateTimeDiffString(uint64(proposal.End), uint64(currentTimestamp)))
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
			}
		}
		fmt.Println("")

		// Withdrawal address & balances
		fmt.Printf("%s=== Withdrawal Address ===%s\n", colorGreen, colorReset)
		if !bytes.Equal(status.AccountAddress.Bytes(), status.WithdrawalAddress.Bytes()) {
			fmt.Printf(
				"The node's withdrawal address %s%s%s has a balance of %.6f ETH and %.6f RPL.\n",
				colorBlue,
				status.WithdrawalAddress.Hex(),
				colorReset,
				math.RoundDown(eth.WeiToEth(status.WithdrawalBalances.ETH), 6),
				math.RoundDown(eth.WeiToEth(status.WithdrawalBalances.RPL), 6))
		} else {
			fmt.Printf("%sThe node's withdrawal address has not been changed, so rewards and withdrawals will be sent to the node itself.\n", colorYellow)
			fmt.Printf("Consider changing this to a cold wallet address that you control using the `set-withdrawal-address` command.\n%s", colorReset)
		}
		fmt.Println("")
		if status.PendingWithdrawalAddress.Hex() != blankAddress.Hex() {
			fmt.Printf("%sThe node's withdrawal address has a pending change to %s which has not been confirmed yet.\n", colorYellow, status.PendingWithdrawalAddress.Hex())
			fmt.Printf("Please visit the Rocket Pool website with a web3-compatible wallet to complete this change.%s\n", colorReset)
			fmt.Println("")
		}

		// Fee distributor details
		if status.IsMergeUpdateDeployed {
			fmt.Printf("%s=== Fee Distributor and Smoothing Pool ===%s\n", colorGreen, colorReset)
			fmt.Printf("The node's fee distributor %s%s%s has a balance of %.6f ETH.\n", colorBlue, status.FeeDistributorAddress.Hex(), colorReset, math.RoundDown(eth.WeiToEth(status.FeeDistributorBalance), 6))
			if !status.IsFeeDistributorInitialized {
				fmt.Printf("%sThe fee distributor hasn't been initialized yet. When you are able, please initialize it with `rocketpool node initialize-fee-distributor`.%s\n\n", colorYellow, colorReset)
			}

			// Smoothing Pool
			if status.IsInSmoothingPool {
				fmt.Println("The node is opted into the Smoothing Pool.")
			} else {
				fmt.Println("The node is not opted into the Smoothing Pool.")
			}
			fmt.Println()
		}

		// RPL stake details
		fmt.Printf("%s=== RPL Stake and Minipools ===%s\n", colorGreen, colorReset)
		fmt.Printf(
			"The node has a total stake of %.6f RPL and an effective stake of %.6f RPL, allowing it to run %d minipool(s) in total.\n",
			math.RoundDown(eth.WeiToEth(status.RplStake), 6),
			math.RoundDown(eth.WeiToEth(status.EffectiveRplStake), 6),
			status.MinipoolLimit)
		if status.CollateralRatio > 0 {
			fmt.Printf(
				"This is currently a %.2f%% collateral ratio.\n",
				status.CollateralRatio*100,
			)
		}

		// Minipool details
		if status.MinipoolCounts.Total > 0 {

			// RPL stake
			fmt.Printf("The node must keep at least %.6f RPL staked to collateralize its minipools and claim RPL rewards.\n", math.RoundDown(eth.WeiToEth(status.MinimumRplStake), 6))
			fmt.Println("")

			// Minipools
			fmt.Printf("The node has a total of %d active minipool(s):\n", status.MinipoolCounts.Total-status.MinipoolCounts.Finalised)
			if status.MinipoolCounts.Initialized > 0 {
				fmt.Printf("- %d initialized\n", status.MinipoolCounts.Initialized)
			}
			if status.MinipoolCounts.Prelaunch > 0 {
				fmt.Printf("- %d at prelaunch\n", status.MinipoolCounts.Prelaunch)
			}
			if status.MinipoolCounts.Staking > 0 {
				fmt.Printf("- %d staking\n", status.MinipoolCounts.Staking)
			}
			if status.MinipoolCounts.Withdrawable > 0 {
				fmt.Printf("- %d withdrawable (after withdrawal delay)\n", status.MinipoolCounts.Withdrawable)
			}
			if status.MinipoolCounts.Dissolved > 0 {
				fmt.Printf("- %d dissolved\n", status.MinipoolCounts.Dissolved)
			}
			if status.MinipoolCounts.RefundAvailable > 0 {
				fmt.Printf("* %d minipool(s) have refunds available!\n", status.MinipoolCounts.RefundAvailable)
			}
			if status.MinipoolCounts.WithdrawalAvailable > 0 {
				fmt.Printf("* %d minipool(s) are ready for withdrawal once Beacon Chain withdrawals are enabled!\n", status.MinipoolCounts.WithdrawalAvailable)
			}
			if status.MinipoolCounts.CloseAvailable > 0 {
				fmt.Printf("* %d dissolved minipool(s) can be closed once Beacon Chain withdrawals are enabled!\n", status.MinipoolCounts.CloseAvailable)
			}
			if status.MinipoolCounts.Finalised > 0 {
				fmt.Printf("* %d minipool(s) are finalized and no longer active.\n", status.MinipoolCounts.Finalised)
			}

		} else {
			fmt.Println("The node does not have any minipools yet.")
		}

	} else {
		fmt.Println("The node is not registered with Rocket Pool.")
	}

	// Return
	return nil

}
