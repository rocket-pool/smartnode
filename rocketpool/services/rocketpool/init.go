package rocketpool

import (
    "errors"

    "github.com/ethereum/go-ethereum/ethclient"
)


// Initialise ethereum client & Rocket Pool contract manager
func InitClient(powProvider string, storageAddress string) (*ethclient.Client, *ContractManager, error) {

    // Connect to ethereum node
    client, err := ethclient.Dial(powProvider)
    if err != nil {
        return nil, nil, errors.New("Error connecting to ethereum node: " + err.Error())
    }

    // Initialise Rocket Pool contract manager
    contractManager, err := NewContractManager(client, storageAddress)
    if err != nil {
        return nil, nil, err
    }

    // Return
    return client, contractManager, nil

}

