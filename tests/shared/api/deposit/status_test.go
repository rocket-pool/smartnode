package deposit

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/api/deposit"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test deposit status methods
func TestDepositStatus(t *testing.T) {

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
        Client: true,
        CM: true,
        NodeContractAddress: true,
        NodeContract: true,
        LoadContracts: []string{"rocketETHToken", "rocketNodeAPI", "rocketNodeSettings", "rocketPoolToken"},
        LoadAbis: []string{"rocketNodeContract"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Get deposit status without deposit reservation
    if status, err := deposit.GetDepositStatus(p); err != nil {
        t.Error(err)
    } else if status.ReservationExists {
        t.Error("ReservationExists flag was set without an existing deposit reservation")
    }

    // Reserve deposit
    if _, err := deposit.ReserveDeposit(p, "3m"); err != nil {
        t.Fatal(err)
    }

    // Get deposit status with deposit reservation
    if status, err := deposit.GetDepositStatus(p); err != nil {
        t.Error(err)
    } else if !status.ReservationExists {
        t.Error("ReservationExists flag was not set with an existing deposit reservation")
    }

}

