package exchange

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/api/exchange"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test exchange price method
func TestExchangePrice(t *testing.T) {

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Create test app context
    c := testapp.GetAppContext(dataPath)

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        CM: true,
        RPLExchange: true,
        LoadContracts: []string{"rocketPoolToken"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Get token price
    if _, err := exchange.GetTokenPrice(p, eth.EthToWei(100), "RPL"); err != nil {
        t.Error(err)
    }

}

