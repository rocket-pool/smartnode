package node

import (
    "bytes"
    "errors"
    "math/big"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Node status response type
type NodeStatusResponse struct {
    AccountAddress common.Address       `json:"accountAddress"`
    AccountBalanceEtherWei *big.Int     `json:"accountBalanceEtherWei"`
    AccountBalanceRethWei *big.Int      `json:"accountBalanceRethWei"`
    AccountBalanceRplWei *big.Int       `json:"accountBalanceRplWei"`
    ContractAddress common.Address      `json:"contractAddress"`
    ContractBalanceEtherWei *big.Int    `json:"contractBalanceEtherWei"`
    ContractBalanceRplWei *big.Int      `json:"contractBalanceRplWei"`
    Registered bool                     `json:"registered"`
    Active bool                         `json:"active"`
    Trusted bool                        `json:"trusted"`
    Timezone string                     `json:"timezone"`
}


// Get node status
func GetNodeStatus(p *services.Provider) (*NodeStatusResponse, error) {

    // Response
    response := &NodeStatusResponse{}

    // Get node account
    nodeAccount, _ := p.AM.GetNodeAccount()
    response.AccountAddress = nodeAccount.Address

    // Get node account balances
    if accountBalances, err := node.GetAccountBalances(nodeAccount.Address, p.Client, p.CM); err != nil {
        return nil, err
    } else {
        response.AccountBalanceEtherWei = accountBalances.EtherWei
        response.AccountBalanceRethWei = accountBalances.RethWei
        response.AccountBalanceRplWei = accountBalances.RplWei
    }

    // Check if node is registered & get node contract address
    nodeContractAddress := new(common.Address)
    if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", nodeAccount.Address); err != nil {
        return nil, errors.New("Error checking node registration: " + err.Error())
    } else if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        return response, nil
    } else {
        response.Registered = true
        response.ContractAddress = *nodeContractAddress
    }

    // Initialise node contract
    nodeContract, err := p.CM.NewContract(nodeContractAddress, "rocketNodeContract")
    if err != nil {
        return nil, errors.New("Error initialising node contract: " + err.Error())
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
    for received := 0; received < 4; {
        select {
            case response.Active = <-nodeActiveChannel:
                received++
            case response.Timezone = <-nodeTimezoneChannel:
                received++
            case nodeBalances := <-nodeBalancesChannel:
                response.ContractBalanceEtherWei = nodeBalances.EtherWei
                response.ContractBalanceRplWei = nodeBalances.RplWei
                received++
            case response.Trusted = <-nodeTrustedChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return response
    return response, nil

}

