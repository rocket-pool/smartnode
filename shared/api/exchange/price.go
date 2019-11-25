package exchange

import (
    "errors"
    "fmt"
    "math/big"

    "github.com/rocket-pool/smartnode/shared/contracts"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Config
const MAX_PRICE_PERCENT int64 = 101 // 1% slippage


// Get token price response type
type GetTokenPriceResponse struct {
    ExpectedEtherPriceWei *big.Int  `json:"expectedEtherPriceWei"`
    MaxEtherPriceWei *big.Int       `json:"maxEtherPriceWei"`
    ExpectedTokenRate float64       `json:"expectedTokenRate"`
    MinTokenRate float64            `json:"minTokenRate"`
}


// Get token price
func GetTokenPrice(p *services.Provider, amountWei *big.Int, token string) (*GetTokenPriceResponse, error) {

    // Response
    response := &GetTokenPriceResponse{}

    // Get token properties
    var tokenName string
    var tokenExchangeContract *contracts.UniswapExchange
    switch token {
        case "RPL":
            tokenName = "RPL"
            tokenExchangeContract = p.RPLExchange
    }

    // Get expected ether price
    etherPriceWei, err := tokenExchangeContract.GetEthToTokenOutputPrice(nil, amountWei)
    if err != nil {
        return nil, errors.New(fmt.Sprintf("Error retrieving %s exchange ether price: " + err.Error(), tokenName))
    } else {
        response.ExpectedEtherPriceWei = etherPriceWei
    }

    // Get max ether price
    maxEtherPriceWei := big.NewInt(0)
    maxEtherPriceWei.Mul(etherPriceWei, big.NewInt(MAX_PRICE_PERCENT))
    maxEtherPriceWei.Div(maxEtherPriceWei, big.NewInt(100))
    response.MaxEtherPriceWei = maxEtherPriceWei

    // Get token rates
    response.ExpectedTokenRate = eth.WeiToEth(amountWei) / eth.WeiToEth(etherPriceWei)
    response.MinTokenRate = eth.WeiToEth(amountWei) / eth.WeiToEth(maxEtherPriceWei)

    // Return response
    return response, nil

}

