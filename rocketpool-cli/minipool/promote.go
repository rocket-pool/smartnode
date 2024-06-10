package minipool

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	rocketpoolapi "github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func promoteMinipools(c *cli.Context) error {

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

	// Get promotable minipools
	promotableMinipools := []api.MinipoolDetails{}
	for _, minipool := range status.Minipools {
		if minipool.CanPromote {
			promotableMinipools = append(promotableMinipools, minipool)
		}
	}

	// Check for promotable minipools
	if len(promotableMinipools) == 0 {
		fmt.Println("No minipools can be promoted.")
		return nil
	}

	// Get selected minipools
	var selectedMinipools []api.MinipoolDetails
	if c.String("minipool") == "" {

		// Prompt for minipool selection
		options := make([]string, len(promotableMinipools)+1)
		options[0] = "All available minipools"
		for mi, minipool := range promotableMinipools {
			options[mi+1] = fmt.Sprintf("%s (%s until dissolved)", minipool.Address.Hex(), minipool.TimeUntilDissolve)
		}
		selected, _ := cliutils.Select("Please select a minipool to promote:", options)

		// Get minipools
		if selected == 0 {
			selectedMinipools = promotableMinipools
		} else {
			selectedMinipools = []api.MinipoolDetails{promotableMinipools[selected-1]}
		}

	} else {

		// Get matching minipools
		if c.String("minipool") == "all" {
			selectedMinipools = promotableMinipools
		} else {
			selectedAddress := common.HexToAddress(c.String("minipool"))
			for _, minipool := range promotableMinipools {
				if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
					selectedMinipools = []api.MinipoolDetails{minipool}
					break
				}
			}
			if selectedMinipools == nil {
				return fmt.Errorf("The minipool %s is not available to promote.", selectedAddress.Hex())
			}
		}

	}

	// Get the total gas limit estimate
	var totalGas uint64 = 0
	var totalSafeGas uint64 = 0
	var gasInfo rocketpoolapi.GasInfo
	for _, minipool := range selectedMinipools {
		canResponse, err := rp.CanPromoteMinipool(minipool.Address)
		if err != nil {
			fmt.Printf("WARNING: Couldn't get gas price for promote transaction (%s)", err)
			break
		} else {
			gasInfo = canResponse.GasInfo
			totalGas += canResponse.GasInfo.EstGasLimit
			totalSafeGas += canResponse.GasInfo.SafeGasLimit
		}
	}
	gasInfo.EstGasLimit = totalGas
	gasInfo.SafeGasLimit = totalSafeGas

	// Get max fees
	g, err := gas.GetMaxFeeAndLimit(gasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to promote %d minipools?", len(selectedMinipools)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Promote minipools
	for _, minipool := range selectedMinipools {
		g.Assign(rp)
		response, err := rp.PromoteMinipool(minipool.Address)
		if err != nil {
			fmt.Printf("Could not promote minipool %s: %s.\n", minipool.Address.Hex(), err)
			continue
		}

		fmt.Printf("Promoting minipool %s...\n", minipool.Address.Hex())
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			fmt.Printf("Could not promote minipool %s: %s.\n", minipool.Address.Hex(), err)
		} else {
			fmt.Printf("Successfully promoted minipool %s.\n", minipool.Address.Hex())
		}
	}

	// Return
	return nil

}
