package auction

import (
	"log"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/rocketpool-go/tests"
	"github.com/rocket-pool/rocketpool-go/tests/testutils/accounts"
	"github.com/rocket-pool/rocketpool-go/tests/utils"
)

var (
	client *ethclient.Client
	rp     *rocketpool.RocketPool

	ownerAccount        *accounts.Account
	trustedNodeAccount1 *accounts.Account
	trustedNodeAccount2 *accounts.Account
	trustedNodeAccount3 *accounts.Account
	userAccount1        *accounts.Account
	userAccount2        *accounts.Account
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

	// Initialize accounts
	ownerAccount, err = accounts.GetAccount(0)
	if err != nil {
		log.Fatal(err)
	}
	trustedNodeAccount1, err = accounts.GetAccount(1)
	if err != nil {
		log.Fatal(err)
	}
	trustedNodeAccount2, err = accounts.GetAccount(2)
	if err != nil {
		log.Fatal(err)
	}
	trustedNodeAccount3, err = accounts.GetAccount(3)
	if err != nil {
		log.Fatal(err)
	}
	userAccount1, err = accounts.GetAccount(8)
	if err != nil {
		log.Fatal(err)
	}
	userAccount2, err = accounts.GetAccount(9)
	if err != nil {
		log.Fatal(err)
	}

	// Do the bootstrap settings
	utils.Stage4Bootstrap(rp, ownerAccount)

	// Run tests
	os.Exit(m.Run())

}
