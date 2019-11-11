package node

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Node initialization response type
type NodeInitResponse struct {
    Password *node.InitNodePasswordResponse  `json:"password"`
    Account *node.InitNodeAccountResponse    `json:"account"`
}


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

    // Check & init password
    passwordSet := node.CanInitNodePassword(p)
    if !passwordSet.HadExistingPassword {
        passwordSet, err = node.InitNodePassword(p, password)
        if err != nil { return err }
    }

    // Check & init account
    accountSet := node.CanInitNodeAccount(p)
    if !accountSet.HadExistingAccount {
        accountSet, err = node.InitNodeAccount(p)
        if err != nil { return err }
    }

    // Print response
    api.PrintResponse(p.Output, NodeInitResponse{
        Password: passwordSet,
        Account: accountSet,
    })
    return nil

}

