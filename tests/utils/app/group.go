package app

import (
    "encoding/hex"
    "errors"
    "math/big"
    "math/rand"
    "time"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
)


// RocketGroupAPI GroupAdd event
type GroupAdd struct {
    ID common.Address
    Name string
    StakingFee *big.Int
    Created *big.Int
}


// RocketGroupAPI GroupCreateDefaultAccessor event
type GroupCreateDefaultAccessor struct {
    ID common.Address
    AccessorAddress common.Address
    Created *big.Int
}


// Create a group and accessor from app options
func AppCreateGroupAccessor(options AppOptions) (common.Address, common.Address, error) {

    // Initialise ethereum client
    client, err := ethclient.Dial(options.ProviderPow)
    if err != nil { return common.Address{}, common.Address{}, err }

    // Initialise contract manager & load contracts
    cm, err := rocketpool.NewContractManager(client, options.StorageAddress)
    if err != nil { return common.Address{}, common.Address{}, err }
    if err := cm.LoadContracts([]string{"rocketGroupAPI", "rocketGroupSettings"}); err != nil { return common.Address{}, common.Address{}, err }
    if err := cm.LoadABIs([]string{"rocketGroupContract"}); err != nil { return common.Address{}, common.Address{}, err }

    // Get new group fee
    newGroupFee := new(*big.Int)
    if err := cm.Contracts["rocketGroupSettings"].Call(nil, newGroupFee, "getNewFee"); err != nil { return common.Address{}, common.Address{}, err }

    // Generate group name from random bytes
    groupNameBytes := make([]byte, 16)
    rand.Seed(time.Now().UnixNano())
    rand.Read(groupNameBytes)
    groupName := make([]byte, hex.EncodedLen(len(groupNameBytes)))
    hex.Encode(groupName, groupNameBytes)

    // Get owner account
    ownerPrivateKey, _, err := test.OwnerAccount()
    if err != nil { return common.Address{}, common.Address{}, err }

    // Create group
    txor := bind.NewKeyedTransactor(ownerPrivateKey)
    txor.Value = *newGroupFee
    txReceipt, err := eth.ExecuteContractTransaction(client, txor, cm.Addresses["rocketGroupAPI"], cm.Abis["rocketGroupAPI"], "add", string(groupName), eth.EthToWei(0))
    if err != nil { return common.Address{}, common.Address{}, err }

    // Get group ID
    groupAddEvents, err := eth.GetTransactionEvents(client, txReceipt, cm.Addresses["rocketGroupAPI"], cm.Abis["rocketGroupAPI"], "GroupAdd", GroupAdd{})
    if err != nil {
        return common.Address{}, common.Address{}, err
    } else if len(groupAddEvents) == 0 {
        return common.Address{}, common.Address{}, errors.New("Failed to retrieve GroupAdd event")
    }
    groupAddEvent := (groupAddEvents[0]).(*GroupAdd)
    groupId := groupAddEvent.ID

    // Create group accessor contract
    txor = bind.NewKeyedTransactor(ownerPrivateKey)
    txReceipt, err = eth.ExecuteContractTransaction(client, txor, cm.Addresses["rocketGroupAPI"], cm.Abis["rocketGroupAPI"], "createDefaultAccessor", groupId)
    if err != nil { return common.Address{}, common.Address{}, err }

    // Get group accessor contract address
    groupCreateAccessorEvents, err := eth.GetTransactionEvents(client, txReceipt, cm.Addresses["rocketGroupAPI"], cm.Abis["rocketGroupAPI"], "GroupCreateDefaultAccessor", GroupCreateDefaultAccessor{})
    if err != nil {
        return common.Address{}, common.Address{}, err
    } else if len(groupCreateAccessorEvents) == 0 {
        return common.Address{}, common.Address{}, errors.New("Failed to retrieve GroupCreateDefaultAccessor event")
    }
    groupCreateAccessorEvent := (groupCreateAccessorEvents[0]).(*GroupCreateDefaultAccessor)
    groupAccessorAddress := groupCreateAccessorEvent.AccessorAddress

    // Add accessor to group
    txor = bind.NewKeyedTransactor(ownerPrivateKey)
    if _, err := eth.ExecuteContractTransaction(client, txor, &groupId, cm.Abis["rocketGroupContract"], "addDepositor", groupAccessorAddress); err != nil { return common.Address{}, common.Address{}, err }
    txor = bind.NewKeyedTransactor(ownerPrivateKey)
    if _, err := eth.ExecuteContractTransaction(client, txor, &groupId, cm.Abis["rocketGroupContract"], "addWithdrawer", groupAccessorAddress); err != nil { return common.Address{}, common.Address{}, err }

    // Return
    return groupId, groupAccessorAddress, nil

}

