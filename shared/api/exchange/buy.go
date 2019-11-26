package exchange

import (
    "context"
    "errors"
    "fmt"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/contracts"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Config
const DEADLINE_DELAY uint64 = 3600 // 1 hour


// Buy token response types
type CanBuyTokensResponse struct {

    // Status
    Success bool                        `json:"success"`

    // Failure reasons
    InsufficientAccountBalance bool     `json:"insufficientAccountBalance"`
    InsufficientExchangeLiquidity bool  `json:"insufficientExchangeLiquidity"`

}
type BuyTokensResponse struct {
    Success bool                        `json:"success"`
}


// Check tokens can be bought
func CanBuyTokens(p *services.Provider, etherAmountWei *big.Int, tokenAmountWei *big.Int, token string) (*CanBuyTokensResponse, error) {

    // Response
    response := &CanBuyTokensResponse{}

    // Get node account
    nodeAccount, _ := p.AM.GetNodeAccount()

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

    // Check node account balance
    if etherBalanceWei, err := p.Client.BalanceAt(context.Background(), nodeAccount.Address, nil); err != nil {
        return nil, errors.New("Error retrieving node account ETH balance: " + err.Error())
    } else if etherBalanceWei.Cmp(etherAmountWei) == -1 {
        response.InsufficientAccountBalance = true
    }

    // Check exchange liquidity
    exchangeTokenBalanceWei := new(*big.Int)
    if err := p.CM.Contracts[tokenContract].Call(nil, exchangeTokenBalanceWei, "balanceOf", tokenExchangeAddress); err != nil {
        return nil, errors.New(fmt.Sprintf("Error retrieving %s exchange balance: " + err.Error(), tokenName))
    } else if (*exchangeTokenBalanceWei).Cmp(tokenAmountWei) == -1 {
        response.InsufficientExchangeLiquidity = true
    }

    // Update & return response
    response.Success = !(response.InsufficientAccountBalance || response.InsufficientExchangeLiquidity)
    return response, nil

}


// Buy tokens
func BuyTokens(p *services.Provider, etherAmountWei *big.Int, tokenAmountWei *big.Int, token string) (*BuyTokensResponse, error) {

    // Get token properties
    var tokenName string
    var tokenExchangeAddress *common.Address
    var tokenExchangeAbi abi.ABI
    switch token {
        case "RPL":
            tokenName = "RPL"
            tokenExchangeAddress = p.RPLExchangeAddress
            tokenExchangeAbi, _ = abi.JSON(strings.NewReader(contracts.UniswapExchangeABI))
    }

    // Get latest block header
    header, err := p.Client.HeaderByNumber(context.Background(), nil)
    if err != nil {
        return nil, errors.New("Error retrieving latest block header: " + err.Error())
    }

    // Get token swap deadline
    deadline := header.Time + DEADLINE_DELAY

    // Buy tokens from exchange
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return nil, err
    } else {
        txor.Value = etherAmountWei
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, tokenExchangeAddress, &tokenExchangeAbi, "ethToTokenSwapOutput", tokenAmountWei, deadline); err != nil {
            return nil, errors.New(fmt.Sprintf("Error buying %s from exchange: " + err.Error(), tokenName))
        }
    }

    // Return response
    return &BuyTokensResponse{
        Success: true,
    }, nil

}

