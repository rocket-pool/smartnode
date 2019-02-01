package node

import (
    "bytes"
    "errors"
    "log"

    "github.com/ethereum/go-ethereum/accounts/keystore"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
)


// Register the node with Rocket Pool
func registerNode(c *cli.Context) error {

    // Initialise keystore
    ks := keystore.NewKeyStore(c.GlobalString("keychain"), keystore.StandardScryptN, keystore.StandardScryptP)

    // Get node account
    if len(ks.Accounts()) == 0 {
        return errors.New("Node account does not exist, please initialize with `rocketpool node init`")
    }
    nodeAccount := ks.Accounts()[0]

    // Connect to ethereum node
    client, err := ethclient.Dial(c.GlobalString("provider"))
    if err != nil {
        return errors.New("Error connecting to ethereum node: " + err.Error())
    }

    // Initialise Rocket Pool contract manager
    rp, err := rocketpool.NewContractManager(client, c.GlobalString("storageAddress"))
    if err != nil {
        return err
    }

    // Load Rocket Pool node contracts
    err = rp.LoadContracts([]string{"rocketNodeAPI"})
    if err != nil {
        return err
    }    

    // Check if node is already registered (contract exists)
    nodeContractAddress := new(common.Address)
    err = rp.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", nodeAccount.Address)
    if err != nil {
        return errors.New("Error checking node registration: " + err.Error())
    }
    if !bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        log.Println("Node already registered with contract:", nodeContractAddress.Hex())
        return nil
    }

    // Register node
    

    // Return
    return nil

}

