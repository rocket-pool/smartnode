package auction

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func createLot(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check lot can be created
    canCreate, err := rp.CanCreateLot()
    if err != nil {
        return err
    }
    if !canCreate.CanCreate {
        fmt.Println("Cannot create lot:")
        if canCreate.InsufficientBalance {
            fmt.Println("The auction contract does not have a sufficient RPL balance to create a lot.")
        }
        if canCreate.CreateLotDisabled {
            fmt.Println("Lot creation is currently disabled.")
        }
        return nil
    }

    // Create lot
    response, err := rp.CreateLot()
    if err != nil {
        return err
    }

    fmt.Printf("Creating lot...\n")
    cliutils.PrintTransactionHash(rp, response.TxHash)
    if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully created a new lot with ID %d.\n", response.LotId)
    return nil

}

