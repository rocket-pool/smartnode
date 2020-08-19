package wallet

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
)


func exportWallet(c *cli.Context) error {

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

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

