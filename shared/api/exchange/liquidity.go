package exchange

import (
    "errors"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Get token liquidity response type
type GetTokenLiquidityResponse struct {
    ExchangeTokenBalanceWei *big.Int    `json:"exchangeTokenBalanceWei"`
}


// Get token liquidity
func GetTokenLiquidity(p *services.Provider, token string) (*GetTokenLiquidityResponse, error) {

    // Response
    response := &GetTokenLiquidityResponse{}

    // Get token properties
    var tokenName string
    var tokenContract string
    var tokenExchangeAddress *common.Address
    switch token {
        case "RPL":
            tokenName = "RPL"
            tokenContract = "rocketPoolToken"
            tokenExchangeAddress = p.RPLExchangeAddress
    }

    // Get exchange token balance
    exchangeTokenBalanceWei := new(*big.Int)
    if err := p.CM.Contracts[tokenContract].Call(nil, exchangeTokenBalanceWei, "balanceOf", tokenExchangeAddress); err != nil {
        return nil, errors.New(fmt.Sprintf("Error retrieving %s exchange balance: " + err.Error(), tokenName))
    } else {
        response.ExchangeTokenBalanceWei = *exchangeTokenBalanceWei
    }

    // Return response
    return response, nil

}

