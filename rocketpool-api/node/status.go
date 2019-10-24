package node

import (
    "bytes"
    "errors"

    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
    "github.com/rocket-pool/smartnode/shared/utils/api"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Get the node's status
func getNodeStatus(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        LoadContracts: []string{"rocketETHToken", "rocketNodeAPI", "rocketPoolToken"},
        LoadAbis: []string{"rocketNodeContract"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Response
    response := api.NodeStatusResponse{}

    // Get node account
    nodeAccount, _ := p.AM.GetNodeAccount()
    response.AccountAddress = nodeAccount.Address

    // Get node account balances
    if accountBalances, err := node.GetAccountBalances(nodeAccount.Address, p.Client, p.CM); err != nil {
        return err
    } else {
        response.AccountBalanceEtherWei = accountBalances.EtherWei
        response.AccountBalanceRethWei = accountBalances.RethWei
        response.AccountBalanceRplWei = accountBalances.RplWei
    }

    // Check if node is registered & get node contract address
    nodeContractAddress := new(common.Address)
    if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", nodeAccount.Address); err != nil {
        return errors.New("Error checking node registration: " + err.Error())
    } else if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        api.PrintResponse(p.Output, response)
        return nil
    } else {
        response.Registered = true
        response.ContractAddress = *nodeContractAddress
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
    nodeTrustedChannel := make(chan bool)
    errorChannel := make(chan error)

    // Get node active status
    go (func() {
        nodeActiveKey := eth.KeccakBytes(bytes.Join([][]byte{[]byte("node.active"), nodeAccount.Address.Bytes()}, []byte{}))
        if nodeActive, err := p.CM.RocketStorage.GetBool(nil, nodeActiveKey); err != nil {
            errorChannel <- errors.New("Error retrieving node active status: " + err.Error())
        } else {
            nodeActiveChannel <- nodeActive
        }
    })()

    // Get node timezone
    go (func() {
        nodeTimezone := new(string)
        if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, nodeTimezone, "getTimezoneLocation", nodeAccount.Address); err != nil {
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

    // Get node trusted status
    go (func() {
        trusted := new(bool)
        if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, trusted, "getTrusted", nodeAccount.Address); err != nil {
            errorChannel <- errors.New("Error retrieving node trusted status: " + err.Error())
        } else {
            nodeTrustedChannel <- *trusted
        }
    })()

    // Receive node details
    var nodeActive bool
    var nodeTimezone string
    var nodeBalances *node.Balances
    var nodeTrusted bool
    for received := 0; received < 4; {
        select {
            case nodeActive = <-nodeActiveChannel:
                received++
            case nodeTimezone = <-nodeTimezoneChannel:
                received++
            case nodeBalances = <-nodeBalancesChannel:
                received++
            case nodeTrusted = <-nodeTrustedChannel:
                received++
            case err := <-errorChannel:
                return err
        }
    }

    // Update response
    response.Active = nodeActive
    response.Timezone = nodeTimezone
    response.ContractBalanceEtherWei = nodeBalances.EtherWei
    response.ContractBalanceRplWei = nodeBalances.RplWei
    response.Trusted = nodeTrusted

    // Print & return
    api.PrintResponse(p.Output, response)
    return nil

}

