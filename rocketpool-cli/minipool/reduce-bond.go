package minipool

import (
	"bytes"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	rocketpoolapi "github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

func beginReduceBondAmount(c *cli.Context) error {

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

	// Check for Atlas
	atlasResponse, err := rp.IsAtlasDeployed()
	if err != nil {
		return fmt.Errorf("error checking if Atlas has been deployed: %w", err)
	}
	if !atlasResponse.IsAtlasDeployed {
		fmt.Println("You cannot reduce a minipool's bond until Atlas has been deployed.")
		return nil
	}

	// Get minipool statuses
	status, err := rp.MinipoolStatus()
	if err != nil {
		return err
	}

	if !status.IsAtlasDeployed {
		fmt.Println("You cannot reduce a minipool's bond until Atlas has been deployed.")
		return nil
	}

	// Get the bond reduction variables
	settingsResponse, err := rp.GetTNDAOMinipoolSettings()
	if err != nil {
		return err
	}

	// TODO POST-ATLAS: Ask the user how much they want the new bond to be; since there's only one option right now there's no point
	fmt.Printf("This will allow you to begin the bond reduction process to reduce your 16 ETH bond for a minipool down to 8 ETH, awarding you 8 ETH in credit and allowing you to create a second minipool for free (plus gas costs).\n\nThere will be a %.0f-hour wait period after you start the process. After this wait period is over, you will have %.0f hours to complete the process. Your `node` container will do this automatically unless you have it disabled, in which case you must manually run `rocketpool minipool reduce-bond`. If you don't run it during this window, your request will time out and you will have to start over.\n\n", (time.Duration(settingsResponse.BondReductionWindowStart) * time.Second).Hours(), (time.Duration(settingsResponse.BondReductionWindowLength) * time.Second).Hours())
	newBondAmount := eth.EthToWei(8)

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Do you understand how the bond reduction process will work?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	bondReductionTimeout := time.Duration(settingsResponse.BondReductionWindowStart+settingsResponse.BondReductionWindowLength) * time.Second

	// Get reduceable minipools
	reduceableMinipools := []api.MinipoolDetails{}
	for _, minipool := range status.Minipools {
		nodeDepositBalance := eth.WeiToEth(minipool.Node.DepositBalance)
		if nodeDepositBalance == 16 &&
			time.Since(minipool.ReduceBondTime) > bondReductionTimeout &&
			!minipool.ReduceBondCancelled &&
			minipool.Status.Status == types.Staking {
			reduceableMinipools = append(reduceableMinipools, minipool)
		}
	}

	if len(reduceableMinipools) == 0 {
		fmt.Println("No minipools can have their bond reduced at this time.")
		return nil
	}

	// Get selected minipools
	var selectedMinipools []api.MinipoolDetails
	if c.String("minipool") == "" {

		// Prompt for minipool selection
		options := make([]string, len(reduceableMinipools)+1)
		options[0] = "All available minipools"
		for mi, minipool := range reduceableMinipools {
			options[mi+1] = fmt.Sprintf("%s (Current bond: %d ETH, commission: %.2f%%)", minipool.Address.Hex(), int(eth.WeiToEth(minipool.Node.DepositBalance)), minipool.Node.Fee*100)
		}
		selected, _ := cliutils.Select("Please select a minipool to begin the ETH bond reduction for:", options)

		// Get minipools
		if selected == 0 {
			selectedMinipools = reduceableMinipools
		} else {
			selectedMinipools = []api.MinipoolDetails{reduceableMinipools[selected-1]}
		}

	} else {

		// Get matching minipools
		if c.String("minipool") == "all" {
			selectedMinipools = reduceableMinipools
		} else {
			selectedAddress := common.HexToAddress(c.String("minipool"))
			for _, minipool := range reduceableMinipools {
				if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
					selectedMinipools = []api.MinipoolDetails{minipool}
					break
				}
			}
			if selectedMinipools == nil {
				return fmt.Errorf("The minipool %s cannot have its bond reduced.", selectedAddress.Hex())
			}
		}

	}

	// Get the total gas limit estimate
	var totalGas uint64 = 0
	var totalSafeGas uint64 = 0
	var gasInfo rocketpoolapi.GasInfo
	for _, minipool := range selectedMinipools {
		canResponse, err := rp.CanBeginReduceBondAmount(minipool.Address, newBondAmount)
		if err != nil {
			fmt.Printf("WARNING: Couldn't get gas price for beginning bond reduction on minipool %s: %s)", minipool.Address.Hex(), err.Error())
			break
		} else {
			if canResponse.BondReductionDisabled {
				fmt.Println("Cannot reduce minipool bonds:")
				fmt.Println("Bond reductions are currently disabled.")
				return nil
			}
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

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to begin bond reduction for %d minipools from 16 ETH to 8 ETH?", len(selectedMinipools)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Begin bond reduction
	for _, minipool := range selectedMinipools {
		response, err := rp.BeginReduceBondAmount(minipool.Address, newBondAmount)
		if err != nil {
			fmt.Printf("Could not begin bond reduction for minipool %s: %s.\n", minipool.Address.Hex(), err.Error())
			continue
		}

		fmt.Printf("Beginning bond reduction for minipool %s...\n", minipool.Address.Hex())
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			fmt.Printf("Could not begin bond reduction for minipool %s: %s.\n", minipool.Address.Hex(), err.Error())
		} else {
			fmt.Printf("Successfully started bond reduction for minipool %s.\n", minipool.Address.Hex())
		}
	}

	// Return
	return nil

}

func reduceBondAmount(c *cli.Context) error {

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

	// Get minipool statuses
	status, err := rp.MinipoolStatus()
	if err != nil {
		return err
	}

	if !status.IsAtlasDeployed {
		fmt.Println("You cannot reduce a minipool's bond until Atlas has been deployed.")
		return nil
	}

	// Get the bond reduction variables
	settingsResponse, err := rp.GetTNDAOMinipoolSettings()
	if err != nil {
		return err
	}

	fmt.Println("NOTE: this function is used to complete the bond reduction process for a minipool. If you haven't started the process already, please run `rocketpool minipool begin-bond-reduction` first.\n")

	// Get reduceable minipools
	reduceableMinipools := []api.MinipoolDetails{}
	for _, minipool := range status.Minipools {
		timeSinceBondReductionStart := time.Since(minipool.ReduceBondTime)
		nodeDepositBalance := eth.WeiToEth(minipool.Node.DepositBalance)
		if nodeDepositBalance == 16 && timeSinceBondReductionStart > time.Duration(settingsResponse.BondReductionWindowStart) && timeSinceBondReductionStart < time.Duration(settingsResponse.BondReductionWindowStart+settingsResponse.BondReductionWindowLength) && !minipool.ReduceBondCancelled {
			reduceableMinipools = append(reduceableMinipools, minipool)
		}
	}

	if len(reduceableMinipools) == 0 {
		fmt.Println("No minipools can have their bond reduced at this time.")
		return nil
	}

	// Get selected minipools
	var selectedMinipools []api.MinipoolDetails
	if c.String("minipool") == "" {

		// Prompt for minipool selection
		options := make([]string, len(reduceableMinipools)+1)
		options[0] = "All available minipools"
		for mi, minipool := range reduceableMinipools {
			options[mi+1] = fmt.Sprintf("%s (Current bond: %d ETH)", minipool.Address.Hex(), int(eth.WeiToEth(minipool.Node.DepositBalance)))
		}
		selected, _ := cliutils.Select("Please select a minipool to reduce the ETH bond for:", options)

		// Get minipools
		if selected == 0 {
			selectedMinipools = reduceableMinipools
		} else {
			selectedMinipools = []api.MinipoolDetails{reduceableMinipools[selected-1]}
		}

	} else {

		// Get matching minipools
		if c.String("minipool") == "all" {
			selectedMinipools = reduceableMinipools
		} else {
			selectedAddress := common.HexToAddress(c.String("minipool"))
			for _, minipool := range reduceableMinipools {
				if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
					selectedMinipools = []api.MinipoolDetails{minipool}
					break
				}
			}
			if selectedMinipools == nil {
				return fmt.Errorf("The minipool %s cannot have its bond reduced.", selectedAddress.Hex())
			}
		}

	}

	// Get the total gas limit estimate
	var totalGas uint64 = 0
	var totalSafeGas uint64 = 0
	var gasInfo rocketpoolapi.GasInfo
	for _, minipool := range selectedMinipools {
		canResponse, err := rp.CanReduceBondAmount(minipool.Address)
		if err != nil {
			fmt.Printf("WARNING: Couldn't get gas price for reducing bond on minipool %s: %s)", minipool.Address.Hex(), err.Error())
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

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to reduce the bond for %d minipools from 16 ETH to 8 ETH?", len(selectedMinipools)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Begin bond reduction
	for _, minipool := range selectedMinipools {
		response, err := rp.ReduceBondAmount(minipool.Address)
		if err != nil {
			fmt.Printf("Could not reduce bond for minipool %s: %s.\n", minipool.Address.Hex(), err.Error())
			continue
		}

		fmt.Printf("Reducing bond for minipool %s...\n", minipool.Address.Hex())
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			fmt.Printf("Could not reduce bond for minipool %s: %s.\n", minipool.Address.Hex(), err.Error())
		} else {
			fmt.Printf("Successfully reduced bond for minipool %s.\n", minipool.Address.Hex())
		}
	}

	// Return
	return nil

}
