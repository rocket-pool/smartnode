package wallet

import (
    "errors"
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
)


func exportWallet(c *cli.Context) error {

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Get & check wallet status
    status, err := rp.WalletStatus()
    if err != nil {
        return err
    }
    if !status.WalletInitialized {
        return errors.New("The node wallet is not initialized.")
    }

    // Export wallet
    export, err := rp.ExportWallet()
    if err != nil {
        return err
    }

    // Print wallet & return
    fmt.Println("Wallet password:")
    fmt.Println("")
    fmt.Println(export.Password)
    fmt.Println("")
    fmt.Println("Wallet file:")
    fmt.Println("============")
    fmt.Println("")
    fmt.Println(export.Wallet)
    fmt.Println("")
    fmt.Println("============")
    return nil

}

