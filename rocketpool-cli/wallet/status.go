package wallet

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
)


func getStatus(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get wallet status
    status, err := rp.WalletStatus()
    if err != nil {
        return err
    }

    // Print status & return
    if status.WalletInitialized {
        fmt.Println("The node wallet is initialized.")
        fmt.Printf("Node account: %s\n", status.AccountAddress.Hex())
        if len(status.ValidatorKeys) > 0 {
            fmt.Println("Validator keys:")
            for _, key := range status.ValidatorKeys {
                fmt.Println(key.Hex())
            }
        }
    } else {
        fmt.Println("The node wallet has not been initialized.")
    }
    return nil

}

