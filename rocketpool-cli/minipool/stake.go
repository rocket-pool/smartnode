package minipool

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

const (
	stakeMinipoolsFlag string = "minipools"
)

func stakeMinipools(c *cli.Context) error {

	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get minipool statuses
	status, err := rp.Api.Minipool.Status()
	if err != nil {
		return err
	}

	// Get stakeable minipools
	stakeableMinipools := []api.MinipoolDetails{}
	for _, minipool := range status.Data.Minipools {
		if minipool.CanStake {
			stakeableMinipools = append(stakeableMinipools, minipool)
		}
	}

	// Check for stakeable minipools
	if len(stakeableMinipools) == 0 {
		fmt.Println("No minipools can be staked.")
		return nil
	}

	// Get selected minipools
	options := make([]utils.SelectionOption[api.MinipoolDetails], len(stakeableMinipools))
	for i, mp := range stakeableMinipools {
		option := &options[i]
		option.Element = &mp
		option.ID = fmt.Sprint(mp.Address)
		option.Display = fmt.Sprintf("%s (%s until dissolved)", mp.Address.Hex(), mp.TimeUntilDissolve)
	}
	selectedMinipools, err := utils.GetMultiselectIndices[api.MinipoolDetails](c, stakeMinipoolsFlag, options, "Please select a minipool to stake:")
	if err != nil {
		return fmt.Errorf("error determining minipool selection: %w", err)
	}

	// Validation
	txs := make([]*core.TransactionInfo, len(selectedMinipools))
	for _, minipool := range selectedMinipools {
		response, err := rp.Api.Minipool.Stake(minipool.Address)
		if err != nil {
			fmt.Printf("WARNING: Couldn't get gas price for stake transaction (%s)", err)
			break
		} else {
			gasInfo = canResponse.GasInfo
			totalGas += canResponse.GasInfo.EstGasLimit
			totalSafeGas += canResponse.GasInfo.SafeGasLimit
		}
	}
	gasInfo.EstGasLimit = totalGas
	gasInfo.SafeGasLimit = totalSafeGas

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(gasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	fmt.Println("\nNOTE: Your validator container will be restarted after this process so it loads the new validator key.\n")

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to stake %d minipools?", len(selectedMinipools)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Stake minipools
	for _, minipool := range selectedMinipools {
		response, err := rp.StakeMinipool(minipool.Address)
		if err != nil {
			fmt.Printf("Could not stake minipool %s: %s.\n", minipool.Address.Hex(), err)
			continue
		}

		fmt.Printf("Staking minipool %s...\n", minipool.Address.Hex())
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			fmt.Printf("Could not stake minipool %s: %s.\n", minipool.Address.Hex(), err)
		} else {
			fmt.Printf("Successfully staked minipool %s.\n", minipool.Address.Hex())
		}
	}

	// Return
	return nil

}
