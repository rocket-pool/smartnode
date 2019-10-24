package node

import (
    "bytes"
    "context"
    "errors"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)

// Register the node with Rocket Pool
func registerNode(c *cli.Context, timezone string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        LoadContracts: []string{"rocketNodeAPI", "rocketNodeSettings"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Response
    response := api.NodeRegisterResponse{}

    // Get node account
    nodeAccount, _ := p.AM.GetNodeAccount()
    response.AccountAddress = nodeAccount.Address

    // Status channels
    nodeContractAddressChannel := make(chan common.Address)
    registrationsAllowedChannel := make(chan bool)
    minEtherBalanceChannel := make(chan *big.Int)
    etherBalanceChannel := make(chan *big.Int)
    errorChannel := make(chan error)

    // Check if node is already registered (contract exists)
    go (func() {
        nodeContractAddress := new(common.Address)
        if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", nodeAccount.Address); err != nil {
            errorChannel <- errors.New("Error checking node registration: " + err.Error())
        } else {
            nodeContractAddressChannel <- *nodeContractAddress
        }
    })()

    // Check node registrations are enabled
    go (func() {
        registrationsAllowed := new(bool)
        if err := p.CM.Contracts["rocketNodeSettings"].Call(nil, registrationsAllowed, "getNewAllowed"); err != nil {
            errorChannel <- errors.New("Error checking node registrations enabled status: " + err.Error())
        } else {
            registrationsAllowedChannel <- *registrationsAllowed
        }
    })()

    // Get min required node account ether balance
    go (func() {
        minNodeAccountEtherBalanceWei := new(*big.Int)
        if err := p.CM.Contracts["rocketNodeSettings"].Call(nil, minNodeAccountEtherBalanceWei, "getEtherMin"); err != nil {
            errorChannel <- errors.New("Error retrieving minimum ether requirement: " + err.Error())
        } else {
            minEtherBalanceChannel <- *minNodeAccountEtherBalanceWei
        }
    })()

    // Get node account ether balance
    go (func() {
        if nodeAccountEtherBalanceWei, err := p.Client.BalanceAt(context.Background(), nodeAccount.Address, nil); err != nil {
            errorChannel <- errors.New("Error retrieving node account balance: " + err.Error())
        } else {
            etherBalanceChannel <- nodeAccountEtherBalanceWei
        }
    })()

    // Receive status
    for received := 0; received < 4; {
        select {
            case response.ContractAddress = <-nodeContractAddressChannel:
                received++
            case response.RegistrationsEnabled = <-registrationsAllowedChannel:
                received++
            case response.MinAccountBalanceEtherWei = <-minEtherBalanceChannel:
                received++
            case response.AccountBalanceEtherWei = <-etherBalanceChannel:
                received++
            case err := <-errorChannel:
                return err
        }
    }

    // Update response
    response.AlreadyRegistered = !bytes.Equal(response.ContractAddress.Bytes(), make([]byte, common.AddressLength))
    response.InsufficientAccountBalance = (response.AccountBalanceEtherWei.Cmp(response.MinAccountBalanceEtherWei) < 0)

    // Check status
    if !response.RegistrationsEnabled || response.AlreadyRegistered || response.InsufficientAccountBalance {
        api.PrintResponse(p.Output, response)
        return nil
    }

    // Register node
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return err
    } else {
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.CM.Addresses["rocketNodeAPI"], p.CM.Abis["rocketNodeAPI"], "add", timezone); err != nil {
            return errors.New("Error registering node: " + err.Error())
        }
    }

    // Get node contract address
    nodeContractAddress := new(common.Address)
    if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", nodeAccount.Address); err != nil {
        return errors.New("Error retrieving node contract address: " + err.Error())
    } else {
        response.Success = true
        response.ContractAddress = *nodeContractAddress
    }

    // Print response & return
    api.PrintResponse(p.Output, response)
    return nil

}

