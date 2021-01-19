package node

import (
    "log"
    "os"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/test"
    "github.com/rocket-pool/rocketpool-go/utils/test/accounts"
    "github.com/rocket-pool/rocketpool-go/utils/test/evm"
)


var (
    client *ethclient.Client
    rp *rocketpool.RocketPool

    nodeAccount *accounts.Account
)


func TestMain(m *testing.M) {
    var err error

    // Initialize eth client
    client, err = ethclient.Dial(test.Eth1ProviderAddress)
    if err != nil { log.Fatal(err) }

    // Initialize contract manager
    rp, err = rocketpool.NewRocketPool(client, common.HexToAddress(test.RocketStorageAddress))
    if err != nil { log.Fatal(err) }

    // Initialize accounts
    nodeAccount, err = accounts.GetAccount(1)
    if err != nil { log.Fatal(err) }

    // Run tests
    os.Exit(m.Run())

}


func TestRegisterNode(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Get initial node exists status
    if exists, err := GetNodeExists(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if exists {
        t.Error("Node already existed before registration")
    }

    // Register node
    nodeTimezoneLocation := "Australia/Brisbane"
    if _, err := RegisterNode(rp, nodeTimezoneLocation, nodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get updated node exists status
    if exists, err := GetNodeExists(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if !exists {
        t.Error("Node did not exist after registration")
    }

    // Get node timezone location
    if timezoneLocation, err := GetNodeTimezoneLocation(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if timezoneLocation != nodeTimezoneLocation {
        t.Errorf("Incorrect node timezone location '%s'", timezoneLocation)
    }

}

