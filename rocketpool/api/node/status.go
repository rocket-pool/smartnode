package node

import (
    "bytes"
    "errors"
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/node"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Get the node's status
func getNodeStatus(c *cli.Context) error {

    // Command setup
    am, client, rp, message, err := setup(c, []string{"rocketNodeAPI", "rocketPoolToken"}, []string{"rocketNodeContract"}, false)
    if message != "" {
        fmt.Println(message)
        return nil
    }
    if err != nil {
        return err
    }

    // Get node account balances
    accountBalances, err := node.GetAccountBalances(am.GetNodeAccount().Address, client, rp)
    if err != nil {
        return err
    }

    // Log
    fmt.Println(fmt.Sprintf(
        "Node account %s has a balance of %.2f ETH and %.2f RPL",
        am.GetNodeAccount().Address.Hex(),
        eth.WeiToEth(accountBalances.EtherWei),
        eth.WeiToEth(accountBalances.RplWei)))

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

