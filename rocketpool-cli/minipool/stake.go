package minipool

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	rocketpoolapi "github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func stakeMinipools(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get minipool statuses
	status, err := rp.MinipoolStatus()
	if err != nil {
		return err
	}

	// Get stakeable minipools
	stakeableMinipools := []api.MinipoolDetails{}
	for _, minipool := range status.Minipools {
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
	var selectedMinipools []api.MinipoolDetails
	if c.String("minipool") == "" {

		// Prompt for minipool selection
		options := make([]string, len(stakeableMinipools)+1)
		options[0] = "All available minipools"
		for mi, minipool := range stakeableMinipools {
			options[mi+1] = fmt.Sprintf("%s (%s until dissolved)", minipool.Address.Hex(), minipool.TimeUntilDissolve)
		}
		selected, _ := prompt.Select("Please select a minipool to stake:", options)

		// Get minipools
		if selected == 0 {
			selectedMinipools = stakeableMinipools
		} else {
			selectedMinipools = []api.MinipoolDetails{stakeableMinipools[selected-1]}
		}

	} else {

		// Get matching minipools
		if c.String("minipool") == "all" {
			selectedMinipools = stakeableMinipools
		} else {
			selectedAddress := common.HexToAddress(c.String("minipool"))
			for _, minipool := range stakeableMinipools {
				if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
					selectedMinipools = []api.MinipoolDetails{minipool}
					break
				}
			}
			if selectedMinipools == nil {
				return fmt.Errorf("The minipool %s is not available to stake.", selectedAddress.Hex())
			}
		}

	}

	// Get the total gas limit estimate
	var totalGas uint64 = 0
	var totalSafeGas uint64 = 0
	var gasInfo rocketpoolapi.GasInfo
	for _, minipool := range selectedMinipools {
		canResponse, err := rp.CanStakeMinipool(minipool.Address)
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

	fmt.Println()
	fmt.Println("NOTE: Your Validator Client must be restarted after this process so it loads the new validator key.")
	fmt.Println("Since you are manually staking the minipool, this must be done manually.")
	fmt.Println("When you have finished staking all your minipools, please restart your validator.")
	fmt.Println()

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to stake %d minipools?", len(selectedMinipools)))) {
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
