package minipool

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	rocketpoolapi "github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func distributeBalance(c *cli.Context) error {

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

	// Get balance distribution details
	details, err := rp.GetDistributeBalanceDetails()
	if err != nil {
		return err
	}

	// Exit if Atlas hasn't been deployed
	if !details.IsAtlasDeployed {
		fmt.Println("Minipool balances cannot be distributed until the Atlas upgrade has been activated.")
		return nil
	}

	// Sort minipools by status
	eligibleMinipools := []api.MinipoolBalanceDistributionDetails{}
	versionTooLowMinipools := []api.MinipoolBalanceDistributionDetails{}
	zero := big.NewInt(0)
	for _, mp := range details.Details {
		if mp.InvalidStatus {
			continue
		} else if mp.VersionTooLow {
			versionTooLowMinipools = append(versionTooLowMinipools, mp)
		} else if mp.Balance.Cmp(zero) == 1 {
			eligibleMinipools = append(eligibleMinipools, mp)
		}
	}

	// Print ineligible ones
	if len(versionTooLowMinipools) > 0 {
		fmt.Printf("%sWARNING: The following minipools are using an old delegate and cannot have their rewards safely distributed:\n", colorYellow)
		for _, mp := range versionTooLowMinipools {
			fmt.Printf("\t%s\n", mp.Address)
		}
		fmt.Printf("\nPlease upgrade the delegate for these minipools using `rocketpool minipool delegate-upgrade` in order to distribute their ETH balances.%s\n", colorReset)
	}

	// Get selected minipools
	var selectedMinipools []api.MinipoolBalanceDistributionDetails
	if c.String("minipool") == "" {

		// Prompt for minipool selection
		options := make([]string, len(eligibleMinipools)+1)
		options[0] = "All available minipools"
		for mi, minipool := range eligibleMinipools {
			options[mi+1] = fmt.Sprintf("%s (%.6f ETH available)", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(minipool.Balance), 6))
		}
		selected, _ := cliutils.Select("Please select a minipool to distribute the balance of:", options)

		// Get minipools
		if selected == 0 {
			selectedMinipools = eligibleMinipools
		} else {
			selectedMinipools = []api.MinipoolBalanceDistributionDetails{eligibleMinipools[selected-1]}
		}

	} else {

		// Get matching minipools
		if c.String("minipool") == "all" {
			selectedMinipools = eligibleMinipools
		} else {
			selectedAddress := common.HexToAddress(c.String("minipool"))
			for _, minipool := range eligibleMinipools {
				if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
					selectedMinipools = []api.MinipoolBalanceDistributionDetails{minipool}
					break
				}
			}
			if selectedMinipools == nil {
				return fmt.Errorf("The minipool %s is not available for balance distribution.", selectedAddress.Hex())
			}
		}

	}

	// Get the total gas limit estimate
	var totalGas uint64 = 0
	var totalSafeGas uint64 = 0
	var gasInfo rocketpoolapi.GasInfo
	for _, minipool := range selectedMinipools {
		estimateGasResponse, err := rp.EstimateDistributeBalanceGas(minipool.Address)
		if err != nil {
			fmt.Printf("WARNING: Couldn't get gas price for distribution transaction for minipool %s (%s)", minipool.Address.Hex(), err.Error())
			break
		} else {
			gasInfo = estimateGasResponse.GasInfo
			totalGas += estimateGasResponse.GasInfo.EstGasLimit
			totalSafeGas += estimateGasResponse.GasInfo.SafeGasLimit
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
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to distribute the ETH balance of %d minipools?", len(selectedMinipools)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Distribute minipool balances
	for _, minipool := range selectedMinipools {

		response, err := rp.DistributeBalance(minipool.Address)
		if err != nil {
			fmt.Printf("Could not distribute the ETH balance of minipool %s: %s.\n", minipool.Address.Hex(), err.Error())
			continue
		}

		fmt.Printf("Distributing balance of minipool %s...\n", minipool.Address.Hex())
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			fmt.Printf("Could not distribute the ETH balance of minipool %s: %s.\n", minipool.Address.Hex(), err.Error())
		} else {
			fmt.Printf("Successfully distributed the ETH balance of minipool %s.\n", minipool.Address.Hex())
		}
	}

	// Return
	return nil

}
