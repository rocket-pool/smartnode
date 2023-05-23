package node

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

const (
	colorReset        string = "\033[0m"
	colorRed          string = "\033[31m"
	colorGreen        string = "\033[32m"
	colorYellow       string = "\033[33m"
	smoothingPoolLink string = "https://docs.rocketpool.net/guides/redstone/whats-new.html#smoothing-pool"
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

	// Get the config
	cfg, _, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading configuration: %w", err)
	}

	// Account address & balances
	fmt.Printf("%s=== Account and Balances ===%s\n", colorGreen, colorReset)
	fmt.Printf(
		"The node %s%s%s has a balance of %.6f ETH and %.6f RPL.\n",
		colorBlue,
		status.AccountAddressFormatted,
		colorReset,
		math.RoundDown(eth.WeiToEth(status.AccountBalances.ETH), 6),
		math.RoundDown(eth.WeiToEth(status.AccountBalances.RPL), 6))
	if status.AccountBalances.FixedSupplyRPL.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("The node has a balance of %.6f old RPL which can be swapped for new RPL.\n", math.RoundDown(eth.WeiToEth(status.AccountBalances.FixedSupplyRPL), 6))
	}
	fmt.Printf(
		"The node has %.6f ETH in its credit balance, which can be used to make new minipools.\n",
		math.RoundDown(eth.WeiToEth(&status.CreditBalance), 6),
	)

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
			fmt.Printf("The node has a voting delegate of %s%s%s which can represent it when voting on Rocket Pool governance proposals.\n", colorBlue, status.VotingDelegateFormatted, colorReset)
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
				fmt.Print("Rocket Pool has no governance proposals being voted on.\n")
			} else {
				fmt.Printf("Rocket Pool has %d governance proposal(s) being voted on. You have voted on %d of those. See details using 'rocketpool network dao-proposals'.\n", len(status.SnapshotResponse.ActiveSnapshotProposals), voteCount)
			}
			fmt.Println("")
		}

		// Withdrawal address & balances
		fmt.Printf("%s=== Withdrawal Address ===%s\n", colorGreen, colorReset)
		if !bytes.Equal(status.AccountAddress.Bytes(), status.WithdrawalAddress.Bytes()) {
			fmt.Printf(
				"The node's withdrawal address %s%s%s has a balance of %.6f ETH and %.6f RPL.\n",
				colorBlue,
				status.WithdrawalAddressFormatted,
				colorReset,
				math.RoundDown(eth.WeiToEth(status.WithdrawalBalances.ETH), 6),
				math.RoundDown(eth.WeiToEth(status.WithdrawalBalances.RPL), 6))
		} else {
			fmt.Printf("%sThe node's withdrawal address has not been changed, so rewards and withdrawals will be sent to the node itself.\n", colorYellow)
			fmt.Printf("Consider changing this to a cold wallet address that you control using the `set-withdrawal-address` command.\n%s", colorReset)
		}
		fmt.Println("")
		if status.PendingWithdrawalAddress.Hex() != blankAddress.Hex() {
			fmt.Printf("%sThe node's withdrawal address has a pending change to %s which has not been confirmed yet.\n", colorYellow, status.PendingWithdrawalAddressFormatted)
			fmt.Printf("Please visit the Rocket Pool website with a web3-compatible wallet to complete this change.%s\n", colorReset)
			fmt.Println("")
		}

		// Fee distributor details
		fmt.Printf("%s=== Fee Distributor and Smoothing Pool ===%s\n", colorGreen, colorReset)
		if status.FeeRecipientInfo.IsInSmoothingPool {
			fmt.Printf(
				"The node is currently opted into the Smoothing Pool (%s%s%s).\n",
				colorBlue,
				status.FeeRecipientInfo.SmoothingPoolAddress.Hex(),
				colorReset)
			if cfg.IsNativeMode {
				fmt.Printf("%sNOTE: You are in Native Mode; you MUST ensure that your Validator Client is using this address as its fee recipient!%s\n", colorYellow, colorReset)
			}
		} else if status.FeeRecipientInfo.IsInOptOutCooldown {
			fmt.Printf(
				"The node is currently opting out of the Smoothing Pool, but cannot safely change its fee recipient yet.\nIt must remain the Smoothing Pool's address (%s%s%s) until the opt-out process is complete.\nIt can safely be changed once Epoch %d is finalized on the Beacon Chain.\n",
				colorBlue,
				status.FeeRecipientInfo.SmoothingPoolAddress.Hex(),
				colorReset,
				status.FeeRecipientInfo.OptOutEpoch)
			if cfg.IsNativeMode {
				fmt.Printf("%sNOTE: You are in Native Mode; you MUST ensure that your Validator Client is using this address as its fee recipient!%s\n", colorYellow, colorReset)
			}
		} else {
			fmt.Printf("The node is not opted into the Smoothing Pool.\nTo learn more about the Smoothing Pool, please visit %s.\n", smoothingPoolLink)
		}

		fmt.Printf("The node's fee distributor %s%s%s has a balance of %.6f ETH.\n", colorBlue, status.FeeRecipientInfo.FeeDistributorAddress.Hex(), colorReset, math.RoundDown(eth.WeiToEth(&status.FeeDistributorBalance), 6))
		if cfg.IsNativeMode && !status.FeeRecipientInfo.IsInSmoothingPool && !status.FeeRecipientInfo.IsInOptOutCooldown {
			fmt.Printf("%sNOTE: You are in Native Mode; you MUST ensure that your Validator Client is using this address as its fee recipient!%s\n", colorYellow, colorReset)
		}
		if !status.IsFeeDistributorInitialized {
			fmt.Printf("\n%sThe fee distributor hasn't been initialized yet. When you are able, please initialize it with `rocketpool node initialize-fee-distributor`.%s\n", colorYellow, colorReset)
		}

		fmt.Println()

		// RPL stake details
		fmt.Printf("%s=== RPL Stake ===%s\n", colorGreen, colorReset)
		fmt.Println("NOTE: The following figures take *any pending bond reductions* into account.\n")
		fmt.Printf(
			"The node has a total stake of %.6f RPL and an effective stake of %.6f RPL.\n",
			math.RoundDown(eth.WeiToEth(&status.RplStake), 6),
			math.RoundDown(eth.WeiToEth(&status.EffectiveRplStake), 6))
		if status.BorrowedCollateralRatio > 0 {
			rplTooLow := (status.RplStake.Cmp(&status.MinimumRplStake) < 0)
			if rplTooLow {
				fmt.Printf(
					"This is currently %s%.2f%% of its borrowed ETH%s and %.2f%% of its bonded ETH.\n",
					colorRed, status.BorrowedCollateralRatio*100, colorReset, status.BondedCollateralRatio*100)
			} else {
				fmt.Printf(
					"This is currently %.2f%% of its borrowed ETH and %.2f%% of its bonded ETH.\n",
					status.BorrowedCollateralRatio*100, status.BondedCollateralRatio*100)
			}
			fmt.Printf(
				"It must keep at least %.6f RPL staked to claim RPL rewards (10%% of borrowed ETH).\n", math.RoundDown(eth.WeiToEth(&status.MinimumRplStake), 6))
			fmt.Printf(
				"It can earn rewards on up to %.6f RPL (150%% of bonded ETH).\n", math.RoundDown(eth.WeiToEth(&status.MaximumRplStake), 6))
			if rplTooLow {
				fmt.Printf("%sWARNING: you are currently undercollateralized. You must stake at least %.6f more RPL in order to claim RPL rewards.%s\n", colorRed, math.RoundUp(eth.WeiToEth(big.NewInt(0).Sub(&status.MinimumRplStake, &status.RplStake)), 6), colorReset)
			}
		}
		fmt.Println()

		if status.PendingEffectiveRplStake.Cmp(&status.EffectiveRplStake) != 0 {
			fmt.Printf("Of this stake, %.6f RPL is eligible for RPL staking rewards.\n", math.RoundDown(eth.WeiToEth(&status.PendingEffectiveRplStake), 6))
			fmt.Println("Eligibility is determined by the number of minipools you have in the *active* state on the Beacon Chain:\n- Validators in the Beacon Chain queue that have not been activated yet are not eligible.\n- Validators that have been exited are not eligible.")
			fmt.Println()
		}

		remainingAmount := big.NewInt(0).Sub(&status.EthMatchedLimit, &status.EthMatched)
		remainingAmount.Sub(remainingAmount, &status.PendingMatchAmount)
		remainingAmountEth := int(eth.WeiToEth(remainingAmount))
		remainingFor8EB := remainingAmountEth / 24
		if remainingFor8EB < 0 {
			remainingFor8EB = 0
		}
		remainingFor16EB := remainingAmountEth / 16
		if remainingFor16EB < 0 {
			remainingFor16EB = 0
		}
		fmt.Printf("The node has enough RPL staked to make %d more 8-ETH minipools (or %d more 16-ETH minipools).\n\n", remainingFor8EB, remainingFor16EB)

		// Minipool details
		fmt.Printf("%s=== Minipools ===%s\n", colorGreen, colorReset)
		if status.MinipoolCounts.Total > 0 {

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
				fmt.Printf("* %d minipool(s) are ready for withdrawal!\n", status.MinipoolCounts.WithdrawalAvailable)
			}
			if status.MinipoolCounts.CloseAvailable > 0 {
				fmt.Printf("* %d dissolved minipool(s) can be closed and your deposit (minus the prelaunch amount) refunded!\n", status.MinipoolCounts.CloseAvailable)
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
