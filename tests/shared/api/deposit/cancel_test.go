package deposit

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/api/deposit"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test deposit cancel methods
func TestDepositCancel(t *testing.T) {

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
        LoadContracts: []string{"rocketNodeAPI", "rocketNodeSettings"},
        LoadAbis: []string{"rocketNodeContract"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Check deposit cannot be cancelled without existing reservation
    if canCancel, err := deposit.CanCancelDeposit(p); err != nil {
        t.Error(err)
    } else if canCancel.Success || !canCancel.ReservationDidNotExist {
        t.Error("ReservationDidNotExist flag was not set without an existing deposit reservation")
    }

    // Attempt to cancel deposit
    if _, err := deposit.CancelDeposit(p); err == nil {
        t.Error("CancelDeposit() method did not return error without an existing deposit reservation")
    }

    // Reserve deposit
    if validatorKey, err := p.KM.CreateValidatorKey(); err != nil {
        t.Fatal(err)
    } else if _, err := deposit.ReserveDeposit(p, validatorKey, "3m"); err != nil {
        t.Fatal(err)
    }

    // Check deposit can be cancelled
    if canCancel, err := deposit.CanCancelDeposit(p); err != nil {
        t.Error(err)
    } else if !canCancel.Success {
        t.Error("Deposit cannot be cancelled")
    }

    // Cancel deposit
    if cancelled, err := deposit.CancelDeposit(p); err != nil {
        t.Error(err)
    } else if !cancelled.Success {
        t.Error("Deposit was not cancelled successfully")
    }

}

