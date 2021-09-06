package wallet

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func getStatus(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Print what network we're on the network
    err = cliutils.PrintNetwork(rp)
    if err != nil {
        return err
    }

    // Get wallet status
    status, err := rp.WalletStatus()
    if err != nil {
        return err
    }

    // Print status & return
    if status.WalletInitialized {
        fmt.Println("The node wallet is initialized.")
        fmt.Printf("Node account: %s\n", status.AccountAddress.Hex())
    } else {
        fmt.Println("The node wallet has not been initialized.")
    }
    return nil

}

