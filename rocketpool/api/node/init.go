package node

import (
    "errors"
    "fmt"

    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/accounts"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/passwords"
)


// Initialise the node with a password and an account
func initNode(c *cli.Context) error {

    // Initialise password manager
    pm := passwords.NewPasswordManager(c.GlobalString("password"))

    // Create password if it isn't set
    if pm.PasswordExists() {
        fmt.Println("Node password already set.")
    } else {
        if password, err := pm.CreatePassword(); err != nil {
            return errors.New("Error setting node password: " + err.Error())
        } else {
            fmt.Println("Node password set successfully:", password)
        }
    }

    // Initialise account manager
    am := accounts.NewAccountManager(c.GlobalString("keychainPow"))

    // Create node account if it doesn't exist
    if am.NodeAccountExists() {
        fmt.Println("Node account already exists:", am.GetNodeAccount().Address.Hex())
    } else {
        if account, err := am.CreateNodeAccount(); err != nil {
            return errors.New("Error creating node account: " + err.Error())
        } else {
            fmt.Println("Node account created successfully:", account.Address.Hex())
        }
    }

    // Return
    return nil

}

