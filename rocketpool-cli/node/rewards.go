package node

import (
	"fmt"
	"math/big"
	"time"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func getRewards(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get eligible intervals
	rewardsInfoResponse, err := rp.GetRewardsInfo()
	if err != nil {
		return fmt.Errorf("error getting rewards info: %w", err)
	}

	if !rewardsInfoResponse.Registered {
		fmt.Printf("This node is not currently registered.\n")
		return nil
	}

	// Check for missing Merkle trees with rewards available
	missingIntervals := []rprewards.IntervalInfo{}
	invalidIntervals := []rprewards.IntervalInfo{}
	for _, intervalInfo := range rewardsInfoResponse.InvalidIntervals {
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
		fmt.Printf("%sNOTE: If you would like to regenerate these tree files manually, please answer `n` to the prompt below and run `rocketpool network generate-rewards-tree` before claiming your rewards.%s\n", colorBlue, colorReset)
		if !prompt.Confirm("Would you like to download all missing rewards tree files now?") {
			fmt.Println("Cancelled.")
			return nil
		}

		// Download the files
		for _, missingInterval := range missingIntervals {
			fmt.Printf("Downloading interval %d file... ", missingInterval.Index)
			_, err := rp.DownloadRewardsFile(missingInterval.Index)
			if err != nil {
				return fmt.Errorf("error downloading rewards file for interval %d: %w", missingInterval.Index, err)
			}
			fmt.Println("done!")
		}
		for _, invalidInterval := range invalidIntervals {
			fmt.Printf("Downloading interval %d file... ", invalidInterval.Index)
			_, err := rp.DownloadRewardsFile(invalidInterval.Index)
			if err != nil {
				return fmt.Errorf("error downloading rewards file for interval %d: %w", invalidInterval.Index, err)
			}
			fmt.Println("done!")
		}
		fmt.Println()

		// Reload rewards now that the files are in place
		rewardsInfoResponse, err = rp.GetRewardsInfo()
		if err != nil {
			return fmt.Errorf("error getting rewards info: %w", err)
		}
	}

	// Get node RPL rewards status
	rewards, err := rp.NodeRewards()
	if err != nil {
		return err
	}

	// Check if Saturn is already deployed
	saturnResp, err := rp.IsSaturnDeployed()
	if err != nil {
		return err
	}
	if saturnResp.IsSaturnDeployed {
		beaconBalances, err := rp.GetValidatorMapAndBalances()
		if err != nil {
			return err
		}
		megapoolUnskimmedRewards := new(big.Int).Sub(beaconBalances.NodeBond, beaconBalances.NodeShareOfCLBalance)
		megapoolUnskimmedRewardsFloat := eth.WeiToEth(megapoolUnskimmedRewards)
		// Add the megapool unskimmed beacon rewards
		rewards.BeaconRewards = rewards.BeaconRewards + megapoolUnskimmedRewardsFloat
	}

	fmt.Println("=== ETH ===")
	fmt.Printf("Your share of unskimmed Beacon Chain (CL) rewards is currently %.6f ETH.\n", rewards.BeaconRewards)
	fmt.Printf("You have claimed %.6f ETH from the Smoothing Pool.\n", rewards.CumulativeEthRewards)
	fmt.Printf("You still have %.6f ETH in unclaimed Smoothing Pool rewards.\n", rewards.UnclaimedEthRewards)

	nextRewardsTime := rewards.LastCheckpoint.Add(rewards.RewardsInterval)
	nextRewardsTimeString := cliutils.GetDateTimeString(uint64(nextRewardsTime.Unix()))
	timeToCheckpointString := time.Until(nextRewardsTime).Round(time.Second).String()

	// // Assume 365 days in a year, 24 hours per day
	rplApr := 0.0
	if rewards.TotalRplStake != 0 && rewards.RewardsInterval.Hours() != 0 {
		rplApr = rewards.EstimatedRewards / rewards.TotalRplStake / rewards.RewardsInterval.Hours() * (24 * 365) * 100
	}

	fmt.Println("\n=== RPL ===")
	fmt.Printf("The current rewards cycle started on %s.\n", cliutils.GetDateTimeString(uint64(rewards.LastCheckpoint.Unix())))
	fmt.Printf("It will end on %s (%s from now).\n", nextRewardsTimeString, timeToCheckpointString)

	if rewards.UnclaimedRplRewards > 0 {
		fmt.Printf("You currently have %f unclaimed RPL from staking rewards.\n", rewards.UnclaimedRplRewards)
	}
	if rewards.UnclaimedTrustedRplRewards > 0 {
		fmt.Printf("You currently have %f unclaimed RPL from Oracle DAO duties.\n", rewards.UnclaimedTrustedRplRewards)
	}

	fmt.Println()
	fmt.Printf("Your estimated RPL staking rewards for this cycle: %f RPL (this may change based on network activity).\n", rewards.EstimatedRewards)
	fmt.Printf("Based on your current total stake of %f RPL, this is approximately %.2f%% APR.\n", rewards.TotalRplStake, rplApr)
	fmt.Printf("Your node has received %f RPL staking rewards in total.\n", rewards.CumulativeRplRewards)

	if rewards.Trusted {
		rplTrustedApr := rewards.EstimatedTrustedRplRewards / rewards.TrustedRplBond / rewards.RewardsInterval.Hours() * (24 * 365) * 100

		fmt.Println()
		fmt.Printf("You will receive an estimated %f RPL in rewards for Oracle DAO duties (this may change based on network activity).\n", rewards.EstimatedTrustedRplRewards)
		fmt.Printf("Based on your bond of %f RPL, this is approximately %.2f%% APR.\n", rewards.TrustedRplBond, rplTrustedApr)
		fmt.Printf("Your node has received %f RPL Oracle DAO rewards in total.\n", rewards.CumulativeTrustedRplRewards)
	}

	fmt.Println()
	fmt.Println("You may claim these rewards at any time.")

	// Return
	return nil

}
