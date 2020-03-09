package deposit

import (
    "io/ioutil"
    "math/big"
    "testing"

    "github.com/rocket-pool/smartnode/shared/api/deposit"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test deposit complete methods
func TestDepositComplete(t *testing.T) {

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Create test app context & options
    c := testapp.GetAppContext(dataPath)
    appOptions := testapp.GetAppOptions(dataPath)

    // Initialise & register node
    if err := testapp.AppInitNode(appOptions); err != nil { t.Fatal(err) }
    if err := testapp.AppSeedNodeAccount(appOptions, eth.EthToWei(1), nil); err != nil { t.Fatal(err) }
    if err := testapp.AppRegisterNode(appOptions); err != nil { t.Fatal(err) }

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        NodeContractAddress: true,
        NodeContract: true,
        LoadContracts: []string{"rocketDepositQueue", "rocketETHToken", "rocketMinipoolSettings", "rocketNodeAPI", "rocketNodeSettings", "rocketPool", "rocketPoolToken"},
        LoadAbis: []string{"rocketNodeContract"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Check deposit cannot be completed without existing reservation
    if canComplete, err := deposit.CanCompleteDeposit(p); err != nil {
        t.Error(err)
    } else if canComplete.Success || !canComplete.ReservationDidNotExist {
        t.Error("ReservationDidNotExist flag was not set without an existing deposit reservation")
    }

    // Attempt to complete deposit
    if _, err := deposit.CompleteDeposit(p, eth.EthToWei(0), "3m"); err == nil {
        t.Error("CompleteDeposit() method did not return error without an existing deposit reservation")
    }

    // Reserve deposit
    if _, err := deposit.ReserveDeposit(p, "3m"); err != nil {
        t.Fatal(err)
    }

    // Seed node contract with required balances
    if required, err := testapp.AppGetNodeRequiredBalances(appOptions); err != nil {
        t.Fatal(err)
    } else if err := testapp.AppSeedNodeContract(appOptions, required.EtherWei, required.RplWei); err != nil {
        t.Fatal(err)
    }

    // Check deposit can be completed
    if canComplete, err := deposit.CanCompleteDeposit(p); err != nil {
        t.Error(err)
    } else if !canComplete.Success {
        t.Error("Deposit cannot be completed")
    }

    // Complete deposit
    if completed, err := deposit.CompleteDeposit(p, eth.EthToWei(0), "3m"); err != nil {
        t.Error(err)
    } else if !completed.Success {
        t.Error("Deposit was not completed successfully")
    }

    // Reserve deposit
    if _, err := deposit.ReserveDeposit(p, "3m"); err != nil {
        t.Fatal(err)
    }

    // Check required balances
    if required, err := testapp.AppGetNodeRequiredBalances(appOptions); err != nil {
        t.Fatal(err)
    } else {
        if required.EtherWei.Cmp(big.NewInt(0)) == 0 { t.Fatal("Required ETH should be > 0") }
        if required.RplWei.Cmp(big.NewInt(0)) == 0 { t.Fatal("Required RPL should be > 0") }
    }

    // Check deposit cannot be completed with insufficient balances
    if canComplete, err := deposit.CanCompleteDeposit(p); err != nil {
        t.Error(err)
    } else if canComplete.Success || !canComplete.InsufficientNodeEtherBalance || !canComplete.InsufficientNodeRplBalance {
        t.Error("InsufficientNodeEtherBalance and InsufficientNodeRplBalance flags were not set without sufficient ETH & RPL balances")
    }

    // Attempt to complete deposit
    if _, err := deposit.CompleteDeposit(p, eth.EthToWei(0), "3m"); err == nil {
        t.Error("CompleteDeposit() method did not return error without sufficient ETH & RPL balances")
    }

}

