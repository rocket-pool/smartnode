package node

import (
    "errors"
    "fmt"

    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode/shared/services"
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

    // Create password if it isn't set
    if p.PM.PasswordExists() {
        fmt.Fprintln(p.Output, "Node password already set.")
    } else {
        if password, err := p.PM.CreatePassword(); err != nil {
            return errors.New("Error setting node password: " + err.Error())
        } else {
            fmt.Fprintln(p.Output, "Node password set successfully:", password)
        }
    }

    // Create node account if it doesn't exist
    if p.AM.NodeAccountExists() {
        nodeAccount, _ := p.AM.GetNodeAccount()
        fmt.Fprintln(p.Output, "Node account already exists:", nodeAccount.Address.Hex())
    } else {
        if account, err := p.AM.CreateNodeAccount(); err != nil {
            return errors.New("Error creating node account: " + err.Error())
        } else {
            fmt.Fprintln(p.Output, "Node account created successfully:", account.Address.Hex())
        }
    }

    // Return
    return nil

}

