package queue

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/api/queue"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test queue process method
func TestQueueProcess(t *testing.T) {

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
        LoadContracts: []string{"rocketDepositQueue", "rocketDepositSettings", "rocketMinipoolSettings", "rocketNode"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Check deposit queue for invalid staking duration cannot be processed
    if canProcess, err := queue.CanProcessQueue(p, "beer"); err != nil {
        t.Error(err)
    } else if canProcess.Success || !canProcess.InvalidStakingDuration {
        t.Error("InvalidStakingDuration flag was not set with an invalid staking duration")
    }

    // Check deposit queue with insufficent balance cannot be processed
    if canProcess, err := queue.CanProcessQueue(p, "12m"); err != nil {
        t.Error(err)
    } else if canProcess.Success || !canProcess.InsufficientBalance {
        t.Error("InsufficientBalance flag was not set with insufficient queue balance")
    }

    // Make deposit
    if _, accessorAddress, err := testapp.AppCreateGroupAccessor(appOptions); err != nil {
        t.Fatal(err)
    } else if err := testapp.AppDeposit(appOptions, "12m", eth.EthToWei(12), accessorAddress); err != nil {
        t.Fatal(err)
    }

    // Check deposit queue with no available nodes cannot be processed
    if canProcess, err := queue.CanProcessQueue(p, "12m"); err != nil {
        t.Error(err)
    } else if canProcess.Success || !canProcess.NoAvailableNodes {
        t.Error("NoAvailableNodes flag was not set with no available nodes")
    }

    // Create minipool
    if _, err := testapp.AppCreateNodeMinipools(appOptions, "12m", 1); err != nil { t.Fatal(err) }

    // Check deposit can be processed
    if canProcess, err := queue.CanProcessQueue(p, "12m"); err != nil {
        t.Error(err)
    } else if !canProcess.Success {
        t.Error("Deposit queue cannot be processed")
    }

    // Process deposit queue
    if processed, err := queue.ProcessQueue(p, "12m"); err != nil {
        t.Error(err)
    } else if !processed.Success {
        t.Error("Deposit queue was not processed successfully")
    }

}

