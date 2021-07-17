package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func processWithdrawal(c *cli.Context, minipoolAddress common.Address) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check if the minipool can be withdrawn
    canResponse, err := rp.CanProcessWithdrawalMinipool(minipoolAddress)
    if err != nil {
        return err
    }

    // Display gas estimate
    rp.PrintGasInfo(canResponse.GasInfo)

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to process a withdrawal on minipool %s?", minipoolAddress.Hex()))) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Withdraw the minipool
    response, err := rp.ProcessWithdrawalMinipool(minipoolAddress)
    if err != nil {
        return err
    }

    fmt.Printf("Withdrawing from minipool %s...\n", minipoolAddress)
    cliutils.PrintTransactionHash(rp, response.TxHash)
    if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully withdrew from the minipool.\n")
    return nil

}


func processWithdrawalAndDestroy(c *cli.Context, minipoolAddress common.Address) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check if the minipool can be withdrawn
    canResponse, err := rp.CanProcessWithdrawalAndDestroyMinipool(minipoolAddress)
    if err != nil {
        return err
    }

    // Display gas estimate
    rp.PrintGasInfo(canResponse.GasInfo)

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to process a withdrawal on minipool %s, and destroy it afterwards? This action cannot be undone!", minipoolAddress.Hex()))) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Withdraw the minipool
    response, err := rp.ProcessWithdrawalAndDestroyMinipool(minipoolAddress)
    if err != nil {
        return err
    }

    fmt.Printf("Withdrawing from minipool %s and destroying it...\n", minipoolAddress)
    cliutils.PrintTransactionHash(rp, response.TxHash)
    if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully withdrew from the minipool and destroyed it.\n")
    return nil

}

