package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func finaliseMinipool(c *cli.Context, minipoolAddress common.Address) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check if the minipool can be finalised
    canResponse, err := rp.CanFinaliseMinipool(minipoolAddress)
    if err != nil {
        return err
    }

    // Assign max fees
    err = services.AssignMaxFee(canResponse.GasInfo, rp)
    if err != nil{
        return err
    }

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to finalize minipool %s?", minipoolAddress.Hex()))) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Finalise the minipool
    response, err := rp.FinaliseMinipool(minipoolAddress)
    if err != nil {
        return err
    }

    fmt.Printf("Finalizing minipool %s...\n", minipoolAddress)
    cliutils.PrintTransactionHash(rp, response.TxHash)
    if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully finalized the minipool.\n")
    return nil

}

