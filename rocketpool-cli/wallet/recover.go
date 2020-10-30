package wallet

import (
    "errors"
    "fmt"

    "github.com/tyler-smith/go-bip39"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/passwords"
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
        var password string
        if c.String("password") == "" {
            password = promptPassword()
        } else {
            password = c.String("password")
            if len(password) < passwords.MinPasswordLength {
                return fmt.Errorf("Password does not meet minimum length of %d characters.", passwords.MinPasswordLength)
            }
        }
        if _, err := rp.SetPassword(password); err != nil {
            return err
        }
    }

    // Prompt for mnemonic
    var mnemonic string
    if c.String("mnemonic") == "" {
        mnemonic = promptMnemonic()
    } else {
        mnemonic = c.String("mnemonic")
        if !bip39.IsMnemonicValid(mnemonic) {
            return errors.New("Invalid mnemonic phrase.")
        }
    }

    // Log
    fmt.Println("Recovering node wallet...")

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
    } else {
        fmt.Println("No validator keys were found.")
    }
    return nil

}

