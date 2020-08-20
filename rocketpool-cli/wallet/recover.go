package wallet

import (
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
        fmt.Println("The node wallet is already initialized.")
        return nil
    }

    // Set password if not set
    if !status.PasswordSet {
        password := promptPassword()
        if _, err := rp.SetPassword(password); err != nil {
            return err
        }
    }

    // Prompt for mnemonic
    mnemonic := promptMnemonic()

    // Recover wallet
    if _, err := rp.RecoverWallet(mnemonic); err != nil {
        return err
    }

    // Log & return
    fmt.Println("The node wallet was successfully recovered.")
    return nil

}

