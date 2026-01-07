package protocol

import (
	"log"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"

	"github.com/rocket-pool/smartnode/bindings/tests"
	"github.com/rocket-pool/smartnode/bindings/tests/testutils"
	"github.com/rocket-pool/smartnode/bindings/tests/testutils/accounts"
)

var (
	client *testutils.EthClientWrapper
	rp     *rocketpool.RocketPool

	ownerAccount *accounts.Account
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

	// Initialize accounts
	ownerAccount, err = accounts.GetAccount(0)
	if err != nil {
		log.Fatal(err)
	}

	// Run tests
	os.Exit(m.Run())

}
