package wallet

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
)


func recoverWallet(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
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
    response, err := rp.RecoverWallet(mnemonic)
    if err != nil {
        return err
    }

    // Log & return
    fmt.Println("The node wallet was successfully recovered.")
    fmt.Printf("Node account: %s\n", response.AccountAddress.Hex())
    if len(response.ValidatorKeys) > 0 {
        fmt.Println("Validator keys:")
        for _, key := range response.ValidatorKeys {
            fmt.Println(key.Hex())
        }
    }
    return nil

}

