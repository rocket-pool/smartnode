package deposit

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/api/deposit"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test deposit reserve methods
func TestDepositReserve(t *testing.T) {

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
        KM: true,
        Client: true,
        CM: true,
        NodeContractAddress: true,
        NodeContract: true,
        LoadContracts: []string{"rocketMinipoolSettings", "rocketNodeAPI", "rocketNodeSettings"},
        LoadAbis: []string{"rocketNodeContract"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Create new validator key
    validatorKey, err := p.KM.CreateValidatorKey()
    if err != nil { t.Fatal(err) }

    // Check deposit can be reserved
    if canReserve, err := deposit.CanReserveDeposit(p, validatorKey, "3m"); err != nil {
        t.Error(err)
    } else if !canReserve.Success {
        t.Error("Deposit cannot be reserved")
    }

    // Reserve deposit
    if reserved, err := deposit.ReserveDeposit(p, validatorKey, "3m"); err != nil {
        t.Error(err)
    } else if !reserved.Success {
        t.Error("Deposit was not reserved successfully")
    }

    // Check deposit cannot be reserved with existing reservation
    if canReserve, err := deposit.CanReserveDeposit(p, validatorKey, "3m"); err != nil {
        t.Error(err)
    } else if canReserve.Success || !canReserve.HadExistingReservation {
        t.Error("HadExistingReservation flag was not set with an existing deposit reservation")
    }

    // Attempt to reserve deposit
    if _, err := deposit.ReserveDeposit(p, validatorKey, "3m"); err == nil {
        t.Error("ReserveDeposit() method did not return error with an existing deposit reservation")
    }

}

