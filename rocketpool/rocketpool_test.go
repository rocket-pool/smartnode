package rocketpool

import (
    "bytes"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/rocketpool-go/utils/test"
)


func TestGetAddress(t *testing.T) {

    // Initialize eth client
    client, err := ethclient.Dial(test.Eth1ProviderAddress)
    if err != nil { t.Fatal(err) }

    // Initialize contract manager
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

