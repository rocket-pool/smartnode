package node

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Initialise the node with a password and an account
func initNode(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        PM: true,
        AM: true,
        PasswordOptional: true,
        NodeAccountOptional: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Check & init password
    passwordSet := node.CanInitNodePassword(p)
    if passwordSet.HadExistingPassword {
        fmt.Fprintln(p.Output, "Node password already set.")
    } else {

        // Prompt for password
        password := cliutils.Prompt(p.Input, p.Output, "Please enter a node password (this will be saved locally and used to generate dynamic keystore passphrases):", "^.{8,}$", "Please enter a password with 8 or more characters")

        // Init password
        passwordSet, err = node.InitNodePassword(p, password)
        if err != nil { return err }

        // Print output
        fmt.Fprintln(p.Output, "Node password set successfully:", password)

    }

    // Check & init account
    accountSet := node.CanInitNodeAccount(p)
    if accountSet.HadExistingAccount {
        fmt.Fprintln(p.Output, "Node account already exists:", accountSet.AccountAddress.Hex())
    } else {

        // Init account
        accountSet, err = node.InitNodeAccount(p)
        if err != nil { return err }

        // Print output
        fmt.Fprintln(p.Output, "Node account created successfully:", accountSet.AccountAddress.Hex())

    }

    // Print backup notice & return
    if passwordSet.Success || accountSet.Success {
        fmt.Fprintln(p.Output, "Please back up your Rocket Pool data folder at ~/.rocketpool in a safe and secure location to protect your node account!")
    }
    return nil

}

