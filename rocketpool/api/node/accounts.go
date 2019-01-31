package node

import (
    "errors"
    "log"

    "github.com/ethereum/go-ethereum/accounts/keystore"
    "github.com/urfave/cli"
)


// Initialise the node with an account
func initNode(c *cli.Context) error {

    // Initialise keystore
    ks := keystore.NewKeyStore(c.GlobalString("keychain"), keystore.StandardScryptN, keystore.StandardScryptP)

    // Check if node account exists
    if len(ks.Accounts()) > 0 {
        log.Println("Node account already exists:", ks.Accounts()[0].Address.Hex())
        return nil
    }

    // Create node account
    account, err := ks.NewAccount("")
    if err != nil {
        return errors.New("Error creating node account: " + err.Error())
    }

    // Log & return
    log.Println("Node account created successfully:", account.Address.Hex())
    return nil

}

