package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	rocketpoolapi "github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func delegateUpgradeMinipools(c *cli.Context) error {

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
	latestDelegateResponse, err := rp.GetLatestDelegate()
	if err != nil {
		return err
	}

	includeFinalized := c.Bool("include-finalized")

	minipools := []api.MinipoolDetails{}
	for _, mp := range status.Minipools {
		if mp.Delegate != latestDelegateResponse.Address && !mp.UseLatestDelegate {
			if includeFinalized || !mp.Finalised {
				minipools = append(minipools, mp)
			}
		}
	}

	if len(minipools) == 0 {
		fmt.Println("No minipools are eligible for delegate upgrades.")
		return nil
	}

	// Get selected minipools
	var selectedMinipools []common.Address

	if c.String("minipool") != "" && c.String("minipool") != "all" {
		selectedAddress := common.HexToAddress(c.String("minipool"))
		selectedMinipools = []common.Address{selectedAddress}
	} else {
		if c.String("minipool") == "" {
			// Prompt for minipool selection
			options := make([]string, len(minipools)+1)
			options[0] = "All available minipools"
			for mi, minipool := range minipools {
				options[mi+1] = fmt.Sprintf("%s (using delegate %s)", minipool.Address.Hex(), minipool.Delegate.Hex())
			}
			selected, _ := prompt.Select("Please select a minipool to upgrade:", options)

			// Get minipools
			if selected == 0 {
				selectedMinipools = make([]common.Address, len(minipools))
				for mi, minipool := range minipools {
					selectedMinipools[mi] = minipool.Address
				}
			} else {
				selectedMinipools = []common.Address{minipools[selected-1].Address}
			}
		} else {
			// All minipools
			selectedMinipools = make([]common.Address, len(minipools))
			for mi, minipool := range minipools {
				selectedMinipools[mi] = minipool.Address
			}
		}
	}

	// Get the total gas limit estimate
	var totalGas uint64 = 0
	var totalSafeGas uint64 = 0
	var gasInfo rocketpoolapi.GasInfo
	for _, minipool := range selectedMinipools {
		canResponse, err := rp.CanDelegateUpgradeMinipool(minipool)
		if err != nil {
			fmt.Printf("WARNING: Couldn't get gas price for upgrade transaction (%s)\n", err)
			break
		} else {
			fmt.Printf("Minipool %s will upgrade to delegate contract %s.\n", minipool.Hex(), canResponse.LatestDelegateAddress.Hex())
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
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to upgrade %d minipools?", len(selectedMinipools)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Upgrade minipools
	for _, minipool := range selectedMinipools {
		g.Assign(rp)
		response, err := rp.DelegateUpgradeMinipool(minipool)
		if err != nil {
			fmt.Printf("Could not upgrade minipool %s: %s.\n", minipool.Hex(), err)
			continue
		}

		fmt.Printf("Upgrading minipool %s...\n", minipool.Hex())
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			fmt.Printf("Could not upgrade minipool %s: %s.\n", minipool.Hex(), err)
		} else {
			fmt.Printf("Successfully upgraded minipool %s.\n", minipool.Hex())
		}
	}

	// Return
	return nil

}
