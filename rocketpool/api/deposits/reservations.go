package deposits

import (
    "bytes"
    "errors"
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/accounts"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
)


// Reserve a node deposit
func reserveDeposit(c *cli.Context, durationId string) error {

    // Initialise account manager
    am := accounts.NewAccountManager(c.GlobalString("keychain"))

    // Get node account
    if !am.NodeAccountExists() {
        fmt.Println("Node account does not exist, please initialize with `rocketpool node init`")
        return nil
    }
    nodeAccount := am.GetNodeAccount()

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
    err = rp.LoadContracts([]string{"rocketNodeAPI", "rocketNodeSettings"})
    if err != nil {
        return err
    }
    err = rp.LoadABIs([]string{"rocketNodeContract"})
    if err != nil {
        return err
    }

    // Check node is registered (contract exists)
    nodeContractAddress := new(common.Address)
    err = rp.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", nodeAccount.Address)
    if err != nil {
        return errors.New("Error checking node registration: " + err.Error())
    }
    if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        fmt.Println("Node is not registered with Rocket Pool, please register with `rocketpool node register`")
        return nil
    }

    // Check node deposits are enabled
    depositsAllowed := new(bool)
    err = rp.Contracts["rocketNodeSettings"].Call(nil, depositsAllowed, "getDepositAllowed")
    if err != nil {
        return errors.New("Error checking node deposits enabled status: " + err.Error())
    }
    if !*depositsAllowed {
        fmt.Println("Node deposits are currently disabled in Rocket Pool")
        return nil
    }

    // Initialise node contract
    nodeContract, err := rp.NewContract(nodeContractAddress, "rocketNodeContract")
    if err != nil {
        return errors.New("Error initialising node contract: " + err.Error())
    }

    // Check node does not have current deposit reservation
    hasReservation := new(bool)
    err = nodeContract.Call(nil, hasReservation, "getHasDepositReservation")
    if err != nil {
        return errors.New("Error retrieving deposit reservation status: " + err.Error())
    }
    if *hasReservation {
        fmt.Println("Node has an existing deposit reservation, please cancel or finalize it")
        return nil
    }

    // Log & return
    fmt.Println("")
    return nil

}

