package node

import (
    "errors"
    "math/big"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Withdraw from node response types
type CanWithdrawFromNodeResponse struct {

    // Status
    Success bool                    `json:"success"`

    // Failure reasons
    InsufficientNodeBalance bool    `json:"insufficientNodeBalance"`

}
type WithdrawFromNodeResponse struct {
    Success bool                    `json:"success"`
}


// Check deposit can be withdrawn from node
func CanWithdrawFromNode(p *services.Provider, amountWei *big.Int, unit string) (*CanWithdrawFromNodeResponse, error) {

    // Response
    response := &CanWithdrawFromNodeResponse{}

    // Get contract method names
    var balanceMethod string
    switch unit {
        case "ETH":
            balanceMethod = "getBalanceETH"
        case "RPL":
            balanceMethod = "getBalanceRPL"
    }

    // Check withdrawal amount is available
    balanceWei := new(*big.Int)
    if err := p.NodeContract.Call(nil, balanceWei, balanceMethod); err != nil {
        return nil, errors.New("Error retrieving node balance: " + err.Error())
    } else if amountWei.Cmp(*balanceWei) > 0 {
        response.InsufficientNodeBalance = true
    }

    // Update & return response
    response.Success = !response.InsufficientNodeBalance
    return response, nil

}


// Withdraw from node
func WithdrawFromNode(p *services.Provider, amountWei *big.Int, unit string) (*WithdrawFromNodeResponse, error) {

    // Get contract method names
    var withdrawMethod string
    switch unit {
        case "ETH":
            withdrawMethod = "withdrawEther"
        case "RPL":
            withdrawMethod = "withdrawRPL"
    }

    // Withdraw amount
    if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
        return nil, err
    } else {
        if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.NodeContractAddress, p.CM.Abis["rocketNodeContract"], withdrawMethod, amountWei); err != nil {
            return nil, errors.New("Error withdrawing from node contract: " + err.Error())
        }
    }

    // Return response
    return &WithdrawFromNodeResponse{
        Success: true,
    }, nil

}

