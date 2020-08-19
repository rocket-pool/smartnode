package wallet

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/passwords"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func initWallet(c *cli.Context) error {

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

    // Initialize wallet
    response, err := rp.InitWallet()
    if err != nil {
        return err
    }

    // Print mnemonic
    fmt.Println("Your mnemonic phrase to recover this wallet is printed below. It can be used to recover your node account and validator keys if they are lost.")
    fmt.Println("Record this phrase somewhere secure and private. Do not share it with anyone as it will give them control of your node account and validators.")
    fmt.Println("")
    fmt.Println(response.Mnemonic)
    fmt.Println("")

    // Confirm mnemonic
    _ = cliutils.Prompt("Please enter 'y' to indicate that you have recorded your mnemonic phrase.", "(?i)^(y|yes)$", "Please enter 'y'")

    // Log & return
    fmt.Println("The node wallet was successfully initialized.")
    return nil

}

