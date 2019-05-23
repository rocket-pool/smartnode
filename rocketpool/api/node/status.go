package node

import (
    "bytes"
    "errors"
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool/node"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Get the node's status
func getNodeStatus(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        LoadContracts: []string{"rocketNodeAPI", "rocketPoolToken"},
        LoadAbis: []string{"rocketNodeContract"},
    })
    if err != nil {
        return err
    }

    // Get node account balances
    accountBalances, err := node.GetAccountBalances(p.AM.GetNodeAccount().Address, p.Client, p.CM)
    if err != nil {
        return err
    }

    // Log
    fmt.Println(fmt.Sprintf(
        "Node account %s has a balance of %.2f ETH and %.2f RPL",
        p.AM.GetNodeAccount().Address.Hex(),
        eth.WeiToEth(accountBalances.EtherWei),
        eth.WeiToEth(accountBalances.RplWei)))

    // Check if node is registered & get node contract address
    nodeContractAddress := new(common.Address)
    if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", p.AM.GetNodeAccount().Address); err != nil {
        return errors.New("Error checking node registration: " + err.Error())
    } else if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        fmt.Println("Node is not registered with Rocket Pool")
        return nil
    }

    // Initialise node contract
    nodeContract, err := p.CM.NewContract(nodeContractAddress, "rocketNodeContract")
    if err != nil {
        return errors.New("Error initialising node contract: " + err.Error())
    }

    // Node details channels
    nodeActiveChannel := make(chan bool)
    nodeTimezoneChannel := make(chan string)
    nodeBalancesChannel := make(chan *node.Balances)
    errorChannel := make(chan error)

    // Get node active status
    go (func() {
        nodeActiveKey := eth.KeccakBytes(bytes.Join([][]byte{[]byte("node.active"), p.AM.GetNodeAccount().Address.Bytes()}, []byte{}))
        if nodeActive, err := p.CM.RocketStorage.GetBool(nil, nodeActiveKey); err != nil {
            errorChannel <- errors.New("Error retrieving node active status: " + err.Error())
        } else {
            nodeActiveChannel <- nodeActive
        }
    })()

    // Get node timezone
    go (func() {
        nodeTimezone := new(string)
        if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, nodeTimezone, "getTimezoneLocation", p.AM.GetNodeAccount().Address); err != nil {
            errorChannel <- errors.New("Error retrieving node timezone: " + err.Error())
        } else {
            nodeTimezoneChannel <- *nodeTimezone
        }
    })()

    // Get node contract balances
    go (func() {
        if nodeBalances, err := node.GetBalances(nodeContract); err != nil {
            errorChannel <- err
        } else {
            nodeBalancesChannel <- nodeBalances
        }
    })()

    // Receive node details
    var nodeActive bool
    var nodeTimezone string
    var nodeBalances *node.Balances
    for received := 0; received < 3; {
        select {
            case nodeActive = <-nodeActiveChannel:
                received++
            case nodeTimezone = <-nodeTimezoneChannel:
                received++
            case nodeBalances = <-nodeBalancesChannel:
                received++
            case err := <-errorChannel:
                return err
        }
    }

    // Log & return
    fmt.Println(fmt.Sprintf(
        "Node registered with Rocket Pool with contract at %s, timezone '%s' and a balance of %.2f ETH and %.2f RPL",
        nodeContractAddress.Hex(),
        nodeTimezone,
        eth.WeiToEth(nodeBalances.EtherWei),
        eth.WeiToEth(nodeBalances.RplWei)))
    if !nodeActive {
        fmt.Println("Node has been marked inactive after failing to check in, and will not receive user deposits!")
        fmt.Println("Please check smart node daemon status with `rocketpool service smartnode status`; check in manually with `rocketpool service smartnode run`")
    }
    return nil

}

