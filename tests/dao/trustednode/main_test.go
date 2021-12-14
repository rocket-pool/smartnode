package trustednode

import (
	"log"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	uc "github.com/rocket-pool/rocketpool-go/utils/client"

	"github.com/rocket-pool/rocketpool-go/tests"
	"github.com/rocket-pool/rocketpool-go/tests/testutils/accounts"
)


var (
    client *uc.EthClientProxy
    rp *rocketpool.RocketPool

    ownerAccount *accounts.Account
    trustedNodeAccount1 *accounts.Account
    trustedNodeAccount2 *accounts.Account
    trustedNodeAccount3 *accounts.Account
    trustedNodeAccount4 *accounts.Account
    nodeAccount *accounts.Account
)


func TestMain(m *testing.M) {
    var err error

    // Initialize eth client
    client = uc.NewEth1ClientProxy(0, tests.Eth1ProviderAddress)

    // Initialize contract manager
    rp, err = rocketpool.NewRocketPool(client, common.HexToAddress(tests.RocketStorageAddress))
    if err != nil { log.Fatal(err) }

    // Initialize accounts
    ownerAccount, err = accounts.GetAccount(0)
    if err != nil { log.Fatal(err) }
    trustedNodeAccount1, err = accounts.GetAccount(1)
    if err != nil { log.Fatal(err) }
    trustedNodeAccount2, err = accounts.GetAccount(2)
    if err != nil { log.Fatal(err) }
    trustedNodeAccount3, err = accounts.GetAccount(3)
    if err != nil { log.Fatal(err) }
    trustedNodeAccount4, err = accounts.GetAccount(4)
    if err != nil { log.Fatal(err) }
    nodeAccount, err = accounts.GetAccount(5)
    if err != nil { log.Fatal(err) }

    // Run tests
    os.Exit(m.Run())

}

