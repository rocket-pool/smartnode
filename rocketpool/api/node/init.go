package node

import (
    "errors"
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/accounts"
)


// Initialise the node with an account
func initNode(c *cli.Context) error {

    // Initialise account manager
    am := accounts.NewAccountManager(c.GlobalString("keychain"))

    // Check if node account exists
    if am.NodeAccountExists() {
        fmt.Println("Node account already exists:", am.GetNodeAccount().Address.Hex())
        return nil
    }

    // Create node account
    account, err := am.CreateNodeAccount()
    if err != nil {
        return errors.New("Error creating node account: " + err.Error())
    }

    // Log & return
    fmt.Println("Node account created successfully:", account.Address.Hex())
    return nil

}

