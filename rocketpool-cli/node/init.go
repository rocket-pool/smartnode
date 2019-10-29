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

    // Prompt for password
    // TODO: don't prompt if password already set
    password := cliutils.Prompt(p.Input, p.Output, "Please enter a node password (this will be saved locally and used to generate dynamic keystore passphrases):", "^.{8,}$", "Please enter a password with 8 or more characters")

    // Init node
    response, err := node.InitNode(p, password)
    if err != nil { return err }

    // Print output & return
    if response.PasswordSet {
        fmt.Fprintln(p.Output, "Node password set successfully:", password)
    } else {
        fmt.Fprintln(p.Output, "Node password already set.")
    }
    if response.AccountCreated {
        fmt.Fprintln(p.Output, "Node account created successfully:", response.AccountAddress.Hex())
    } else {
        fmt.Fprintln(p.Output, "Node account already exists:", response.AccountAddress.Hex())
    }
    if response.Success {
        fmt.Fprintln(p.Output, "Please back up your Rocket Pool data folder at ~/.rocketpool in a safe and secure location to protect your node account!")
    }
    return nil

}

