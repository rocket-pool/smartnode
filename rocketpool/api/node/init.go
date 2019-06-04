package node

import (
    "errors"
    "fmt"

    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/accounts"
)


// Initialise the node with an account
func initNode(c *cli.Context) error {

    // Initialise account manager
    am := accounts.NewAccountManager(c.GlobalString("keychainPow"))

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

