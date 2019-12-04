package queue

import (
    "io/ioutil"
    "math/big"
    "testing"

    "github.com/rocket-pool/smartnode/shared/api/queue"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test queue status method
func TestQueueStatus(t *testing.T) {

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Create test app context & options
    c := testapp.GetAppContext(dataPath)
    appOptions := testapp.GetAppOptions(dataPath)

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        CM: true,
        LoadContracts: []string{"rocketDepositQueue", "rocketDepositSettings", "rocketMinipoolSettings"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Get queue status with no deposits
    if status, err := queue.GetQueueStatus(p); err != nil {
        t.Error(err)
    } else {
        if len(status.Queues) != 3 {
            t.Error("Expected statuses for 3 queues")
        }
        for _, queue := range status.Queues {
            if queue.DurationId == "6m" {
                if queue.Balance.Cmp(big.NewInt(0)) != 0 {
                    t.Error("Expected queue balance to be 0 ETH")
                }
                if queue.Chunks != 0 {
                    t.Error("Expected queue chunk count to be 0")
                }
            }
        }
    }

    // Make deposit
    depositAmount := eth.EthToWei(12)
    if _, accessorAddress, err := testapp.AppCreateGroupAccessor(appOptions); err != nil {
        t.Fatal(err)
    } else if err := testapp.AppDeposit(appOptions, "6m", depositAmount, accessorAddress); err != nil {
        t.Fatal(err)
    }

    // Get queue status with deposits
    if status, err := queue.GetQueueStatus(p); err != nil {
        t.Error(err)
    } else {
        if len(status.Queues) != 3 {
            t.Error("Expected statuses for 3 queues")
        }
        for _, queue := range status.Queues {
            if queue.DurationId == "6m" {
                if queue.Balance.Cmp(depositAmount) != 0 {
                    t.Error("Expected queue balance to be 12 ETH")
                }
                if queue.Chunks != 3 {
                    t.Error("Expected queue chunk count to be 3")
                }
            }
        }
    }

}

