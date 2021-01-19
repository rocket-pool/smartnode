package rocketpool

import (
    "bytes"
    "encoding/json"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/rocketpool-go/utils/test"
)


func TestGetAddress(t *testing.T) {

    // Setup
    client, err := ethclient.Dial(test.Eth1ProviderAddress)
    if err != nil { t.Fatal(err) }
    rp, err := NewRocketPool(client, common.HexToAddress(test.RocketStorageAddress))
    if err != nil { t.Fatal(err) }

    // Get contract address
    address1, err := rp.GetAddress("rocketDepositPool")
    if err != nil {
        t.Errorf("Could not get contract address: %s", err)
    } else if bytes.Equal(address1.Bytes(), common.Address{}.Bytes()) {
        t.Error("Contract address was not found")
    }

    // Get cached contract address
    address2, err := rp.GetAddress("rocketDepositPool")
    if err != nil {
        t.Errorf("Could not get cached contract address: %s", err)
    } else if !bytes.Equal(address2.Bytes(), address1.Bytes()) {
        t.Error("Cached contract address did not match original contract address")
    }

}


func TestGetAddresses(t *testing.T) {

    // Setup
    client, err := ethclient.Dial(test.Eth1ProviderAddress)
    if err != nil { t.Fatal(err) }
    rp, err := NewRocketPool(client, common.HexToAddress(test.RocketStorageAddress))
    if err != nil { t.Fatal(err) }

    // Get contract addresses
    addresses1, err := rp.GetAddresses("rocketNodeManager", "rocketNodeDeposit")
    if err != nil {
        t.Errorf("Could not get contract addresses: %s", err)
    } else {
        for ai, address := range addresses1 {
            if bytes.Equal(address.Bytes(), common.Address{}.Bytes()) {
                t.Errorf("Contract address %d was not found", ai)
            }
        }
    }

    // Get cached contract addresses
    addresses2, err := rp.GetAddresses("rocketNodeManager", "rocketNodeDeposit")
    if err != nil {
        t.Errorf("Could not get cached contract addresses: %s", err)
    } else {
        for ai := 0; ai < len(addresses2); ai++ {
            if !bytes.Equal(addresses2[ai].Bytes(), addresses1[ai].Bytes()) {
                t.Errorf("Cached contract address %d did not match original contract address", ai)
            }
        }
    }

}


func TestGetABI(t *testing.T) {

    // Setup
    client, err := ethclient.Dial(test.Eth1ProviderAddress)
    if err != nil { t.Fatal(err) }
    rp, err := NewRocketPool(client, common.HexToAddress(test.RocketStorageAddress))
    if err != nil { t.Fatal(err) }

    // Get ABI
    abi1, err := rp.GetABI("rocketDepositPool")
    if err != nil {
        t.Errorf("Could not get contract ABI: %s", err)
    }

    // Get cached ABI
    abi2, err := rp.GetABI("rocketDepositPool")
    if err != nil {
        t.Errorf("Could not get cached contract ABI: %s", err)
    } else {
        abi2Json, err := json.Marshal(abi2)
        if err != nil { t.Fatal(err) }
        abi1Json, err := json.Marshal(abi1)
        if err != nil { t.Fatal(err) }
        if !bytes.Equal(abi2Json, abi1Json) {
            t.Error("Cached contract ABI did not match original contract ABI")
        }
    }

}


func TestGetABIs(t *testing.T) {

    // Setup
    client, err := ethclient.Dial(test.Eth1ProviderAddress)
    if err != nil { t.Fatal(err) }
    rp, err := NewRocketPool(client, common.HexToAddress(test.RocketStorageAddress))
    if err != nil { t.Fatal(err) }

    // Get ABIs
    abis1, err := rp.GetABIs("rocketNodeManager", "rocketNodeDeposit")
    if err != nil {
        t.Errorf("Could not get contract ABIs: %s", err)
    }

    // Get cached ABIs
    abis2, err := rp.GetABIs("rocketNodeManager", "rocketNodeDeposit")
    if err != nil {
        t.Errorf("Could not get cached contract ABIs: %s", err)
    } else {
        for ai := 0; ai < len(abis2); ai++ {
            abi2Json, err := json.Marshal(abis2[ai])
            if err != nil { t.Fatal(err) }
            abi1Json, err := json.Marshal(abis1[ai])
            if err != nil { t.Fatal(err) }
            if !bytes.Equal(abi2Json, abi1Json) {
                t.Errorf("Cached contract ABI %d did not match original contract ABI", ai)
            }
        }
    }

}

