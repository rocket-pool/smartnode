package node

import (
    "io/ioutil"
    "testing"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test node send methods
func TestNodeSend(t *testing.T) {

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Create test app context & options
    c := testapp.GetAppContext(dataPath)
    appOptions := testapp.GetAppOptions(dataPath)

    // Initialise node
    if err := testapp.AppInitNode(appOptions); err != nil { t.Fatal(err) }

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        LoadContracts: []string{"rocketETHToken", "rocketPoolToken"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Check tokens cannot be sent from node with no balance
    if canSend, err := node.CanSendFromNode(p, eth.EthToWei(5), "ETH"); err != nil {
        t.Error(err)
    } else if canSend.Success || !canSend.InsufficientAccountBalance {
        t.Error("InsufficientAccountBalance flag was not set with an insufficient account ETH balance")
    }
    if canSend, err := node.CanSendFromNode(p, eth.EthToWei(5), "RPL"); err != nil {
        t.Error(err)
    } else if canSend.Success || !canSend.InsufficientAccountBalance {
        t.Error("InsufficientAccountBalance flag was not set with an insufficient account RPL balance")
    }

    // Attempt to send tokens
    if _, err := node.SendFromNode(p, common.HexToAddress("0x97799ecb990b907d0ca989630423d56f3f968c9a"), eth.EthToWei(5), "ETH"); err == nil {
        t.Error("SendFromNode() method did not return error with an insufficient account ETH balance")
    }
    if _, err := node.SendFromNode(p, common.HexToAddress("0x97799ecb990b907d0ca989630423d56f3f968c9a"), eth.EthToWei(5), "RPL"); err == nil {
        t.Error("SendFromNode() method did not return error with an insufficient account RPL balance")
    }

    // Seed node account
    if err := testapp.AppSeedNodeAccount(appOptions, eth.EthToWei(10), eth.EthToWei(10)); err != nil { t.Fatal(err) }

    // Check tokens can be sent
    if canSend, err := node.CanSendFromNode(p, eth.EthToWei(5), "ETH"); err != nil {
        t.Error(err)
    } else if !canSend.Success {
        t.Error("Node account ETH tokens cannot be sent")
    }
    if canSend, err := node.CanSendFromNode(p, eth.EthToWei(5), "RPL"); err != nil {
        t.Error(err)
    } else if !canSend.Success {
        t.Error("Node account RPL tokens cannot be sent")
    }

    // Send tokens
    if sent, err := node.SendFromNode(p, common.HexToAddress("0x97799ecb990b907d0ca989630423d56f3f968c9a"), eth.EthToWei(5), "ETH"); err != nil {
        t.Error(err)
    } else if !sent.Success {
        t.Error("Node account ETH tokens were not sent successfully")
    }
    if sent, err := node.SendFromNode(p, common.HexToAddress("0x97799ecb990b907d0ca989630423d56f3f968c9a"), eth.EthToWei(5), "RPL"); err != nil {
        t.Error(err)
    } else if !sent.Success {
        t.Error("Node account RPL tokens were not sent successfully")
    }

}

