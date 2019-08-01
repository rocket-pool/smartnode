package rocketpool

import (
    "errors"
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode/shared/services/accounts"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
)


// RocketNodeAPI NodeAdd event
type NodeAdd struct {
    ID common.Address
    ContractAddress common.Address
    Created *big.Int
}


// Register a node
func RegisterNode(client *ethclient.Client, cm *rocketpool.ContractManager, am *accounts.AccountManager) (*bind.BoundContract, common.Address, error) {

    // Seed node account
    account, err := am.GetNodeAccount()
    if err != nil { return nil, common.Address{}, err }
    if err := test.SeedAccount(client, account.Address, eth.EthToWei(10)); err != nil { return nil, common.Address{}, err }

    // Register node
    txor, err := am.GetNodeAccountTransactor()
    if err != nil { return nil, common.Address{}, err }
    txReceipt, err := eth.ExecuteContractTransaction(client, txor, cm.Addresses["rocketNodeAPI"], cm.Abis["rocketNodeAPI"], "add", "Australia/Brisbane")
    if err != nil { return nil, common.Address{}, err }

    // Get NodeAdd event
    nodeAddEvents, err := eth.GetTransactionEvents(client, txReceipt, cm.Addresses["rocketNodeAPI"], cm.Abis["rocketNodeAPI"], "NodeAdd", NodeAdd{})
    if err != nil {
        return nil, common.Address{}, err
    } else if len(nodeAddEvents) == 0 {
        return nil, common.Address{}, errors.New("Failed to retrieve NodeAdd event")
    }
    nodeAddEvent := (nodeAddEvents[0]).(*NodeAdd)

    // Create and return node contract
    nodeContract, err := cm.NewContract(&nodeAddEvent.ContractAddress, "rocketNodeContract")
    if err != nil { return nil, common.Address{}, err }

    // Return
    return nodeContract, nodeAddEvent.ContractAddress, nil

}

