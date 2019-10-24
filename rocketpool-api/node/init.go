package node

import (
    "errors"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Initialise the node with a password and an account
func initNode(c *cli.Context, password string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        PM: true,
        AM: true,
        PasswordOptional: true,
        NodeAccountOptional: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Response
    response := api.NodeInitResponse{Success: true}

    // Create password if it isn't set
    if !p.PM.PasswordExists() {
        if err := p.PM.SetPassword(password); err != nil {
            return errors.New("Error setting node password: " + err.Error())
        } else {
            response.PasswordSet = true
        }
    }

    // Create node account if it doesn't exist
    if p.AM.NodeAccountExists() {
        nodeAccount, _ := p.AM.GetNodeAccount()
        response.AccountAddress = nodeAccount.Address
    } else {
        if account, err := p.AM.CreateNodeAccount(); err != nil {
            return errors.New("Error creating node account: " + err.Error())
        } else {
            response.AccountCreated = true
            response.AccountAddress = account.Address
        }
    }

    // Print response & return
    api.PrintResponse(p.Output, response)
    return nil

}

