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


// Shared command setup
func setup(c *cli.Context, loadContracts []string) (*accounts.AccountManager, *ethclient.Client, *rocketpool.ContractManager, *bind.BoundContract, string, error) {

    // Initialise account manager
    am := accounts.NewAccountManager(c.GlobalString("keychain"))

    // Check node account
    if !am.NodeAccountExists() {
        return nil, nil, nil, nil, "Node account does not exist, please initialize with `rocketpool node init`", nil
    }

    // Connect to ethereum node
    client, err := ethclient.Dial(c.GlobalString("provider"))
    if err != nil {
        return nil, nil, nil, nil, "", errors.New("Error connecting to ethereum node: " + err.Error())
    }

    // Initialise Rocket Pool contract manager
    rp, err := rocketpool.NewContractManager(client, c.GlobalString("storageAddress"))
    if err != nil {
        return nil, nil, nil, nil, "", err
    }

    // Load Rocket Pool contracts
    err = rp.LoadContracts(loadContracts)
    if err != nil {
        return nil, nil, nil, nil, "", err
    }
    err = rp.LoadABIs([]string{"rocketNodeContract"})
    if err != nil {
        return nil, nil, nil, nil, "", err
    }

    // Check node is registered & get node contract address
    nodeContractAddress := new(common.Address)
    err = rp.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", am.GetNodeAccount().Address)
    if err != nil {
        return nil, nil, nil, nil, "", errors.New("Error checking node registration: " + err.Error())
    }
    if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        return nil, nil, nil, nil, "Node is not registered with Rocket Pool, please register with `rocketpool node register`", nil
    }

    // Initialise node contract
    nodeContract, err := rp.NewContract(nodeContractAddress, "rocketNodeContract")
    if err != nil {
        return nil, nil, nil, nil, "", errors.New("Error initialising node contract: " + err.Error())
    }

    // Return
    return am, client, rp, nodeContract, "", nil

}

