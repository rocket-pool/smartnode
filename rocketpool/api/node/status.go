package node

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/accounts"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/node"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Get the node's status
func getNodeStatus(c *cli.Context) error {

    // Initialise account manager
    am := accounts.NewAccountManager(c.GlobalString("keychain"))

    // Check if node account is initialised
    if !am.NodeAccountExists() {
        fmt.Println("Node account has not been initialized")
        return nil
    }

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

    // Load Rocket Pool contracts
    err = rp.LoadContracts([]string{"rocketNodeAPI", "rocketPoolToken"})
    if err != nil {
        return err
    }
    err = rp.LoadABIs([]string{"rocketNodeContract"})
    if err != nil {
        return err
    }

    // Get node account ether balance
    nodeAccountEtherBalanceWei, err := client.BalanceAt(context.Background(), am.GetNodeAccount().Address, nil)
    if err != nil {
        return errors.New("Error retrieving node account ether balance: " + err.Error())
    }

    // Get node account RPL balance
    nodeAccountRplBalanceWei := new(*big.Int)
    err = rp.Contracts["rocketPoolToken"].Call(nil, nodeAccountRplBalanceWei, "balanceOf", am.GetNodeAccount().Address)
    if err != nil {
        return errors.New("Error retrieving node account RPL balance: " + err.Error())
    }

    // Log
    fmt.Println(fmt.Sprintf(
        "Node account %s has a balance of %.2f ETH and %.2f RPL",
        am.GetNodeAccount().Address.Hex(),
        eth.WeiToEth(nodeAccountEtherBalanceWei),
        eth.WeiToEth(*nodeAccountRplBalanceWei)))

    // Check if node is registered & get node contract address
    nodeContractAddress := new(common.Address)
    err = rp.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", am.GetNodeAccount().Address)
    if err != nil {
        return errors.New("Error checking node registration: " + err.Error())
    }
    if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        fmt.Println("Node is not registered with Rocket Pool")
        return nil
    }

    // Initialise node contract
    nodeContract, err := rp.NewContract(nodeContractAddress, "rocketNodeContract")
    if err != nil {
        return errors.New("Error initialising node contract: " + err.Error())
    }

    // Get node timezone
    nodeTimezone := new(string)
    err = rp.Contracts["rocketNodeAPI"].Call(nil, nodeTimezone, "getTimezoneLocation", am.GetNodeAccount().Address)
    if err != nil {
        return errors.New("Error retrieving node timezone: " + err.Error())
    }

    // Get node contract balances
    balances, err := node.GetBalances(nodeContract)
    if err != nil {
        return err
    }

    // Log & return
    fmt.Println(fmt.Sprintf(
        "Node registered with Rocket Pool with contract at %s, timezone %s and a balance of %.2f ETH and %.2f RPL",
        nodeContractAddress.Hex(),
        *nodeTimezone,
        eth.WeiToEth(balances.EtherWei),
        eth.WeiToEth(balances.RplWei)))
    return nil

}

