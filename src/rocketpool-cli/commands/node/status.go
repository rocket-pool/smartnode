package node

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/wallet"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/addons/rescue_node"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
)

const (
	smoothingPoolLink string = "https://docs.rocketpool.net/guides/redstone/whats-new.html#smoothing-pool"
	maxAlertItems     int    = 3
)

func getStatus(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get the config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Get wallet status
	statusResponse, err := rp.Api.Wallet.Status()
	if err != nil {
		return err
	}
	walletStatus := statusResponse.Data.WalletStatus

	// Rescue Node Plugin - ensure that we print the rescue node stuff even
	// when the eth1 node is syncing by deferring it here.
	//
	// Since we collected all the data we need for this message, we can safely
	// defer it and let it execute even if we fail further down, e.g. because
	// the EC is still syncing.
	if wallet.IsWalletReady(walletStatus) {
		defer func() {
			if cfg.Addons.RescueNode.Enabled.Value {
				fmt.Println()

				rn := rescue_node.NewRescueNode(cfg.Addons.RescueNode)
				rn.PrintStatusText(walletStatus.Address.NodeAddress)
			}
		}()
	}

	// Print what network we're on
	err = utils.PrintNetwork(cfg.Network.Value, isNew)
	if err != nil {
		return err
	}

	// rp.NodeStatus() will fail with an error, but we can short-circuit it here.
	if !walletStatus.Wallet.IsOnDisk {
		return errors.New("The node wallet is not initialized.")
	}

	// Get node status
	status, err := rp.Api.Node.Status()
	if err != nil {
		return err
	}

	// Account address & balances
	fmt.Printf("%s=== Account and Balances ===%s\n", terminal.ColorGreen, terminal.ColorReset)
	fmt.Printf(
		"The node %s%s%s has a balance of %.6f ETH and %.6f RPL.\n",
		terminal.ColorBlue,
		status.Data.AccountAddressFormatted,
		terminal.ColorReset,
		math.RoundDown(eth.WeiToEth(status.Data.NodeBalances.Eth), 6),
		math.RoundDown(eth.WeiToEth(status.Data.NodeBalances.Rpl), 6))
	if status.Data.NodeBalances.Fsrpl.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("The node has a balance of %.6f old RPL which can be swapped for new RPL.\n", math.RoundDown(eth.WeiToEth(status.Data.NodeBalances.Fsrpl), 6))
	}
	fmt.Printf(
		"The node has %.6f ETH in its credit balance and %.6f ETH staked on its behalf. %.6f can be used to make new minipools.\n",
		math.RoundDown(eth.WeiToEth(status.Data.CreditBalance), 6),
		math.RoundDown(eth.WeiToEth(status.Data.EthOnBehalfBalance), 6),
		math.RoundDown(eth.WeiToEth(status.Data.UsableCreditAndEthOnBehalfBalance), 6),
	)

	// Registered node details
	if status.Data.Registered {
		// Node status
		fmt.Printf("The node is registered with Rocket Pool with a timezone location of %s.\n", status.Data.TimezoneLocation)
		if status.Data.Trusted {
			fmt.Println("The node is a member of the oracle DAO - it can vote on DAO proposals and perform watchtower duties.")
		}
		fmt.Println("")

		// Penalties
		fmt.Printf("%s=== Penalty Status ===%s\n", terminal.ColorGreen, terminal.ColorReset)
		if len(status.Data.PenalizedMinipools) > 0 {
			strikeMinipools := []common.Address{}
			infractionMinipools := []common.Address{}
			for mp, count := range status.Data.PenalizedMinipools {
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
				fmt.Printf("%sWARNING: The following minipools have been given strikes for cheating with an invalid fee recipient:\n", terminal.ColorYellow)
				for _, mp := range strikeMinipools {
					fmt.Printf("\t%s: %d strikes\n", mp.Hex(), status.Data.PenalizedMinipools[mp])
				}
				fmt.Println(terminal.ColorReset)
				fmt.Println()
			}

			if len(infractionMinipools) > 0 {
				sort.Slice(infractionMinipools, func(i, j int) bool { // Sort them lexicographically
					return infractionMinipools[i].Hex() < infractionMinipools[j].Hex()
				})
				fmt.Printf("%sWARNING: The following minipools have been given infractions for cheating with an invalid fee recipient:\n", terminal.ColorRed)
				for _, mp := range infractionMinipools {
					fmt.Printf("\t%s: %d infractions\n", mp.Hex(), status.Data.PenalizedMinipools[mp]-2)
				}
				fmt.Println(terminal.ColorReset)
				fmt.Println()
			}
		} else {
			fmt.Println("The node does not have any penalties for cheating with an invalid fee recipient.")
			fmt.Println()
		}

		// Snapshot voting status
		fmt.Printf("%s=== DAO Voting ===%s\n", terminal.ColorGreen, terminal.ColorReset)
		blankAddress := common.Address{}
		if status.Data.SnapshotVotingDelegate == blankAddress {
			fmt.Println("The node does not currently have a voting delegate set, and will not be able to vote on Rocket Pool Snapshot governance proposals.")
		} else {
			fmt.Printf("The node has a voting delegate of %s%s%s which can represent it when voting on Rocket Pool Snapshot governance proposals.\n", terminal.ColorBlue, status.Data.SnapshotVotingDelegateFormatted, terminal.ColorReset)
		}

		if status.Data.SnapshotResponse.Error != "" {
			fmt.Printf("Unable to fetch latest voting information from snapshot.org: %s\n", status.Data.SnapshotResponse.Error)
		} else {
			voteCount := 0
			for _, activeProposal := range status.Data.SnapshotResponse.ActiveSnapshotProposals {
				if len(activeProposal.DelegateVotes) > 0 || len(activeProposal.UserVotes) > 0 {
					voteCount++
					break
				}
			}
			if len(status.Data.SnapshotResponse.ActiveSnapshotProposals) == 0 {
				fmt.Print("Rocket Pool has no Snapshot governance proposals being voted on.\n")
			} else {
				fmt.Printf("Rocket Pool has %d Snapshot governance proposal(s) being voted on. You have voted on %d of those. See details using 'rocketpool network dao-proposals'.\n", len(status.Data.SnapshotResponse.ActiveSnapshotProposals), voteCount)
			}
			fmt.Println("")
		}

		// Onchain voting status
		fmt.Printf("%s=== Onchain Voting ===%s\n", terminal.ColorGreen, terminal.ColorReset)
		if status.Data.IsVotingInitialized {
			fmt.Println("The node has been initialized for onchain voting.")
		} else {
			fmt.Println("The node has NOT been initialized for onchain voting. You need to run `rocketpool pdao initialize-voting` to participate in onchain votes.")
		}

		if status.Data.OnchainVotingDelegate == status.Data.AccountAddress {
			fmt.Println("The node doesn't have a delegate, which means it can vote directly on onchain proposals.")
		} else {
			fmt.Printf("The node has a voting delegate of %s%s%s which can represent it when voting on Rocket Pool onchain governance proposals.\n", terminal.ColorBlue, status.Data.OnchainVotingDelegateFormatted, terminal.ColorReset)
		}
		fmt.Println("")

		// Primary withdrawal address & balances
		fmt.Printf("%s=== Primary Withdrawal Address ===%s\n", terminal.ColorGreen, terminal.ColorReset)
		if !bytes.Equal(status.Data.AccountAddress.Bytes(), status.Data.PrimaryWithdrawalAddress.Bytes()) {
			fmt.Printf(
				"The node's primary withdrawal address %s%s%s has a balance of %.6f ETH and %.6f RPL.\n",
				terminal.ColorBlue,
				status.Data.PrimaryWithdrawalAddressFormatted,
				terminal.ColorReset,
				math.RoundDown(eth.WeiToEth(status.Data.PrimaryWithdrawalBalances.Eth), 6),
				math.RoundDown(eth.WeiToEth(status.Data.PrimaryWithdrawalBalances.Rpl), 6))
		} else {
			fmt.Printf("%sThe node's primary withdrawal address has not been changed, so ETH rewards and minipool withdrawals will be sent to the node itself.\n", terminal.ColorYellow)
			fmt.Printf("Consider changing this to a cold wallet address that you control using the `set-withdrawal-address` command.\n%s", terminal.ColorReset)
		}
		fmt.Println("")
		if status.Data.PendingPrimaryWithdrawalAddress.Hex() != blankAddress.Hex() {
			fmt.Printf("%sThe node's primary withdrawal address has a pending change to %s which has not been confirmed yet.\n", terminal.ColorYellow, status.Data.PendingPrimaryWithdrawalAddressFormatted)
			fmt.Printf("Please visit the Rocket Pool website with a web3-compatible wallet to complete this change.%s\n", terminal.ColorReset)
			fmt.Println("")
		}

		// RPL withdrawal address & balances
		fmt.Printf("%s=== RPL Withdrawal Address ===%s\n", terminal.ColorGreen, terminal.ColorReset)
		if !status.Data.IsRplWithdrawalAddressSet {
			fmt.Printf("The node's RPL withdrawal address has not been set. All RPL rewards will be sent to the primary withdrawal address.\n")
		} else if bytes.Equal(status.Data.AccountAddress.Bytes(), status.Data.RplWithdrawalAddress.Bytes()) {
			fmt.Printf("The node's RPL withdrawal address has been explicitly set to the node address itself (%s%s%s).\n", terminal.ColorBlue, status.Data.RplWithdrawalAddressFormatted, terminal.ColorReset)
		} else if bytes.Equal(status.Data.PrimaryWithdrawalAddress.Bytes(), status.Data.RplWithdrawalAddress.Bytes()) {
			fmt.Printf("The node's RPL withdrawal address has been explicitly set to the primary withdrawal address (%s%s%s).\n", terminal.ColorBlue, status.Data.RplWithdrawalAddressFormatted, terminal.ColorReset)
		} else {
			fmt.Printf(
				"The node's RPL withdrawal address %s%s%s has a balance of %.6f ETH and %.6f RPL.\n",
				terminal.ColorBlue,
				status.Data.RplWithdrawalAddressFormatted,
				terminal.ColorReset,
				math.RoundDown(eth.WeiToEth(status.Data.RplWithdrawalBalances.Eth), 6),
				math.RoundDown(eth.WeiToEth(status.Data.RplWithdrawalBalances.Rpl), 6))
		}
		if status.Data.IsRplLockingAllowed {
			fmt.Print("The node is allowed to lock RPL to create governance proposals/challenges.\n")
			if status.Data.RplLocked.Cmp(big.NewInt(0)) != 0 {
				fmt.Printf("There is currently %.6f RPL locked.\n", math.RoundDown(eth.WeiToEth(status.Data.RplLocked), 6))
			}
		} else {
			fmt.Print("The node is NOT allowed to lock RPL to create governance proposals/challenges.\n")
		}
		fmt.Println("")
		if status.Data.PendingRplWithdrawalAddress.Hex() != blankAddress.Hex() {
			fmt.Printf("%sThe node's RPL withdrawal address has a pending change to %s which has not been confirmed yet.\n", terminal.ColorYellow, status.Data.PendingRplWithdrawalAddressFormatted)
			fmt.Printf("Please visit the Rocket Pool website with a web3-compatible wallet to complete this change.%s\n", terminal.ColorReset)
			fmt.Println("")
		}

		// Fee distributor details
		fmt.Printf("%s=== Fee Distributor and Smoothing Pool ===%s\n", terminal.ColorGreen, terminal.ColorReset)
		if status.Data.FeeRecipientInfo.IsInSmoothingPool {
			fmt.Printf(
				"The node is currently opted into the Smoothing Pool (%s%s%s).\n",
				terminal.ColorBlue,
				status.Data.FeeRecipientInfo.SmoothingPoolAddress.Hex(),
				terminal.ColorReset)
			if cfg.IsNativeMode {
				fmt.Printf("%sNOTE: You are in Native Mode; you MUST ensure that your Validator Client is using this address as its fee recipient!%s\n", terminal.ColorYellow, terminal.ColorReset)
			}
		} else if status.Data.FeeRecipientInfo.IsInOptOutCooldown {
			fmt.Printf(
				"The node is currently opting out of the Smoothing Pool, but cannot safely change its fee recipient yet.\nIt must remain the Smoothing Pool's address (%s%s%s) until the opt-out process is complete.\nIt can safely be changed once Epoch %d is finalized on the Beacon Chain.\n",
				terminal.ColorBlue,
				status.Data.FeeRecipientInfo.SmoothingPoolAddress.Hex(),
				terminal.ColorReset,
				status.Data.FeeRecipientInfo.OptOutEpoch)
			if cfg.IsNativeMode {
				fmt.Printf("%sNOTE: You are in Native Mode; you MUST ensure that your Validator Client is using this address as its fee recipient!%s\n", terminal.ColorYellow, terminal.ColorReset)
			}
		} else {
			fmt.Printf("The node is not opted into the Smoothing Pool.\nTo learn more about the Smoothing Pool, please visit %s.\n", smoothingPoolLink)
		}

		fmt.Printf("The node's fee distributor %s%s%s has a balance of %.6f ETH.\n", terminal.ColorBlue, status.Data.FeeRecipientInfo.FeeDistributorAddress.Hex(), terminal.ColorReset, math.RoundDown(eth.WeiToEth(status.Data.FeeDistributorBalance), 6))
		if cfg.IsNativeMode && !status.Data.FeeRecipientInfo.IsInSmoothingPool && !status.Data.FeeRecipientInfo.IsInOptOutCooldown {
			fmt.Printf("%sNOTE: You are in Native Mode; you MUST ensure that your Validator Client is using this address as its fee recipient!%s\n", terminal.ColorYellow, terminal.ColorReset)
		}
		if !status.Data.IsFeeDistributorInitialized {
			fmt.Printf("\n%sThe fee distributor hasn't been initialized yet. When you are able, please initialize it with `rocketpool node initialize-fee-distributor`.%s\n", terminal.ColorYellow, terminal.ColorReset)
		}

		fmt.Println()

		// RPL stake details
		fmt.Printf("%s=== RPL Stake ===%s\n", terminal.ColorGreen, terminal.ColorReset)
		fmt.Println("NOTE: The following figures take *any pending bond reductions* into account.")
		fmt.Println()
		fmt.Printf(
			"The node has a total stake of %.6f RPL and an effective stake of %.6f RPL.\n",
			math.RoundDown(eth.WeiToEth(status.Data.RplStake), 6),
			math.RoundDown(eth.WeiToEth(status.Data.EffectiveRplStake), 6))
		if status.Data.BorrowedCollateralRatio > 0 {
			rplTooLow := (status.Data.RplStake.Cmp(status.Data.MinimumRplStake) < 0)
			if rplTooLow {
				fmt.Printf(
					"This is currently %s%.2f%% of its borrowed ETH%s and %.2f%% of its bonded ETH.\n",
					terminal.ColorRed, status.Data.BorrowedCollateralRatio*100, terminal.ColorReset, status.Data.BondedCollateralRatio*100)
			} else {
				fmt.Printf(
					"This is currently %.2f%% of its borrowed ETH and %.2f%% of its bonded ETH.\n",
					status.Data.BorrowedCollateralRatio*100, status.Data.BondedCollateralRatio*100)
			}
			fmt.Printf(
				"It must keep at least %.6f RPL staked to claim RPL rewards (10%% of borrowed ETH).\n", math.RoundDown(eth.WeiToEth(status.Data.MinimumRplStake), 6))
			fmt.Printf(
				"RPIP-30 is in effect and the node will gradually earn rewards in amounts above the previous limit of %.6f RPL (150%% of bonded ETH). Read more at https://github.com/rocket-pool/RPIPs/blob/main/RPIPs/RPIP-30.md\n", math.RoundDown(eth.WeiToEth(status.Data.MaximumRplStake), 6))
			if rplTooLow {
				fmt.Printf("%sWARNING: you are currently undercollateralized. You must stake at least %.6f more RPL in order to claim RPL rewards.%s\n", terminal.ColorRed, math.RoundUp(eth.WeiToEth(big.NewInt(0).Sub(status.Data.MinimumRplStake, status.Data.RplStake)), 6), terminal.ColorReset)
			}
		}
		fmt.Println()

		remainingAmount := big.NewInt(0).Sub(status.Data.EthMatchedLimit, status.Data.EthMatched)
		remainingAmount.Sub(remainingAmount, status.Data.PendingMatchAmount)
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
		fmt.Printf("%s=== Minipools ===%s\n", terminal.ColorGreen, terminal.ColorReset)
		if status.Data.MinipoolCounts.Total > 0 {

			// Minipools
			fmt.Printf("The node has a total of %d active minipool(s):\n", status.Data.MinipoolCounts.Total-status.Data.MinipoolCounts.Finalised)
			if status.Data.MinipoolCounts.Initialized > 0 {
				fmt.Printf("- %d initialized\n", status.Data.MinipoolCounts.Initialized)
			}
			if status.Data.MinipoolCounts.Prelaunch > 0 {
				fmt.Printf("- %d at prelaunch\n", status.Data.MinipoolCounts.Prelaunch)
			}
			if status.Data.MinipoolCounts.Staking > 0 {
				fmt.Printf("- %d staking\n", status.Data.MinipoolCounts.Staking)
			}
			if status.Data.MinipoolCounts.Withdrawable > 0 {
				fmt.Printf("- %d withdrawable (after withdrawal delay)\n", status.Data.MinipoolCounts.Withdrawable)
			}
			if status.Data.MinipoolCounts.Dissolved > 0 {
				fmt.Printf("- %d dissolved\n", status.Data.MinipoolCounts.Dissolved)
			}
			if status.Data.MinipoolCounts.RefundAvailable > 0 {
				fmt.Printf("* %d minipool(s) have refunds available!\n", status.Data.MinipoolCounts.RefundAvailable)
			}
			if status.Data.MinipoolCounts.Finalised > 0 {
				fmt.Printf("* %d minipool(s) are finalized and no longer active.\n", status.Data.MinipoolCounts.Finalised)
			}

		} else {
			fmt.Println("The node does not have any minipools yet.")
		}

	} else {
		fmt.Println("The node is not registered with Rocket Pool.")
	}

	// Alerts
	alerts := status.Data.Alerts
	if cfg.Metrics.EnableMetrics.Value && len(alerts) > 0 {
		// only print alerts if enabled; to avoid misleading the user to thinking everything is fine (since we really don't know).
		fmt.Printf("\n%s=== Alerts ===%s\n", terminal.ColorGreen, terminal.ColorReset)
		for i, alert := range alerts {
			fmt.Println(alert.ColorString())
			if i == maxAlertItems-1 {
				break
			}
		}
		if len(alerts) > maxAlertItems {
			fmt.Printf("... and %d more.\n", len(alerts)-maxAlertItems)
		}
	}

	if status.Data.Warning != "" {
		fmt.Printf("\n%sWARNING: %s%s\n", terminal.ColorRed, status.Data.Warning, terminal.ColorReset)
	}

	// Return
	return nil
}
