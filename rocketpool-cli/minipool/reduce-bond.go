package minipool

import (
	"bytes"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	rocketpoolapi "github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/utils/math"
	"github.com/urfave/cli"
)

func beginReduceBondAmount(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check the fee distributor
	distribResponse, err := rp.IsFeeDistributorInitialized()
	if err != nil {
		return fmt.Errorf("error checking the node's fee distributor status: %w", err)
	}
	if !distribResponse.IsInitialized {
		fmt.Println("Minipools cannot have their bonds reduced until your fee distributor has been initialized.\nPlease run `rocketpool node initialize-fee-distributor` first, then return here to reduce your bonds.")
		return nil
	}

	// Get minipool statuses
	status, err := rp.MinipoolStatus()
	if err != nil {
		return err
	}

	// Get the bond reduction variables
	settingsResponse, err := rp.GetTNDAOMinipoolSettings()
	if err != nil {
		return err
	}

	// TODO POST-ATLAS: Ask the user how much they want the new bond to be; since there's only one option right now there's no point
	fmt.Printf("This will allow you to begin the bond reduction process to reduce your 16 ETH bond for a minipool down to 8 ETH, awarding you 8 ETH in credit and allowing you to create a second minipool for free (plus gas costs).\n\nThere will be a %.0f-hour wait period after you start the process. After this wait period is over, you will have %.0f hours to complete the process. Your `node` container will do this automatically unless you have it disabled, in which case you must manually run `rocketpool minipool reduce-bond`.\n\n%sNOTE: If you don't run it during this window, your request will time out and you will have to start over.%s\n\n", (time.Duration(settingsResponse.BondReductionWindowStart) * time.Second).Hours(), (time.Duration(settingsResponse.BondReductionWindowLength) * time.Second).Hours(), colorYellow, colorReset)
	newBondAmount := eth.EthToWei(8)

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Do you understand how the bond reduction process will work?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	bondReductionTimeout := time.Duration(settingsResponse.BondReductionWindowStart+settingsResponse.BondReductionWindowLength) * time.Second

	// Get reduceable minipools
	reduceableMinipools := []api.MinipoolDetails{}
	scrubbedMinipools := []api.MinipoolDetails{}

	for _, minipool := range status.Minipools {
		if minipool.ReduceBondCancelled {
			scrubbedMinipools = append(scrubbedMinipools, minipool)
		} else {
			nodeDepositBalance := eth.WeiToEth(minipool.Node.DepositBalance)
			if nodeDepositBalance == 16 &&
				time.Since(minipool.ReduceBondTime) > bondReductionTimeout &&
				minipool.Status.Status == types.Staking &&
				!minipool.Finalised {
				reduceableMinipools = append(reduceableMinipools, minipool)
			}
		}
	}

	// Print scrubs
	if len(scrubbedMinipools) > 0 {
		fmt.Printf("%sNOTE: The following minipools had a previous bond reducton attempt scrubbed by the Oracle DAO and are no longer reduceable:\n", colorYellow)
		for _, mp := range scrubbedMinipools {
			fmt.Printf("\t%s\n", mp.Address)
		}
		fmt.Printf("%s\n\n", colorReset)
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
		selected, _ := prompt.Select("Please select a minipool to begin the ETH bond reduction for:", options)

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
	totalBorrowRequest := big.NewInt(0)
	for _, minipool := range selectedMinipools {
		canResponse, err := rp.CanBeginReduceBondAmount(minipool.Address, newBondAmount)
		if err != nil {
			return fmt.Errorf("couldn't check if minipool %s could have its bond reduced: %s)", minipool.Address.Hex(), err.Error())
		} else {
			if !canResponse.CanReduce {
				fmt.Printf("Cannot reduce bond for minipool %s:\n", minipool.Address.Hex())
				if canResponse.BondReductionDisabled {
					fmt.Println("Bond reductions are currently disabled.")
				}
				if canResponse.MinipoolVersionTooLow {
					fmt.Println("The minipool version is too low. It must be upgraded first using `rocketpool minipool delegate-upgrade`.")
				}
				if canResponse.BalanceTooLow {
					fmt.Printf("The minipool's validator balance on the Beacon Chain is too low (must be 32 ETH or higher, currently %.6f ETH).\n", math.RoundDown(float64(canResponse.Balance)/1e9, 6))
				}
				if canResponse.InvalidBeaconState {
					fmt.Printf("The minipool's validator is not in a legal state on the Beacon Chain. It must be pending or active (current state: %s)\n", canResponse.BeaconState)
				}
				return nil
			}
			gasInfo = canResponse.GasInfo
			totalGas += canResponse.GasInfo.EstGasLimit
			totalSafeGas += canResponse.GasInfo.SafeGasLimit
			totalBorrowRequest.Add(totalBorrowRequest, canResponse.BorrowRequest)
		}
	}
	gasInfo.EstGasLimit = totalGas
	gasInfo.SafeGasLimit = totalSafeGas

	// Make sure there's enough collateral to cover all of the pending bond reductions
	collateralResponse, err := rp.CheckCollateral()
	if err != nil {
		return fmt.Errorf("error checking the node's total collateral: %w", err)
	}
	totalBorrowAvailable := big.NewInt(0).Sub(collateralResponse.EthBorrowedLimit, collateralResponse.EthBorrowed)
	totalBorrowAvailable.Sub(totalBorrowAvailable, collateralResponse.PendingBorrowAmount)
	if totalBorrowAvailable.Cmp(totalBorrowRequest) < 0 {
		fmt.Printf("You do not have enough RPL staked to support all of the selected bond reductions.\nYou can borrow %.6f more ETH, but are requesting %.6f ETH with these bond reductions.\nIn total, they would bring you below the minimum RPL staking requirement (including the RPL required for any pending bond reductions you've already started).\nYou will have to stake more RPL first.\n", eth.WeiToEth(totalBorrowAvailable), eth.WeiToEth(totalBorrowRequest))
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(gasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to begin bond reduction for %d minipools from 16 ETH to 8 ETH?", len(selectedMinipools)))) {
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

	// Get the bond reduction variables
	settingsResponse, err := rp.GetTNDAOMinipoolSettings()
	if err != nil {
		return err
	}

	fmt.Println("NOTE: this function is used to complete the bond reduction process for a minipool. If you haven't started the process already, please run `rocketpool minipool begin-bond-reduction` first.")
	fmt.Println()

	// Get reduceable minipools
	reduceableMinipools := []api.MinipoolDetails{}
	for _, minipool := range status.Minipools {
		timeSinceBondReductionStart := time.Since(minipool.ReduceBondTime)
		nodeDepositBalance := eth.WeiToEth(minipool.Node.DepositBalance)
		if nodeDepositBalance == 16 && timeSinceBondReductionStart > (time.Duration(settingsResponse.BondReductionWindowStart)*time.Second) && timeSinceBondReductionStart < (time.Duration(settingsResponse.BondReductionWindowStart+settingsResponse.BondReductionWindowLength)*time.Second) && !minipool.ReduceBondCancelled {
			reduceableMinipools = append(reduceableMinipools, minipool)
		}
	}

	if len(reduceableMinipools) == 0 {
		fmt.Println("No minipools can have their bond reduced at this time.")
		return nil
	}

	// Workaround for the fee distribution issue
	err = forceFeeDistribution(c, rp)
	if err != nil {
		return err
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
		selected, _ := prompt.Select("Please select a minipool to reduce the ETH bond for:", options)

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
			return fmt.Errorf("error checking if minipool %s can have its bond reduced: %w", minipool.Address.Hex(), err)
		} else if !canResponse.CanReduce {
			fmt.Printf("Minipool %s cannot have its bond reduced:\n", minipool.Address.Hex())
			fmt.Println("The minipool version is too low. Please run `rocketpool minipool delegate-upgrade` to update it.")
			return nil
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
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to reduce the bond for %d minipools from 16 ETH to 8 ETH?", len(selectedMinipools)))) {
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

func forceFeeDistribution(c *cli.Context, rp *rocketpool.Client) error {
	// Get the gas estimate
	canDistributeResponse, err := rp.CanDistribute()
	if err != nil {
		return err
	}

	balance := eth.WeiToEth(canDistributeResponse.Balance)
	if balance == 0 {
		fmt.Println("Your fee distributor does not have any ETH and does not need to be distributed.")
		fmt.Println()
		return nil
	}
	fmt.Println("NOTE: prior to bond reduction, you must distribute the funds in your fee distributor.")
	fmt.Println()

	// Print info
	rEthShare := balance - canDistributeResponse.NodeShare
	fmt.Printf("Your fee distributor's balance of %.6f ETH will be distributed as follows:\n", balance)
	fmt.Printf("\tYour withdrawal address will receive %.6f ETH.\n", canDistributeResponse.NodeShare)
	fmt.Printf("\trETH pool stakers will receive %.6f ETH.\n\n", rEthShare)

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canDistributeResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to distribute the ETH from your node's fee distributor?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Distribute
	response, err := rp.Distribute()
	if err != nil {
		return err
	}

	fmt.Printf("Distributing rewards...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("Successfully distributed your fee distributor's balance. Your rewards should arrive in your withdrawal address shortly.")
	return nil
}
