package node

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test node status methods
func TestNodeStatus(t *testing.T) {

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
        LoadContracts: []string{"rocketETHToken", "rocketNodeAPI", "rocketPoolToken"},
        LoadAbis: []string{"rocketNodeContract"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Get unregistered node status
    if status, err := node.GetNodeStatus(p); err != nil {
        t.Error(err)
    } else if status.Registered {
        t.Error("Registered flag was set for unregistered node")
    }

    // Seed node account & register node
    if err := testapp.AppSeedNodeAccount(appOptions, eth.EthToWei(10), nil); err != nil { t.Fatal(err) }
    if err := testapp.AppRegisterNode(appOptions); err != nil { t.Fatal(err) }

    // Get registered node status
    if status, err := node.GetNodeStatus(p); err != nil {
        t.Error(err)
    } else {
        if !status.Registered { t.Error("Registered flag was not set for registered node") }
        if !status.Active { t.Error("Active flag was not set for active node") }
        if status.Trusted { t.Error("Trusted flag was set for untrusted node") }
    }

    // Make node trusted
    if err := testapp.AppSetNodeTrusted(appOptions); err != nil { t.Fatal(err) }

    // Get trusted node status
    if status, err := node.GetNodeStatus(p); err != nil {
        t.Error(err)
    } else if !status.Trusted {
        t.Error("Trusted flag was not set for trusted node")
    }

}

