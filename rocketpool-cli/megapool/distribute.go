package megapool

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

func distribute(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if Saturn is already deployed
	saturnResp, err := rp.IsSaturnDeployed()
	if err != nil {
		return err
	}
	if !saturnResp.IsSaturnDeployed {
		fmt.Println("This command is only available after the Saturn upgrade.")
		return nil
	}

	// Get the gas estimate
	canResponse, err := rp.CanDistributeMegapool()
	if err != nil {
		return fmt.Errorf("error checking if megapool can distribute rewards: %w", err)
	}

	if !canResponse.CanDistribute {
		fmt.Println("Could not distribute rewards")
		if canResponse.MegapoolNotDeployed {
			fmt.Println("The node does not have a megapool deployed")
		}
		if canResponse.LastDistributionTime == 0 {
			fmt.Printf("The node's megapool: %s does not have any staking validators\n", canResponse.MegapoolAddress)
		}
		if canResponse.ExitingValidatorCount > 0 {
			fmt.Printf("The megapool has %d validator(s) exiting", canResponse.ExitingValidatorCount)
			fmt.Println()
			for _, val := range canResponse.Details.Validators {
				if val.Activated && val.BeaconStatus.WithdrawableEpoch != FarFutureEpoch && !val.Exited {
					if !val.Exiting {
						fmt.Printf("Validator ID %d needs an exit proof (run 'rp megapool notify-validator-exit')", val.ValidatorId)
						fmt.Println()
					}
					if val.BeaconStatus.Status == beacon.ValidatorState_ExitedUnslashed {
						fmt.Printf("Validator ID %d has exited but is still pending full beacon withdrawal", val.ValidatorId)
						fmt.Println()
					}
					if val.BeaconStatus.Status == beacon.ValidatorState_WithdrawalDone {
						fmt.Printf("Validator ID %d needs a final balance proof (run 'rp megapool notify-final-balance')", val.ValidatorId)
						fmt.Println()
					}
				}
			}
		}
		if canResponse.LockedValidatorCount > 0 {
			fmt.Printf("The megapool has %d validator(s) locked", canResponse.LockedValidatorCount)
			fmt.Println()
		}
		return nil
	}

	// Get pending rewards
	rewardsSplit, err := rp.CalculatePendingRewards()
	if err != nil {
		return fmt.Errorf("error calculating pending rewards: %w", err)
	}

	if rewardsSplit.RewardSplit.NodeRewards.Cmp(big.NewInt(0)) <= 0 {
		fmt.Println("There are no pending rewards to distribute.")
		return nil
	}
	// Print rewards
	fmt.Printf("You're about to claim pending rewards from the megapool. The rewards will be distributed to the node's withdrawal address. The node share of rewards is %.4f ETH.", eth.WeiToEth(rewardsSplit.RewardSplit.NodeRewards))
	fmt.Println()

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to distribute your megapool rewards?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Distribute
	response, err := rp.DistributeMegapool()
	if err != nil {
		fmt.Printf("Could not distribute megapool rewards: %s. \n", err)
		return nil
	}

	// Log and wait for the transaction
	fmt.Printf("Distributing megapool rewards...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Return
	fmt.Printf("Successfully distributed megapool rewards.\n")
	return nil
}
