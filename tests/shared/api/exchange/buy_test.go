package exchange

import (
    "io/ioutil"
    "testing"

    "github.com/rocket-pool/smartnode/shared/api/exchange"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    testapp "github.com/rocket-pool/smartnode/tests/utils/app"
)


// Test exchange buy method
//func TestExchangeBuy(t *testing.T) {
func ExchangeBuy(t *testing.T) {

    // Create temporary data path
    dataPath, err := ioutil.TempDir("", "")
    if err != nil { t.Fatal(err) }

    // Create test app context & options
    c := testapp.GetAppContext(dataPath)
    appOptions := testapp.GetAppOptions(dataPath)

    // Initialise node
    if err := testapp.AppInitNode(appOptions); err != nil { t.Fatal(err) }

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        RPLExchangeAddress: true,
        RPLExchange: true,
        LoadContracts: []string{"rocketPoolToken"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { t.Fatal(err) }
    defer p.Cleanup()

    // Token buy amount
    tokenAmountWei := eth.EthToWei(5000)

    // Get token price
    price, err := exchange.GetTokenPrice(p, tokenAmountWei, "RPL")
    if err != nil { t.Fatal(err) }

    // Check tokens cannot be bought with insufficient liquidity
    if canBuy, err := exchange.CanBuyTokens(p, price.MaxEtherPriceWei, tokenAmountWei, "RPL"); err != nil {
        t.Error(err)
    } else if canBuy.Success || !canBuy.InsufficientExchangeLiquidity {
        t.Error("InsufficientExchangeLiquidity flag was not set with insufficient exchange liquidity")
    }

    // Attempt to buy tokens with insufficient liquidity
    if _, err := exchange.BuyTokens(p, price.MaxEtherPriceWei, tokenAmountWei, "RPL"); err == nil {
        t.Error("BuyTokens() method did not return error with insufficient exchange liquidity")
    }

    // Add liquidity
    if err := testapp.AppAddTokenLiquidity(appOptions, "RPL", eth.EthToWei(50), tokenAmountWei); err != nil { t.Fatal(err) }

    // Get token price
    price, err = exchange.GetTokenPrice(p, tokenAmountWei, "RPL")
    if err != nil { t.Fatal(err) }

    // Check tokens cannot be bought with insufficient account balance
    if canBuy, err := exchange.CanBuyTokens(p, price.MaxEtherPriceWei, tokenAmountWei, "RPL"); err != nil {
        t.Error(err)
    } else if canBuy.Success || !canBuy.InsufficientAccountBalance {
        t.Error("InsufficientAccountBalance flag was not set with insufficient account balance")
    }

    // Attempt to buy tokens with insufficient account balance
    if _, err := exchange.BuyTokens(p, price.MaxEtherPriceWei, tokenAmountWei, "RPL"); err == nil {
        t.Error("BuyTokens() method did not return error with insufficient account balance")
    }

    // Seed node account
    if err := testapp.AppSeedNodeAccount(appOptions, eth.EthToWei(100), eth.EthToWei(0)); err != nil { t.Fatal(err) }

    // Check tokens can be bought
    if canBuy, err := exchange.CanBuyTokens(p, price.MaxEtherPriceWei, tokenAmountWei, "RPL"); err != nil {
        t.Error(err)
    } else if !canBuy.Success {
        t.Error("Tokens cannot be bought")
    }

    // Buy tokens
    if bought, err := exchange.BuyTokens(p, price.MaxEtherPriceWei, tokenAmountWei, "RPL"); err != nil {
        t.Error(err)
    } else if !bought.Success {
        t.Error("Tokens were not bought successfully")
    }

}

