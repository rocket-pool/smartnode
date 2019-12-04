package exchange

import (
    "io/ioutil"
    "math/big"
    "testing"

    "github.com/rocket-pool/smartnode/shared/api/exchange"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test exchange liquidity method
func TestExchangeLiquidity(t *testing.T) {

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Create test app context & options
    c := testapp.GetAppContext(dataPath)
    appOptions := testapp.GetAppOptions(dataPath)

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        CM: true,
        RPLExchangeAddress: true,
        LoadContracts: []string{"rocketPoolToken"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Get RPL exchange liquidity with no liquidity
    if liquidity, err := exchange.GetTokenLiquidity(p, "RPL"); err != nil {
        t.Error(err)
    } else if liquidity.ExchangeTokenBalanceWei.Cmp(big.NewInt(0)) != 0 {
        t.Error("Expected liquidity to be 0")
    }

    // Add liquidity
    if err := testapp.AppAddTokenLiquidity(appOptions, "RPL", eth.EthToWei(5), eth.EthToWei(500)); err != nil { t.Fatal(err) }

    // Get RPL exchange liquidity with liquidity added
    if liquidity, err := exchange.GetTokenLiquidity(p, "RPL"); err != nil {
        t.Error(err)
    } else if liquidity.ExchangeTokenBalanceWei.Cmp(big.NewInt(0)) == 0 {
        t.Error("Expected liquidity to be greater than 0")
    }

}

