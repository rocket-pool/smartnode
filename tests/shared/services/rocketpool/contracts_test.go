package rocketpool

import (
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"

    test "github.com/rocket-pool/smartnode/tests/utils"
)


// Test contract manager functionality
func TestContractManager(t *testing.T) {

    // Null address
    address := common.HexToAddress("0x0000000000000000000000000000000000000000")

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Initialise contract manager with incorrect storage address
    cm, err := rocketpool.NewContractManager(client, address.Hex())
    if err != nil { t.Fatal(err) }

    // Attempt to load contract & ABI
    if err := cm.LoadContracts([]string{"rocketNodeAPI"}); err == nil {
        t.Error("Contract manager LoadContracts() method should return error with incorrect storage address")
    }
    if err := cm.LoadABIs([]string{"rocketMinipool"}); err == nil {
        t.Error("Contract manager LoadABIs() method should return error with incorrect storage address")
    }

    // Initialise contract manager with correct storage address
    cm, err = rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }

    // Load contract & ABI
    if err := cm.LoadContracts([]string{"rocketNodeAPI"}); err != nil { t.Error(err) }
    if err := cm.LoadABIs([]string{"rocketMinipool"}); err != nil { t.Error(err) }

    // Attempt to load nonexistent contract & ABI
    if err := cm.LoadContracts([]string{"rocketFoo"}); err == nil {
        t.Error("Contract manager LoadContracts() method should return error for nonexistent contracts")
    }
    if err := cm.LoadABIs([]string{"rocketBar"}); err == nil {
        t.Error("Contract manager LoadABIs() method should return error for nonexistent ABIs")
    }

    // Initialise a new contract
    if _, err := cm.NewContract(&address, "rocketMinipool"); err != nil { t.Error(err) }

    // Attempt to initialise a new contract without an ABI
    if _, err := cm.NewContract(&address, "rocketNodeContract"); err == nil {
        t.Error("Contract manager NewContract() method should return error for unloaded ABIs")
    }

}

