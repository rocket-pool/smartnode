package minipool

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/api/minipool"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test minipool status methods
func TestMinipoolStatus(t *testing.T) {

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
        LoadContracts: []string{"rocketPoolToken", "utilAddressSetStorage"},
        LoadAbis: []string{"rocketMinipool"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Get minipool status with no minipools
    if status, err := minipool.GetMinipoolStatus(p); err != nil {
        t.Error(err)
    } else if len(status.Minipools) != 0 {
        t.Error("Minipools returned with no existing minipools")
    }

    // Create minipools
    if _, err := testapp.AppCreateNodeMinipools(appOptions, "3m", 3); err != nil { t.Fatal(err) }

    // Get minipool status with minipools
    if status, err := minipool.GetMinipoolStatus(p); err != nil {
        t.Error(err)
    } else if len(status.Minipools) != 3 {
        t.Error("Existing minipools not returned")
    }

}

