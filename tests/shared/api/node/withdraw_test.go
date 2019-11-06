package node

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test node withdraw methods
func TestNodeWithdraw(t *testing.T) {

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Create test app context & options
    c := testapp.GetAppContext(dataPath)
    appOptions := testapp.GetAppOptions(dataPath)

    // Initialise & register node
    if err := testapp.AppInitNode(appOptions); err != nil { t.Fatal(err) }
    if err := testapp.AppSeedNodeAccount(appOptions, eth.EthToWei(10), nil); err != nil { t.Fatal(err) }
    if err := testapp.AppRegisterNode(appOptions); err != nil { t.Fatal(err) }

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        CM: true,
        NodeContract: true,
        LoadContracts: []string{"rocketNodeAPI"},
        LoadAbis: []string{"rocketNodeContract"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Check withdrawals cannot be made from node with no balance
    if canWithdraw, err := node.CanWithdrawFromNode(p, eth.EthToWei(5), "ETH"); err != nil {
        t.Error(err)
    } else if canWithdraw.Success || !canWithdraw.InsufficientNodeBalance {
        t.Error("InsufficientNodeBalance flag was not set with an insufficient node contract ETH balance")
    }
    if canWithdraw, err := node.CanWithdrawFromNode(p, eth.EthToWei(5), "RPL"); err != nil {
        t.Error(err)
    } else if canWithdraw.Success || !canWithdraw.InsufficientNodeBalance {
        t.Error("InsufficientNodeBalance flag was not set with an insufficient node contract RPL balance")
    }

    // Attempt to withdraw from node
    if _, err := node.WithdrawFromNode(p, eth.EthToWei(5), "ETH"); err == nil {
        t.Error("WithdrawFromNode() method did not return error with an insufficient node contract ETH balance")
    }
    if _, err := node.WithdrawFromNode(p, eth.EthToWei(5), "RPL"); err == nil {
        t.Error("WithdrawFromNode() method did not return error with an insufficient node contract RPL balance")
    }

    // Seed node contract
    if err := testapp.AppSeedNodeContract(appOptions, eth.EthToWei(10), eth.EthToWei(10)); err != nil { t.Fatal(err) }

    // Check withdrawals can be made
    if canWithdraw, err := node.CanWithdrawFromNode(p, eth.EthToWei(5), "ETH"); err != nil {
        t.Error(err)
    } else if !canWithdraw.Success {
        t.Error("Node ETH withdrawal cannot be made")
    }
    if canWithdraw, err := node.CanWithdrawFromNode(p, eth.EthToWei(5), "RPL"); err != nil {
        t.Error(err)
    } else if !canWithdraw.Success {
        t.Error("Node RPL withdrawal cannot be made")
    }

    // Withdraw from node
    if withdrew, err := node.WithdrawFromNode(p, eth.EthToWei(5), "ETH"); err != nil {
        t.Error(err)
    } else if !withdrew.Success {
        t.Error("Node ETH withdrawal was not made successfully")
    }
    if withdrew, err := node.WithdrawFromNode(p, eth.EthToWei(5), "RPL"); err != nil {
        t.Error(err)
    } else if !withdrew.Success {
        t.Error("Node RPL withdrawal was not made successfully")
    }

}

