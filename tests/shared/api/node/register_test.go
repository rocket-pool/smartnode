package node

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test node register methods
func TestNodeRegister(t *testing.T) {

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
        LoadContracts: []string{"rocketNodeAPI", "rocketNodeSettings"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Check node cannot be registered with insufficient account balance
    if canRegister, err := node.CanRegisterNode(p); err != nil {
        t.Error(err)
    } else if canRegister.Success || !canRegister.InsufficientAccountBalance {
        t.Error("InsufficientAccountBalance flag was not set with an insufficient node account balance")
    }

    // Attempt to register node
    if _, err := node.RegisterNode(p, "foo/bar"); err == nil {
        t.Error("RegisterNode() method did not return error with an insufficient node account balance")
    }

    // Seed node account
    if err := testapp.AppSeedNodeAccount(appOptions, eth.EthToWei(10), nil); err != nil { t.Fatal(err) }

    // Check node can be registered
    if canRegister, err := node.CanRegisterNode(p); err != nil {
        t.Error(err)
    } else if !canRegister.Success {
        t.Error("Node cannot be registered")
    }

    // Register node
    if registered, err := node.RegisterNode(p, "foo/bar"); err != nil {
        t.Error(err)
    } else if !registered.Success {
        t.Error("Node was not registered successfully")
    }

    // Check node cannot be registered with existing registration
    if canRegister, err := node.CanRegisterNode(p); err != nil {
        t.Error(err)
    } else if canRegister.Success || !canRegister.HadExistingContract {
        t.Error("HadExistingContract flag was not set with an existing registration")
    }

    // Attempt to register node
    if _, err := node.RegisterNode(p, "foo/bar"); err == nil {
        t.Error("RegisterNode() method did not return error with an existing registration")
    }

}

