package minipool

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	rocketpoolapi "github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func delegateUpgradeMinipools(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get minipool statuses
    status, err := rp.MinipoolStatus()
    if err != nil {
        return err
    }
    minipools := status.Minipools

    // Get selected minipools
    var selectedMinipools []api.MinipoolDetails
    if c.String("minipool") == "" {

        // Prompt for minipool selection
        options := make([]string, len(minipools) + 1)
        options[0] = "All available minipools"
        for mi, minipool := range minipools {
            options[mi + 1] = fmt.Sprintf("%s (using delegate %s)", minipool.Address.Hex(), minipool.Delegate.Hex())
        }
        selected, _ := cliutils.Select("Please select a minipool to upgrade:", options)

        // Get minipools
        if selected == 0 {
            selectedMinipools = minipools
        } else {
            selectedMinipools = []api.MinipoolDetails{minipools[selected - 1]}
        }

    } else {

        // Get matching minipools
        if c.String("minipool") == "all" {
            selectedMinipools = minipools
        } else {
            selectedAddress := common.HexToAddress(c.String("minipool"))
            for _, minipool := range minipools {
                if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
                    selectedMinipools = []api.MinipoolDetails{minipool}
                    break
                }
            }
            if selectedMinipools == nil {
                return fmt.Errorf("The minipool %s is not available for upgrade.", selectedAddress.Hex())
            }
        }

    }

    // Get the total gas limit estimate
    var totalGas uint64 = 0
    var gasInfo rocketpoolapi.GasInfo
    for _, minipool := range selectedMinipools {
        canResponse, err := rp.CanDelegateUpgradeMinipool(minipool.Address)
        if err != nil {
            fmt.Printf("WARNING: Couldn't get gas price for upgrade transaction (%s)", err)
            break
        } else {
            gasInfo = canResponse.GasInfo
            totalGas += canResponse.GasInfo.EstGasLimit
        }
    }
    gasInfo.EstGasLimit = totalGas

    // Display gas estimate
    rp.PrintGasInfo(gasInfo)

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to upgrade %d minipools?", len(selectedMinipools)))) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Upgrade minipools
    for _, minipool := range selectedMinipools {
        response, err := rp.DelegateUpgradeMinipool(minipool.Address)
        if err != nil {
            fmt.Printf("Could not upgrade minipool %s: %s.\n", minipool.Address.Hex(), err)
            continue
        }
    
        fmt.Printf("Upgrading minipool %s...\n", minipool.Address.Hex())
        cliutils.PrintTransactionHash(rp, response.TxHash)
        if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
            fmt.Printf("Could not upgrade minipool %s: %s.\n", minipool.Address.Hex(), err)
        } else {
            fmt.Printf("Successfully upgraded minipool %s.\n", minipool.Address.Hex())
        }
    }

    // Return
    return nil

}


func delegateRollbackMinipools(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get minipool statuses
    status, err := rp.MinipoolStatus()
    if err != nil {
        return err
    }
    minipools := status.Minipools

    // Get selected minipools
    var selectedMinipools []api.MinipoolDetails
    if c.String("minipool") == "" {

        // Prompt for minipool selection
        options := make([]string, len(minipools) + 1)
        options[0] = "All available minipools"
        for mi, minipool := range minipools {
            options[mi + 1] = fmt.Sprintf("%s (using delegate %s, will rollback to %s)", 
                minipool.Address.Hex(), minipool.Delegate.Hex(), minipool.PreviousDelegate.Hex())
        }
        selected, _ := cliutils.Select("Please select a minipool to rollback:", options)

        // Get minipools
        if selected == 0 {
            selectedMinipools = minipools
        } else {
            selectedMinipools = []api.MinipoolDetails{minipools[selected - 1]}
        }

    } else {

        // Get matching minipools
        if c.String("minipool") == "all" {
            selectedMinipools = minipools
        } else {
            selectedAddress := common.HexToAddress(c.String("minipool"))
            for _, minipool := range minipools {
                if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
                    selectedMinipools = []api.MinipoolDetails{minipool}
                    break
                }
            }
            if selectedMinipools == nil {
                return fmt.Errorf("The minipool %s is not available for rollback.", selectedAddress.Hex())
            }
        }

    }

    // Get the total gas limit estimate
    var totalGas uint64 = 0
    var gasInfo rocketpoolapi.GasInfo
    for _, minipool := range selectedMinipools {
        canResponse, err := rp.CanDelegateRollbackMinipool(minipool.Address)
        if err != nil {
            fmt.Printf("WARNING: Couldn't get gas price for rollback transaction (%s)", err)
            break
        } else {
            gasInfo = canResponse.GasInfo
            totalGas += canResponse.GasInfo.EstGasLimit
        }
    }
    gasInfo.EstGasLimit = totalGas

    // Display gas estimate
    rp.PrintGasInfo(gasInfo)

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to rollback %d minipools?", len(selectedMinipools)))) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Rollback minipools
    for _, minipool := range selectedMinipools {
        response, err := rp.DelegateRollbackMinipool(minipool.Address)
        if err != nil {
            fmt.Printf("Could not rollback minipool %s: %s.\n", minipool.Address.Hex(), err)
            continue
        }
    
        fmt.Printf("Rolling back minipool %s...\n", minipool.Address.Hex())
        cliutils.PrintTransactionHash(rp, response.TxHash)
        if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
            fmt.Printf("Could not rollback minipool %s: %s.\n", minipool.Address.Hex(), err)
        } else {
            fmt.Printf("Successfully rolled back minipool %s.\n", minipool.Address.Hex())
        }
    }

    // Return
    return nil

}


