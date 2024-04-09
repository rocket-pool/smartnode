package node

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/shared/types"
)

func getRewards(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get eligible intervals
	rewardsInfoResponse, err := rp.Api.Node.GetRewardsInfo()
	if err != nil {
		return fmt.Errorf("error getting rewards info: %w", err)
	}

	// Check for missing Merkle trees with rewards available
	missingIntervals := []types.IntervalInfo{}
	invalidIntervals := []types.IntervalInfo{}
	for _, intervalInfo := range rewardsInfoResponse.Data.InvalidIntervals {
		if !intervalInfo.TreeFileExists {
			fmt.Printf("You are missing the rewards tree file for interval %d.\n", intervalInfo.Index)
			missingIntervals = append(missingIntervals, intervalInfo)
		} else if !intervalInfo.MerkleRootValid {
			fmt.Printf("Your local copy of the rewards tree file for interval %d does not match the canonical one.\n", intervalInfo.Index)
			invalidIntervals = append(invalidIntervals, intervalInfo)
		}
	}

	// Download the Merkle trees for all unclaimed intervals that don't exist
	if len(missingIntervals) > 0 || len(invalidIntervals) > 0 {
		fmt.Println()
		fmt.Printf("%sNOTE: If you would like to regenerate these tree files manually, please answer `n` to the prompt below and run `rocketpool network generate-rewards-tree` before claiming your rewards.%s\n", terminal.ColorBlue, terminal.ColorReset)
		if !utils.Confirm("Would you like to download all missing rewards tree files now?") {
			fmt.Println("Cancelled.")
			return nil
		}

		// Download the files
		for _, missingInterval := range missingIntervals {
			fmt.Printf("Downloading interval %d file... ", missingInterval.Index)
			_, err := rp.Api.Network.DownloadRewardsFile(missingInterval.Index)
			if err != nil {
				return fmt.Errorf("error downloading rewards file for interval %d: %w", missingInterval.Index, err)
			}
			fmt.Println("done!")
		}
		for _, invalidInterval := range invalidIntervals {
			fmt.Printf("Downloading interval %d file... ", invalidInterval.Index)
			_, err := rp.Api.Network.DownloadRewardsFile(invalidInterval.Index)
			if err != nil {
				return fmt.Errorf("error downloading rewards file for interval %d: %w", invalidInterval.Index, err)
			}
			fmt.Println("done!")
		}
		fmt.Println()

		// Reload rewards now that the files are in place
		_, err = rp.Api.Node.GetRewardsInfo()
		if err != nil {
			return fmt.Errorf("error getting rewards info: %w", err)
		}
	}

	// Get node RPL rewards status
	rewards, err := rp.Api.Node.Rewards()
	if err != nil {
		return err
	}

	fmt.Printf("%sNOTE: Legacy rewards from pre-Redstone are temporarily not being included in the below figures. They will be added back in a future release. We apologize for the inconvenience!%s\n\n", terminal.ColorYellow, terminal.ColorReset)

	fmt.Println("=== ETH ===")
	fmt.Printf("You have earned %.4f ETH from the Beacon Chain (including your commissions) so far.\n", rewards.Data.BeaconRewards)
	fmt.Printf("You have claimed %.4f ETH from the Smoothing Pool.\n", rewards.Data.CumulativeEthRewards)
	fmt.Printf("You still have %.4f ETH in unclaimed Smoothing Pool rewards.\n", rewards.Data.UnclaimedEthRewards)

	nextRewardsTime := rewards.Data.LastCheckpoint.Add(rewards.Data.RewardsInterval)
	nextRewardsTimeString := utils.GetDateTimeString(uint64(nextRewardsTime.Unix()))
	timeToCheckpointString := time.Until(nextRewardsTime).Round(time.Second).String()

	// Assume 365 days in a year, 24 hours per day
	rplApr := rewards.Data.EstimatedRewards / rewards.Data.TotalRplStake / rewards.Data.RewardsInterval.Hours() * (24 * 365) * 100

	fmt.Println("\n=== RPL ===")
	fmt.Printf("The current rewards cycle started on %s.\n", utils.GetDateTimeString(uint64(rewards.Data.LastCheckpoint.Unix())))
	fmt.Printf("It will end on %s (%s from now).\n", nextRewardsTimeString, timeToCheckpointString)

	if rewards.Data.UnclaimedRplRewards > 0 {
		fmt.Printf("You currently have %f unclaimed RPL from staking rewards.\n", rewards.Data.UnclaimedRplRewards)
	}
	if rewards.Data.UnclaimedTrustedRplRewards > 0 {
		fmt.Printf("You currently have %f unclaimed RPL from Oracle DAO duties.\n", rewards.Data.UnclaimedTrustedRplRewards)
	}

	fmt.Println()
	fmt.Printf("Your estimated RPL staking rewards for this cycle: %f RPL (this may change based on network activity).\n", rewards.Data.EstimatedRewards)
	fmt.Printf("Based on your current total stake of %f RPL, this is approximately %.2f%% APR.\n", rewards.Data.TotalRplStake, rplApr)
	fmt.Printf("Your node has received %f RPL staking rewards in total.\n", rewards.Data.CumulativeRplRewards)

	if rewards.Data.Trusted {
		rplTrustedApr := rewards.Data.EstimatedTrustedRplRewards / rewards.Data.TrustedRplBond / rewards.Data.RewardsInterval.Hours() * (24 * 365) * 100

		fmt.Println()
		fmt.Printf("You will receive an estimated %f RPL in rewards for Oracle DAO duties (this may change based on network activity).\n", rewards.Data.EstimatedTrustedRplRewards)
		fmt.Printf("Based on your bond of %f RPL, this is approximately %.2f%% APR.\n", rewards.Data.TrustedRplBond, rplTrustedApr)
		fmt.Printf("Your node has received %f RPL Oracle DAO rewards in total.\n", rewards.Data.CumulativeTrustedRplRewards)
	}

	fmt.Println()
	fmt.Println("You may claim these rewards at any time. You no longer need to claim them within this interval.")

	// Return
	return nil
}
