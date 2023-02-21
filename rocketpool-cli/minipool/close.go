package minipool

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	rocketpoolapi "github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func closeMinipools(c *cli.Context) error {

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
	details, err := rp.GetMinipoolCloseDetailsForNode()
	if err != nil {
		return err
	}

	// Exit if Atlas hasn't been deployed
	if !details.IsAtlasDeployed {
		fmt.Println("Minipools cannot be closed until the Atlas upgrade has been activated.")
		return nil
	}

	closableMinipools := []api.MinipoolCloseDetails{}
	versionTooLowMinipools := []api.MinipoolCloseDetails{}
	balanceLessThanRefundMinipools := []api.MinipoolCloseDetails{}

	for _, mp := range details.Details {
		if mp.CanClose {
			closableMinipools = append(closableMinipools, mp)
		} else {
			if mp.MinipoolVersion < 3 {
				versionTooLowMinipools = append(versionTooLowMinipools, mp)
			}
			if mp.Balance.Cmp(mp.Refund) == -1 {
				balanceLessThanRefundMinipools = append(balanceLessThanRefundMinipools, mp)
			}
		}
	}

	// Print ineligible ones
	if len(versionTooLowMinipools) > 0 {
		fmt.Printf("%sWARNING: The following minipools are using an old delegate and cannot be safely closed:\n", colorYellow)
		for _, mp := range versionTooLowMinipools {
			fmt.Printf("\t%s\n", mp.Address)
		}
		fmt.Printf("\nPlease upgrade the delegate for these minipools using `rocketpool minipool delegate-upgrade` in order to close them.%s\n\n", colorReset)
	}
	if len(balanceLessThanRefundMinipools) > 0 {
		fmt.Printf("%sWARNING: The following minipools have refunds larger than their current balances and cannot be closed at this time:\n", colorYellow)
		for _, mp := range balanceLessThanRefundMinipools {
			fmt.Printf("\t%s\n", mp.Address)
		}
		fmt.Printf("\nIf you have recently exited their validators from the Beacon Chain, please wait until their balances have been sent to the minipools before closing them.%s\n\n", colorReset)
	}

	// Check for closable minipools
	if len(closableMinipools) == 0 {
		fmt.Println("No minipools can be closed.")
		return nil
	}

	// Get selected minipools
	var selectedMinipools []api.MinipoolCloseDetails
	if c.String("minipool") == "" {

		// Prompt for minipool selection
		options := make([]string, len(closableMinipools)+1)
		options[0] = "All available minipools"
		for mi, minipool := range closableMinipools {
			if minipool.MinipoolStatus == types.Dissolved {
				options[mi+1] = fmt.Sprintf("%s (%.6f ETH will be returned)", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(minipool.Balance), 6))
			} else {
				options[mi+1] = fmt.Sprintf("%s (%.6f ETH available, %.6f ETH is yours plus a refund of %.6f ETH)", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(minipool.Balance), 6), math.RoundDown(eth.WeiToEth(minipool.NodeShare), 6), math.RoundDown(eth.WeiToEth(minipool.Refund), 6))
			}
		}
		selected, _ := cliutils.Select("Please select a minipool to close:", options)

		// Get minipools
		if selected == 0 {
			selectedMinipools = closableMinipools
		} else {
			selectedMinipools = []api.MinipoolCloseDetails{closableMinipools[selected-1]}
		}

	} else {

		// Get matching minipools
		if c.String("minipool") == "all" {
			selectedMinipools = closableMinipools
		} else {
			selectedAddress := common.HexToAddress(c.String("minipool"))
			for _, minipool := range closableMinipools {
				if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
					selectedMinipools = []api.MinipoolCloseDetails{minipool}
					break
				}
			}
			if selectedMinipools == nil {
				return fmt.Errorf("The minipool %s is not available for closing.", selectedAddress.Hex())
			}
		}

	}

	// Get the total gas limit estimate
	var gasInfo rocketpoolapi.GasInfo
	for _, minipool := range selectedMinipools {
		gasInfo.EstGasLimit += minipool.GasInfo.EstGasLimit
		gasInfo.SafeGasLimit += minipool.GasInfo.SafeGasLimit
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(gasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to close %d minipools?", len(selectedMinipools)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Close minipools
	for _, minipool := range selectedMinipools {

		response, err := rp.CloseMinipool(minipool.Address)
		if err != nil {
			fmt.Printf("Could not close minipool %s: %s.\n", minipool.Address.Hex(), err.Error())
			continue
		}

		fmt.Printf("Closing minipool %s...\n", minipool.Address.Hex())
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			fmt.Printf("Could not close minipool %s: %s.\n", minipool.Address.Hex(), err.Error())
		} else {
			fmt.Printf("Successfully closed minipool %s.\n", minipool.Address.Hex())
		}
	}

	// Return
	return nil

}
