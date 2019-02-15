package deposit

import (
    "bytes"
    "errors"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/accounts"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
)


// Shared command vars
var am = new(accounts.AccountManager)
var client = new(ethclient.Client)
var cm = new(rocketpool.ContractManager)
var nodeContractAddress = new(common.Address)
var nodeContract = new(bind.BoundContract)


// Shared command setup
func setup(c *cli.Context, loadContracts []string, loadAbis []string) (string, error) {

    // Initialise account manager
    *am = *accounts.NewAccountManager(c.GlobalString("keychain"))

    // Check node account
    if !am.NodeAccountExists() {
        return "Node account does not exist, please initialize with `rocketpool node init`", nil
    }

    // Connect to ethereum node
    clientV, err := ethclient.Dial(c.GlobalString("provider"))
    if err != nil {
        return "", errors.New("Error connecting to ethereum node: " + err.Error())
    }
    *client = *clientV

    // Initialise Rocket Pool contract manager
    cmV, err := rocketpool.NewContractManager(client, c.GlobalString("storageAddress"))
    if err != nil {
        return "", err
    }
    *cm = *cmV

    // Loading channels
    successChannel := make(chan bool)
    errorChannel := make(chan error)

    // Load Rocket Pool contracts
    go (func() {
        err := cm.LoadContracts(loadContracts)
        if err != nil {
            errorChannel <- err
        } else {
            successChannel <- true
        }
    })()
    go (func() {
        err := cm.LoadABIs(loadAbis)
        if err != nil {
            errorChannel <- err
        } else {
            successChannel <- true
        }
    })()

    // Await loading
    for received := 0; received < 2; {
        select {
            case <-successChannel:
                received++
            case err := <-errorChannel:
                return "", err
        }
    }

    // Check node is registered & get node contract address
    err = cm.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", am.GetNodeAccount().Address)
    if err != nil {
        return "", errors.New("Error checking node registration: " + err.Error())
    }
    if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        return "Node is not registered with Rocket Pool, please register with `rocketpool node register`", nil
    }

    // Initialise node contract
    nodeContractV, err := cm.NewContract(nodeContractAddress, "rocketNodeContract")
    if err != nil {
        return "", errors.New("Error initialising node contract: " + err.Error())
    }
    *nodeContract = *nodeContractV

    // Return
    return "", nil

}

