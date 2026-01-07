package rocketpool

import (
	"log"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"

	"github.com/rocket-pool/smartnode/bindings/tests"
	"github.com/rocket-pool/smartnode/bindings/tests/testutils"
)

var (
	client *testutils.EthClientWrapper
	rp     *rocketpool.RocketPool
)

func TestMain(m *testing.M) {
	var err error

	// Initialize eth client
	rawClient, err := ethclient.Dial(tests.Eth1ProviderAddress)
	if err != nil {
		log.Fatal(err)
	}
	client = testutils.NewEthClientWrapper(rawClient)

	// Initialize contract manager
	rp, err = rocketpool.NewRocketPool(client, common.HexToAddress(tests.RocketStorageAddress))
	if err != nil {
		log.Fatal(err)
	}

	// Run tests
	os.Exit(m.Run())

}