func setUseLatestDelegateMinipools(c *cli.Context, setting bool) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get minipool statuses
    status, err := rp.MinipoolStatus()
    if err != nil {
        return err
    }
    minipools := status.Minipools

    // Get selected minipools
    var selectedMinipools []api.MinipoolDetails
    if c.String("minipool") == "" {

        // Prompt for minipool selection
        options := make([]string, len(minipools) + 1)
        options[0] = "All available minipools"
        for mi, minipool := range minipools {
            if minipool.UseLatestDelegate {
                options[mi + 1] = fmt.Sprintf("%s (auto-upgrade on)", minipool.Address.Hex())
            } else {
                options[mi + 1] = fmt.Sprintf("%s (auto-upgrade off)", minipool.Address.Hex())
            }
        }
        selected, _ := cliutils.Select("Please select a minipool to change the auto-upgrade setting for:", options)

        // Get minipools
        if selected == 0 {
            selectedMinipools = minipools
        } else {
            selectedMinipools = []api.MinipoolDetails{minipools[selected - 1]}
        }

    } else {

        // Get matching minipools
        if c.String("minipool") == "all" {
            selectedMinipools = minipools
        } else {
            selectedAddress := common.HexToAddress(c.String("minipool"))
            for _, minipool := range minipools {
                if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
                    selectedMinipools = []api.MinipoolDetails{minipool}
                    break
                }
            }
            if selectedMinipools == nil {
                return fmt.Errorf("The minipool %s is not available to modify.", selectedAddress.Hex())
            }
        }

    }

    // Get the total gas limit estimate
    var totalGas uint64 = 0
    var gasInfo rocketpoolapi.GasInfo
    for _, minipool := range selectedMinipools {
        canResponse, err := rp.CanSetUseLatestDelegateMinipool(minipool.Address, setting)
        if err != nil {
            fmt.Printf("WARNING: Couldn't get gas price for auto-upgrade setting transaction (%s)", err)
            break
        } else {
            gasInfo = canResponse.GasInfo
            totalGas += canResponse.GasInfo.EstGasLimit
        }
    }
    gasInfo.EstGasLimit = totalGas

    // Display gas estimate
    rp.PrintGasInfo(gasInfo)

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to change the auto-upgrade setting for %d minipools to %t?", len(selectedMinipools), setting))) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Update minipools
    for _, minipool := range selectedMinipools {
        response, err := rp.SetUseLatestDelegateMinipool(minipool.Address, setting)
        if err != nil {
            fmt.Printf("Could not update the auto-upgrade setting for minipool %s: %s.\n", minipool.Address.Hex(), err)
            continue
        }
    
        fmt.Printf("Updating the auto-upgrade setting for minipool %s...\n", minipool.Address.Hex())
        cliutils.PrintTransactionHash(rp, response.TxHash)
        if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
            fmt.Printf("Could not update the auto-upgrade setting for minipool %s: %s.\n", minipool.Address.Hex(), err)
        } else {
            fmt.Printf("Successfully updated the setting for minipool %s.\n", minipool.Address.Hex())
        }
    }

    // Return
    return nil

}

