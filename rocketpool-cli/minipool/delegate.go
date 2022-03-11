package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	rocketpoolapi "github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func delegateUpgradeMinipools(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get selected minipools
	var selectedMinipools []common.Address

	if c.String("minipool") != "" && c.String("minipool") != "all" {
		selectedAddress := common.HexToAddress(c.String("minipool"))
		selectedMinipools = []common.Address{selectedAddress}
	} else {
		// Get minipool statuses
		status, err := rp.MinipoolStatus()
		if err != nil {
			return err
		}
		minipools := status.Minipools

		if c.String("minipool") == "" {
			// Prompt for minipool selection
			options := make([]string, len(minipools)+1)
			options[0] = "All available minipools"
			for mi, minipool := range minipools {
				options[mi+1] = fmt.Sprintf("%s (using delegate %s)", minipool.Address.Hex(), minipool.Delegate.Hex())
			}
			selected, _ := cliutils.Select("Please select a minipool to upgrade:", options)

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
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to upgrade %d minipools?", len(selectedMinipools)))) {
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

func delegateRollbackMinipools(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get selected minipools
	var selectedMinipools []common.Address

	if c.String("minipool") != "" && c.String("minipool") != "all" {
		selectedAddress := common.HexToAddress(c.String("minipool"))
		selectedMinipools = []common.Address{selectedAddress}
	} else {
		// Get minipool statuses
		status, err := rp.MinipoolStatus()
		if err != nil {
			return err
		}
		minipools := status.Minipools

		if c.String("minipool") == "" {
			// Prompt for minipool selection
			options := make([]string, len(minipools)+1)
			options[0] = "All available minipools"
			for mi, minipool := range minipools {
				options[mi+1] = fmt.Sprintf("%s (using delegate %s)", minipool.Address.Hex(), minipool.Delegate.Hex())
			}
			selected, _ := cliutils.Select("Please select a minipool to upgrade:", options)

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
		canResponse, err := rp.CanDelegateRollbackMinipool(minipool)
		if err != nil {
			fmt.Printf("WARNING: Couldn't get gas price for rollback transaction (%s)", err)
			break
		} else {
			fmt.Printf("Minipool %s will roll back to delegate contract %s.\n", minipool.Hex(), canResponse.RollbackAddress.Hex())
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
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to rollback %d minipools?", len(selectedMinipools)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Rollback minipools
	for _, minipool := range selectedMinipools {
		g.Assign(rp)
		response, err := rp.DelegateRollbackMinipool(minipool)
		if err != nil {
			fmt.Printf("Could not rollback minipool %s: %s.\n", minipool.Hex(), err)
			continue
		}

		fmt.Printf("Rolling back minipool %s...\n", minipool.Hex())
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			fmt.Printf("Could not rollback minipool %s: %s.\n", minipool.Hex(), err)
		} else {
			fmt.Printf("Successfully rolled back minipool %s.\n", minipool.Hex())
		}
	}

	// Return
	return nil

}

func setUseLatestDelegateMinipools(c *cli.Context, setting bool) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get selected minipools
	var selectedMinipools []common.Address

	if c.String("minipool") != "" && c.String("minipool") != "all" {
		selectedAddress := common.HexToAddress(c.String("minipool"))
		selectedMinipools = []common.Address{selectedAddress}
	} else {
		// Get minipool statuses
		status, err := rp.MinipoolStatus()
		if err != nil {
			return err
		}
		minipools := status.Minipools

		if c.String("minipool") == "" {
			// Prompt for minipool selection
			options := make([]string, len(minipools)+1)
			options[0] = "All available minipools"
			for mi, minipool := range minipools {
				options[mi+1] = fmt.Sprintf("%s (using delegate %s)", minipool.Address.Hex(), minipool.Delegate.Hex())
			}
			selected, _ := cliutils.Select("Please select a minipool to upgrade:", options)

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
		canResponse, err := rp.CanSetUseLatestDelegateMinipool(minipool, setting)
		if err != nil {
			fmt.Printf("WARNING: Couldn't get gas price for auto-upgrade setting transaction (%s)", err)
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
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to change the auto-upgrade setting for %d minipools to %t?", len(selectedMinipools), setting))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Update minipools
	for _, minipool := range selectedMinipools {
		g.Assign(rp)
		response, err := rp.SetUseLatestDelegateMinipool(minipool, setting)
		if err != nil {
			fmt.Printf("Could not update the auto-upgrade setting for minipool %s: %s.\n", minipool.Hex(), err)
			continue
		}

		fmt.Printf("Updating the auto-upgrade setting for minipool %s...\n", minipool.Hex())
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			fmt.Printf("Could not update the auto-upgrade setting for minipool %s: %s.\n", minipool.Hex(), err)
		} else {
			fmt.Printf("Successfully updated the setting for minipool %s.\n", minipool.Hex())
		}
	}

	// Return
	return nil

}
