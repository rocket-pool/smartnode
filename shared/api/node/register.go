package node

import (
    "bytes"
    "context"
    "errors"
    "math/big"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Register node response types
type CanRegisterNodeResponse struct {

    // Status
    Success bool                        `json:"success"`

    // Failure reasons
    HadExistingContract bool            `json:"hadExistingContract"`
    RegistrationsDisabled bool          `json:"registrationsDisabled"`
    InsufficientAccountBalance bool     `json:"insufficientAccountBalance"`

    // Failure info
    ContractAddress common.Address      `json:"contractAddress"`
    AccountAddress common.Address       `json:"accountAddress"`
    MinAccountBalanceEtherWei *big.Int  `json:"minAccountBalanceEtherWei"`
    AccountBalanceEtherWei *big.Int     `json:"accountBalanceEtherWei"`

}
type RegisterNodeResponse struct {
    Success bool                        `json:"success"`
    ContractAddress common.Address      `json:"contractAddress"`
}


// Check node can be registered
func CanRegisterNode(p *services.Provider) (*CanRegisterNodeResponse, error) {

    // Response
    response := &CanRegisterNodeResponse{}

    // Get node account
    nodeAccount, _ := p.AM.GetNodeAccount()
    response.AccountAddress = nodeAccount.Address

    // Status channels
    nodeContractAddressChannel := make(chan common.Address)
    registrationsDisabledChannel := make(chan bool)
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
            registrationsDisabledChannel <- !*registrationsAllowed
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
            case response.RegistrationsDisabled = <-registrationsDisabledChannel:
                received++
            case response.MinAccountBalanceEtherWei = <-minEtherBalanceChannel:
                received++
            case response.AccountBalanceEtherWei = <-etherBalanceChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Update status
    response.HadExistingContract = !bytes.Equal(response.ContractAddress.Bytes(), make([]byte, common.AddressLength))
    response.InsufficientAccountBalance = (response.AccountBalanceEtherWei.Cmp(response.MinAccountBalanceEtherWei) < 0)

    // Update & return response
    response.Success = !(response.HadExistingContract || response.RegistrationsDisabled || response.InsufficientAccountBalance)
    return response, nil

}


// Register node
func RegisterNode(p *services.Provider, timezone string) (*RegisterNodeResponse, error) {

    // Get node account
    nodeAccount, _ := p.AM.GetNodeAccount()

    // Register node
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return nil, err
    } else {
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.CM.Addresses["rocketNodeAPI"], p.CM.Abis["rocketNodeAPI"], "add", timezone); err != nil {
            return nil, errors.New("Error registering node: " + err.Error())
        }
    }

    // Get node contract address
    nodeContractAddress := new(common.Address)
    if err := p.CM.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", nodeAccount.Address); err != nil {
        return nil, errors.New("Error retrieving node contract address: " + err.Error())
    }

    // Return response
    return &RegisterNodeResponse{
        Success: true,
        ContractAddress: *nodeContractAddress,
    }, nil

}

