package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func destroyMinipool(c *cli.Context, minipoolAddress common.Address) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check if the minipool can be destroyed
    canResponse, err := rp.CanDestroyMinipool(minipoolAddress)
    if err != nil {
        return err
    }

    // Display gas estimate
    rp.PrintGasInfo(canResponse.GasInfo)

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to destroy minipool %s?", minipoolAddress.Hex()))) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Destroy the minipool
    response, err := rp.DestroyMinipool(minipoolAddress)
    if err != nil {
        return err
    }

    fmt.Printf("Destroying minipool %s...\n", minipoolAddress)
    cliutils.PrintTransactionHash(rp, response.TxHash)
    if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully destroyed the minipool.\n")
    return nil

}

