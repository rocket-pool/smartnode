package minipool

import (
	"bytes"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	rocketpoolapi "github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

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

	fmt.Println("NOTE: this function is used to complete the bond reduction process for a minipool. Note that `rocketpool minipool begin-bond-reduction` has been removed after Saturn 1.")
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
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to reduce the bond for %d minipools from 16 ETH to 8 ETH?", len(selectedMinipools))) {
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
