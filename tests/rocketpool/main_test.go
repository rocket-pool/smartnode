package rocketpool

import (
	"log"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	uc "github.com/rocket-pool/rocketpool-go/utils/client"

	"github.com/rocket-pool/rocketpool-go/tests"
)


var (
    client *uc.EthClientProxy
    rp *rocketpool.RocketPool
)


func TestMain(m *testing.M) {
    var err error

    // Initialize eth client
    client = uc.NewEth1ClientProxy(0, tests.Eth1ProviderAddress)

    // Initialize contract manager
    rp, err = rocketpool.NewRocketPool(client, common.HexToAddress(tests.RocketStorageAddress))
    if err != nil { log.Fatal(err) }

    // Run tests
    os.Exit(m.Run())

}

