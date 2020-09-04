package wallet

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
)


func initWallet(c *cli.Context) error {

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

    // Initialize wallet
    response, err := rp.InitWallet()
    if err != nil {
        return err
    }

    // Print mnemonic
    fmt.Println("Your mnemonic phrase to recover your wallet is printed below. It can be used to recover your node account and validator keys if they are lost.")
    fmt.Println("Record this phrase somewhere secure and private. Do not share it with anyone as it will give them control of your node account and validators.")
    fmt.Println("")
    fmt.Println(response.Mnemonic)
    fmt.Println("")

    // Confirm mnemonic
    confirmMnemonic(response.Mnemonic)

    // Log & return
    fmt.Println("The node wallet was successfully initialized.")
    return nil

}

