package wallet

import (
    "fmt"

    "github.com/tyler-smith/go-bip39"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/passwords"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
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

        // Prompt for password and confirm
        var password string
        for confirmed := false; !confirmed; {
            password = cliutils.Prompt(
                "Please enter a password to secure your wallet with:",
                fmt.Sprintf("^.{%d,}$", passwords.MinPasswordLength),
                fmt.Sprintf("Your password must be at least %d characters long", passwords.MinPasswordLength),
            )
            confirmation := cliutils.Prompt(
                "Please confirm your password:",
                fmt.Sprintf("^.{%d,}$", passwords.MinPasswordLength),
                fmt.Sprintf("Your password must be at least %d characters long", passwords.MinPasswordLength),
            )
            if password == confirmation {
                confirmed = true
            } else {
                fmt.Println("Password confirmation does not match.")
            }
        }

        // Set password
        if _, err := rp.SetPassword(password); err != nil {
            return err
        }

    }

    // Prompt for mnemonic
    var mnemonic string
    for valid := false; !valid; {
        mnemonic = cliutils.Prompt("Please enter your recovery mnemonic phrase:", "^\\w+(\\s\\w+)*$", "Invalid mnemonic phrase")
        if bip39.IsMnemonicValid(mnemonic) {
            valid = true
        } else {
            fmt.Println("Invalid mnemonic phrase.")
        }
    }

    // Recover wallet
    if _, err := rp.RecoverWallet(mnemonic); err != nil {
        return err
    }

    // Log & return
    fmt.Println("The node wallet was successfully recovered.")
    return nil

}

