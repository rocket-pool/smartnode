package wallet

import (
    "errors"
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
)


func recoverWallet(c *cli.Context) error {

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Get & check wallet status
    status, err := rp.WalletStatus()
    if err != nil {
        return err
    }
    if status.WalletInitialized {
        return errors.New("The node wallet is already initialized.")
    }

    // Set password if not set
    if !status.PasswordSet {
        if _, err := rp.SetPassword(promptPassword()); err != nil {
            return err
        }
    }

    // Recover wallet
    if _, err := rp.RecoverWallet(promptMnemonic()); err != nil {
        return err
    }

    // Log & return
    fmt.Println("The node wallet was successfully recovered.")
    return nil

}

