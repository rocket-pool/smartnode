package minipool

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	rocketpoolapi "github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/urfave/cli"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/shared/services/gas"
	"github.com/rocket-pool/smartnode/v2/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/v2/shared/utils/cli"
)

func dissolveMinipools(c *cli.Context) error {

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

	// Get initialized minipools
	initializedMinipools := []api.MinipoolDetails{}
	for _, minipool := range status.Minipools {
		if minipool.Status.Status == types.Initialized {
			initializedMinipools = append(initializedMinipools, minipool)
		}
	}

	// Check for initialized minipools
	if len(initializedMinipools) == 0 {
		fmt.Println("No minipools can be dissolved.")
		return nil
	}

	// Get selected minipools
	var selectedMinipools []api.MinipoolDetails
	if c.String("minipool") == "" {

		// Prompt for minipool selection
		options := make([]string, len(initializedMinipools)+1)
		options[0] = "All available minipools"
		for mi, minipool := range initializedMinipools {
			options[mi+1] = fmt.Sprintf("%s (%.6f ETH deposited)", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(minipool.Node.DepositBalance), 6))
		}
		selected, _ := cliutils.Select("Please select a minipool to dissolve:", options)

		// Get minipools
		if selected == 0 {
			selectedMinipools = initializedMinipools
		} else {
			selectedMinipools = []api.MinipoolDetails{initializedMinipools[selected-1]}
		}

	} else {

		// Get matching minipools
		if c.String("minipool") == "all" {
			selectedMinipools = initializedMinipools
		} else {
			selectedAddress := common.HexToAddress(c.String("minipool"))
			for _, minipool := range initializedMinipools {
				if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
					selectedMinipools = []api.MinipoolDetails{minipool}
					break
				}
			}
			if selectedMinipools == nil {
				return fmt.Errorf("The minipool %s is not available for dissolving.", selectedAddress.Hex())
			}
		}

	}

	// Get the total gas limit estimate
	var totalGas uint64 = 0
	var totalSafeGas uint64 = 0
	var gasInfo rocketpoolapi.GasInfo
	for _, minipool := range selectedMinipools {
		canResponse, err := rp.CanDissolveMinipool(minipool.Address)
		if err != nil {
			fmt.Printf("WARNING: Couldn't get gas price for dissolve transaction (%s)", err)
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
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to dissolve %d minipool(s)? This action cannot be undone!", len(selectedMinipools)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Dissolve and close minipools
	for _, minipool := range selectedMinipools {
		response, err := rp.DissolveMinipool(minipool.Address)
		if err != nil {
			fmt.Printf("Could not dissolve minipool %s: %s.\n", minipool.Address.Hex(), err)
			continue
		}

		fmt.Printf("Dissolving minipool %s...\n", minipool.Address.Hex())
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			fmt.Printf("Could not dissolve minipool %s: %s.\n", minipool.Address.Hex(), err)
			continue
		} else {
			fmt.Printf("Successfully dissolved minipool %s.\n", minipool.Address.Hex())
		}

		closeResponse, err := rp.CloseMinipool(minipool.Address)
		if err != nil {
			fmt.Printf("Could not close minipool %s: %s.\n", minipool.Address.Hex(), err)
			continue
		}

		fmt.Printf("Closing minipool %s...\n", minipool.Address.Hex())
		cliutils.PrintTransactionHash(rp, closeResponse.TxHash)
		if _, err = rp.WaitForTransaction(closeResponse.TxHash); err != nil {
			fmt.Printf("Could not close minipool %s: %s.\n", minipool.Address.Hex(), err)
		} else {
			fmt.Printf("Successfully closed minipool %s.\n", minipool.Address.Hex())
		}
	}

	// Return
	return nil

}
