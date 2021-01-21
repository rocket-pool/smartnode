package node

import (
    "bytes"
    "log"
    "os"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/tests"
    "github.com/rocket-pool/rocketpool-go/tests/utils/accounts"
    "github.com/rocket-pool/rocketpool-go/tests/utils/evm"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


var (
    client *ethclient.Client
    rp *rocketpool.RocketPool

    nodeAccount *accounts.Account
)


func TestMain(m *testing.M) {
    var err error

    // Initialize eth client
    client, err = ethclient.Dial(tests.Eth1ProviderAddress)
    if err != nil { log.Fatal(err) }

    // Initialize contract manager
    rp, err = rocketpool.NewRocketPool(client, common.HexToAddress(tests.RocketStorageAddress))
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

    // Get & check initial node exists status
    if exists, err := node.GetNodeExists(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if exists {
        t.Error("Node already existed before registration")
    }

    // Get & check initial node details
    if details, err := node.GetNodes(rp, nil); err != nil {
        t.Error(err)
    } else if len(details) > 0 {
        t.Error("Incorrect initial node count")
    }

    // Register node
    timezoneLocation := "Australia/Brisbane"
    if _, err := node.RegisterNode(rp, timezoneLocation, nodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check updated node details
    if details, err := node.GetNodes(rp, nil); err != nil {
        t.Error(err)
    } else if len(details) == 0 {
        t.Error("Incorrect updated node count")
    } else {
        nodeDetails := details[0]
        if !bytes.Equal(nodeDetails.Address.Bytes(), nodeAccount.Address.Bytes()) {
            t.Errorf("Incorrect node address %s", nodeDetails.Address.Hex())
        }
        if !nodeDetails.Exists {
            t.Error("Incorrect node exists status")
        }
        if nodeDetails.TimezoneLocation != timezoneLocation {
            t.Errorf("Incorrect node timezone location '%s'", nodeDetails.TimezoneLocation)
        }
    }

}


func TestSetTimezoneLocation(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register node
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Set timezone
    timezoneLocation := "Australia/Sydney"
    if _, err := node.SetTimezoneLocation(rp, timezoneLocation, nodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check node timezone location
    if nodeTimezoneLocation, err := node.GetNodeTimezoneLocation(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if nodeTimezoneLocation != timezoneLocation {
        t.Errorf("Incorrect node timezone location '%s'", nodeTimezoneLocation)
    }

}


func TestDeposit(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register node
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get initial node minipool count
    minipoolCount1, err := minipool.GetNodeMinipoolCount(rp, nodeAccount.Address, nil)
    if err != nil {
        t.Fatal(err)
    }

    // Deposit
    opts := nodeAccount.GetTransactor()
    opts.Value = eth.EthToWei(16)
    if _, err := node.Deposit(rp, 0, opts); err != nil {
        t.Fatal(err)
    }

    // Get & check updated node minipool count
    minipoolCount2, err := minipool.GetNodeMinipoolCount(rp, nodeAccount.Address, nil)
    if err != nil {
        t.Fatal(err)
    } else if minipoolCount2 != minipoolCount1 + 1 {
        t.Error("Incorrect node minipool count")
    }

}

