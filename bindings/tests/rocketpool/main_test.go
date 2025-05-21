package rocketpool

import (
	"log"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/rocketpool-go/tests"
)

var (
	client *ethclient.Client
	rp     *rocketpool.RocketPool
)

func TestMain(m *testing.M) {
	var err error

	// Initialize eth client
	client, err = ethclient.Dial(tests.Eth1ProviderAddress)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize contract manager
	rp, err = rocketpool.NewRocketPool(client, common.HexToAddress(tests.RocketStorageAddress))
	if err != nil {
		log.Fatal(err)
	}

	// Run tests
	os.Exit(m.Run())

}
