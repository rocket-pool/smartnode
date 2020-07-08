package rocketpool

import (
    "fmt"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/rocketpool-go/contracts"
)


// Rocket Pool contract manager
type RocketPool struct {
    RocketStorage   *contracts.RocketStorage
    client          *ethclient.Client
}


// Create new contract manager
func NewRocketPool(client *ethclient.Client, rocketStorageAddress common.Address) (*RocketPool, error) {

    // Initialize RocketStorage contract
    rocketStorage, err := contracts.NewRocketStorage(rocketStorageAddress, client)
    if err != nil {
        return nil, fmt.Errorf("Could not initialize Rocket Pool storage contract: %w", err)
    }

    // Create and return
    return &RocketPool{
        RocketStorage: rocketStorage,
        client: client,
    }, nil

}


// Load a Rocket Pool contract address
func (rp *RocketPool) Address(contractName string) (*common.Address, error) {
    return nil, nil
}


// Load a Rocket Pool contract ABI
func (rp *RocketPool) ABI(contractName string) (*abi.ABI, error) {
    return nil, nil
}


// Load a Rocket Pool contract
func (rp *RocketPool) Get(contractName string) (*bind.BoundContract, error) {
    return nil, nil
}


// Create a Rocket Pool contract instance
func (rp *RocketPool) Make(contractName string, address common.Address) (*bind.BoundContract, error) {
    return nil, nil
}

