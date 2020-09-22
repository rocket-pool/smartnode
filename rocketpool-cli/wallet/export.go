package wallet

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
)


func exportWallet(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get & check wallet status
    status, err := rp.WalletStatus()
    if err != nil {
        return err
    }
    if !status.WalletInitialized {
        fmt.Println("The node wallet is not initialized.")
        return nil
    }

    // Export wallet
    export, err := rp.ExportWallet()
    if err != nil {
        return err
    }

    // Print wallet & return
    fmt.Println("Node account private key:")
    fmt.Println("")
    fmt.Println(export.AccountPrivateKey)
    fmt.Println("")
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

