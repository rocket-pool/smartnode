package node

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

const (
	colorBlue string = "\033[36m"
)

func nodeClaimRewards(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Provide a notice
	fmt.Printf("%sWelcome to the new rewards system!\nYou no longer need to claim rewards at each interval - you can simply let them accumulate and claim them whenever you want.\nHere you can see which intervals you haven't claimed yet, and how many rewards you earned during each one.%s\n\n", colorBlue, colorReset)

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
		if !cliutils.Confirm("Would you like to download all missing rewards tree files now?") {
			fmt.Println("Cancelled.")
			return nil
		}

		// Load the config file
		cfg, _, err := rp.LoadConfig()
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		// Download the files
		for _, missingInterval := range missingIntervals {
			fmt.Printf("Downloading interval %d file... ", missingInterval.Index)
			err := rprewards.DownloadRewardsFile(cfg, missingInterval.Index, missingInterval.CID, missingInterval.MerkleRoot, false)
			if err != nil {
				fmt.Println()
				return err
			}
			fmt.Println("done!")
		}
		for _, invalidInterval := range invalidIntervals {
			fmt.Printf("Downloading interval %d file... ", invalidInterval.Index)
			err := rprewards.DownloadRewardsFile(cfg, invalidInterval.Index, invalidInterval.CID, invalidInterval.MerkleRoot, false)
			if err != nil {
				fmt.Println()
				return err
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

	if len(rewardsInfoResponse.UnclaimedIntervals) == 0 {
		fmt.Println("Your node does not have any unclaimed rewards yet.")
		return nil
	}

	// Print the info for all available periods
	totalRpl := big.NewInt(0)
	totalEth := big.NewInt(0)
	for _, intervalInfo := range rewardsInfoResponse.UnclaimedIntervals {
		fmt.Printf("Rewards for Interval %d (%s to %s):\n", intervalInfo.Index, intervalInfo.StartTime.Local(), intervalInfo.EndTime.Local())
		fmt.Printf("\tStaking:        %.6f RPL\n", eth.WeiToEth(&intervalInfo.CollateralRplAmount.Int))
		if intervalInfo.ODaoRplAmount.Cmp(big.NewInt(0)) == 1 {
			fmt.Printf("\tOracle DAO:     %.6f RPL\n", eth.WeiToEth(&intervalInfo.ODaoRplAmount.Int))
		}
		fmt.Printf("\tSmoothing Pool: %.6f ETH\n\n", eth.WeiToEth(&intervalInfo.SmoothingPoolEthAmount.Int))

		totalRpl.Add(totalRpl, &intervalInfo.CollateralRplAmount.Int)
		totalRpl.Add(totalRpl, &intervalInfo.ODaoRplAmount.Int)
		totalEth.Add(totalEth, &intervalInfo.SmoothingPoolEthAmount.Int)
	}

	fmt.Println("Total Pending Rewards:")
	fmt.Printf("\t%.6f RPL\n", eth.WeiToEth(totalRpl))
	fmt.Printf("\t%.6f ETH\n\n", eth.WeiToEth(totalEth))

	// Get the list of intervals to claim
	var indices []uint64
	validIndices := []string{}
	for _, intervalInfo := range rewardsInfoResponse.UnclaimedIntervals {
		validIndices = append(validIndices, fmt.Sprint(intervalInfo.Index))
	}
	for {
		indexSelection := ""
		if !c.Bool("yes") {
			indexSelection = cliutils.Prompt("Which intervals would you like to claim? Use a comma separated list (such as '1,2,3') or leave it blank to claim all intervals at once.", "^$|^\\d+(,\\d+)*$", "Invalid index selection")
		}

		indices = []uint64{}
		if indexSelection == "" {
			for _, intervalInfo := range rewardsInfoResponse.UnclaimedIntervals {
				indices = append(indices, intervalInfo.Index)
			}
			break
		} else {
			elements := strings.Split(indexSelection, ",")
			allValid := true
			seenIndices := map[uint64]bool{}

			for _, element := range elements {
				found := false
				for _, validIndex := range validIndices {
					if validIndex == element {
						found = true
						break
					}
				}
				if !found {
					fmt.Printf("'%s' is an invalid index.\nValid indices are: %s\n", element, strings.Join(validIndices, ","))
					allValid = false
					break
				}
				index, err := strconv.ParseUint(element, 0, 64)
				if err != nil {
					fmt.Printf("'%s' is an invalid index.\nValid indices are: %s\n", element, strings.Join(validIndices, ","))
					allValid = false
					break
				}

				// Ignore duplicates
				_, exists := seenIndices[index]
				if !exists {
					indices = append(indices, index)
					seenIndices[index] = true
				}
			}
			if allValid {
				break
			}
		}
	}

	// Calculate amount to be claimed
	claimRpl := big.NewInt(0)
	claimEth := big.NewInt(0)
	for _, intervalInfo := range rewardsInfoResponse.UnclaimedIntervals {
		for _, index := range indices {
			if intervalInfo.Index == index {
				claimRpl.Add(claimRpl, &intervalInfo.CollateralRplAmount.Int)
				claimRpl.Add(claimRpl, &intervalInfo.ODaoRplAmount.Int)
				claimEth.Add(claimEth, &intervalInfo.SmoothingPoolEthAmount.Int)
			}
		}
	}
	fmt.Printf("With this selection, you will claim %.6f RPL and %.6f ETH.\n\n", eth.WeiToEth(claimRpl), eth.WeiToEth(claimEth))

	// Get restake amount
	restakeAmountWei, err := getRestakeAmount(c, rewardsInfoResponse, claimRpl)
	if err != nil {
		return err
	}

	// Check claim ability
	if restakeAmountWei == nil {
		canClaim, err := rp.CanNodeClaimRewards(indices)
		if err != nil {
			return err
		}

		// Assign max fees
		err = gas.AssignMaxFeeAndLimit(canClaim.GasInfo, rp, c.Bool("yes"))
		if err != nil {
			return err
		}
	} else {
		canClaim, err := rp.CanNodeClaimAndStakeRewards(indices, restakeAmountWei)
		if err != nil {
			return err
		}

		// Assign max fees
		err = gas.AssignMaxFeeAndLimit(canClaim.GasInfo, rp, c.Bool("yes"))
		if err != nil {
			return err
		}
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to claim your rewards?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Claim rewards
	var txHash common.Hash
	if restakeAmountWei == nil {
		response, err := rp.NodeClaimRewards(indices)
		if err != nil {
			return err
		}
		txHash = response.TxHash
	} else {
		response, err := rp.NodeClaimAndStakeRewards(indices, restakeAmountWei)
		if err != nil {
			return err
		}
		txHash = response.TxHash
	}

	fmt.Printf("Claiming Rewards...\n")
	cliutils.PrintTransactionHash(rp, txHash)
	if _, err = rp.WaitForTransaction(txHash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("Successfully claimed rewards.")
	return nil
}

// Determine how much RPL to restake
func getRestakeAmount(c *cli.Context, rewardsInfoResponse api.NodeGetRewardsInfoResponse, claimRpl *big.Int) (*big.Int, error) {

	// Get the current collateral
	currentBondedCollateral := float64(0)
	currentBorrowedCollateral := float64(0)
	rplToMaxCollateral := float64(0)
	rplPrice := eth.WeiToEth(rewardsInfoResponse.RplPrice)
	currentRplStake := eth.WeiToEth(rewardsInfoResponse.RplStake)
	availableRpl := eth.WeiToEth(claimRpl)

	// Print info about autostaking RPL
	var bestTotal float64
	var bestBondedCollateral float64
	var bestBorrowedCollateral float64
	if rewardsInfoResponse.ActiveMinipools > 0 {
		currentBondedCollateral = rewardsInfoResponse.BondedCollateralRatio
		currentBorrowedCollateral = rewardsInfoResponse.BorrowedCollateralRatio
		maxRplRequired := eth.WeiToEth(rewardsInfoResponse.MaximumRplStake)
		rplToMaxCollateral = maxRplRequired - currentRplStake

		bestTotal = availableRpl + currentRplStake
		bestBondedCollateral = rplPrice * bestTotal / (float64(rewardsInfoResponse.ActiveMinipools)*32.0 - eth.WeiToEth(rewardsInfoResponse.EthMatched) - eth.WeiToEth(rewardsInfoResponse.PendingMatchAmount))
		bestBorrowedCollateral = rplPrice * bestTotal / (eth.WeiToEth(rewardsInfoResponse.EthMatched) + eth.WeiToEth(rewardsInfoResponse.PendingMatchAmount))

		fmt.Printf("You currently have %.6f RPL staked (%.2f%% borrowed collateral, %.2f%% bonded collateral).\n", currentRplStake, currentBorrowedCollateral*100, currentBondedCollateral*100)
		if rplToMaxCollateral <= 0 {
			fmt.Println("You are already at maximum collateral. Restaking more RPL will not lead to more rewards.")
		} else if availableRpl < rplToMaxCollateral {
			fmt.Printf("You can restake a max of %.6f RPL which will bring you to a total of %.6f RPL staked (%.2f%% borrowed collateral, %.2f%% bonded collateral).\n", availableRpl, bestTotal, bestBorrowedCollateral*100, bestBondedCollateral*100)
		} else {
			total := rplToMaxCollateral + currentRplStake
			fmt.Printf("If you restake %.6f RPL, you will have a total of %.6f RPL staked (the max bonded collateral of 150%%).\nRestaking more than this will not result in higher rewards.\n\n", rplToMaxCollateral, total)
		}
	} else {
		fmt.Println("You do not have any active minipools, so restaking RPL will not lead to any rewards.")
	}

	// Handle restaking automation or prompts
	var restakeAmountWei *big.Int
	restakeAmountFlag := c.String("restake-amount")

	if restakeAmountFlag == "150%" {
		// Figure out how much to stake to get to 150% or the max available to claim, whichever is smaller
		if rplToMaxCollateral <= 0 {
			fmt.Println("Ignoring automatic staking request since your collateral is already maximized.")
			restakeAmountWei = nil
		} else if availableRpl < rplToMaxCollateral {
			fmt.Printf("Automatically restaking all of the claimable RPL, which will bring you to a total of %.6f RPL staked (%.2f%% bonded collateral).\n", bestTotal, bestBondedCollateral*100)
			restakeAmountWei = claimRpl
		} else {
			total := rplToMaxCollateral + currentRplStake
			fmt.Printf("Automatically restaking %.6f RPL, which will bring you to a total of %.6f RPL staked (150%% bonded collateral).\n", rplToMaxCollateral, total)
			restakeAmountWei = eth.EthToWei(rplToMaxCollateral)
		}
	} else if restakeAmountFlag == "all" {
		// Restake everything with no regard for collateral level
		total := availableRpl + currentRplStake
		totalBondedCollateral := rplPrice * total / (float64(rewardsInfoResponse.ActiveMinipools)*32.0 - eth.WeiToEth(rewardsInfoResponse.EthMatched) - eth.WeiToEth(rewardsInfoResponse.PendingMatchAmount))
		totalBorrowedCollateral := rplPrice * total / (eth.WeiToEth(rewardsInfoResponse.EthMatched) + eth.WeiToEth(rewardsInfoResponse.PendingMatchAmount))
		fmt.Printf("Automatically restaking all of the claimable RPL, which will bring you to a total of %.6f RPL staked (%.2f%% borrowed collateral, %.2f%% bonded collateral).\n", total, totalBorrowedCollateral*100, totalBondedCollateral*100)
		restakeAmountWei = claimRpl
	} else if restakeAmountFlag != "" {
		// Restake a specific amount, capped at how much is available to claim
		stakeAmount, err := strconv.ParseFloat(restakeAmountFlag, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid restake amount '%s': %w", restakeAmountFlag, err)
		}
		if availableRpl < stakeAmount {
			fmt.Printf("Limiting the automatic restake to all of the claimable RPL, which will bring you to a total of %.6f RPL staked (%.2f%% collateral).\n", bestTotal, bestBondedCollateral*100)
			restakeAmountWei = claimRpl
		} else {
			total := stakeAmount + currentRplStake
			totalBondedCollateral := rplPrice * total / (float64(rewardsInfoResponse.ActiveMinipools)*32.0 - eth.WeiToEth(rewardsInfoResponse.EthMatched) - eth.WeiToEth(rewardsInfoResponse.PendingMatchAmount))
			totalBorrowedCollateral := rplPrice * total / (eth.WeiToEth(rewardsInfoResponse.EthMatched) + eth.WeiToEth(rewardsInfoResponse.PendingMatchAmount))
			fmt.Printf("Automatically restaking %.6f RPL, which will bring you to a total of %.6f RPL staked (%.2f%% borrowed collateral, %.2f%% bonded collateral).\n", stakeAmount, total, totalBorrowedCollateral*100, totalBondedCollateral*100)
			restakeAmountWei = eth.EthToWei(stakeAmount)
		}
	} else if c.Bool("yes") {
		// Ignore automatic restaking if `-y` is specified but `-a` isn't
		fmt.Println("Automatic restaking is not requested.")
		restakeAmountWei = nil
	} else {
		// Prompt the user
		if rplToMaxCollateral <= 0 || availableRpl < rplToMaxCollateral {
			var collateralString string
			collateralString = fmt.Sprintf("All %.6f RPL, which will bring you to %.2f%% borrowed collateral (%.2f%% bonded collateral)", availableRpl, bestBorrowedCollateral*100, bestBondedCollateral*100)

			amountOptions := []string{
				"None (do not restake any RPL)",
				collateralString,
				"A custom amount",
			}
			selected, _ := cliutils.Select("Please choose an amount to restake here:", amountOptions)
			switch selected {
			case 0:
				restakeAmountWei = nil
			case 1:
				restakeAmountWei = claimRpl
			case 2:
				for {
					inputAmount := cliutils.Prompt("Please enter an amount of RPL to stake:", "^\\d+(\\.\\d+)?$", "Invalid amount")
					stakeAmount, err := strconv.ParseFloat(inputAmount, 64)
					if err != nil {
						fmt.Printf("Invalid stake amount '%s': %s\n", inputAmount, err.Error())
					} else if stakeAmount < 0 {
						fmt.Println("Amount must be greater than zero.")
					} else if stakeAmount > availableRpl {
						fmt.Println("Amount must be less than the RPL available to claim.")
					} else {
						restakeAmountWei = eth.EthToWei(stakeAmount)
						break
					}
				}
			}
		} else {
			bestTotal = availableRpl + currentRplStake
			var collateralString string
			bestBondedCollateral = rplPrice * bestTotal / (float64(rewardsInfoResponse.ActiveMinipools)*32.0 - eth.WeiToEth(rewardsInfoResponse.EthMatched) - eth.WeiToEth(rewardsInfoResponse.PendingMatchAmount))
			bestBorrowedCollateral = rplPrice * bestTotal / (eth.WeiToEth(rewardsInfoResponse.EthMatched) + eth.WeiToEth(rewardsInfoResponse.PendingMatchAmount))
			collateralString = fmt.Sprintf("All %.6f RPL, which will bring you to %.2f%% borrowed collateral (%.2f%% bonded collateral)", availableRpl, bestBorrowedCollateral*100, bestBondedCollateral*100)
			amountOptions := []string{
				"None (do not restake any RPL)",
				fmt.Sprintf("Enough to get to 150%% bonded collateral (%.6f RPL)", rplToMaxCollateral),
				collateralString,
				"A custom amount",
			}
			selected, _ := cliutils.Select("Please choose an amount to restake here:", amountOptions)
			switch selected {
			case 0:
				restakeAmountWei = nil
			case 1:
				restakeAmountWei = eth.EthToWei(rplToMaxCollateral)
			case 2:
				restakeAmountWei = claimRpl
			case 3:
				for {
					inputAmount := cliutils.Prompt("Please enter an amount of RPL to stake:", "^\\d+(\\.\\d+)?$", "Invalid amount")
					stakeAmount, err := strconv.ParseFloat(inputAmount, 64)
					if err != nil {
						fmt.Printf("Invalid stake amount '%s': %s\n", inputAmount, err.Error())
					} else if stakeAmount < 0 {
						fmt.Println("Amount must be greater than zero.")
					} else if stakeAmount > availableRpl {
						fmt.Println("Amount must be less than the RPL available to claim.")
					} else {
						restakeAmountWei = eth.EthToWei(stakeAmount)
						break
					}
				}
			}
		}
	}

	return restakeAmountWei, nil

}
