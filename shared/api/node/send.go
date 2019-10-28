package node

import (
    "context"
    "errors"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Node send response type
type NodeSendResponse struct {
    Success bool                        `json:"success"`
    InsufficientAccountBalance bool     `json:"insufficientAccountBalance"`
}


// Send from node
func SendFromNode(p *services.Provider, toAddress common.Address, sendAmountWei *big.Int, unit string) (*NodeSendResponse, error) {

    // Response
    response := &NodeSendResponse{}

    // Get node account
    nodeAccount, _ := p.AM.GetNodeAccount()

    // Handle unit types
    switch unit {
        case "ETH":

            // Check balance
            if etherBalanceWei, err := p.Client.BalanceAt(context.Background(), nodeAccount.Address, nil); err != nil {
                return nil, errors.New("Error retrieving node account ETH balance: " + err.Error())
            } else if etherBalanceWei.Cmp(sendAmountWei) == -1 {
                response.InsufficientAccountBalance = true
                break
            }

            // Send
            if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
                return nil, err
            } else {
                if _, err := eth.SendEther(p.Client, txor, &toAddress, sendAmountWei); err != nil {
                    return nil, errors.New("Error transferring ETH to address: " + err.Error())
                } else {
                    response.Success = true
                }
            }

        case "RETH": fallthrough
        case "RPL":

            // Get token properties
            var tokenName string
            var tokenContract string
            switch unit {
                case "RETH":
                    tokenName = "rETH"
                    tokenContract = "rocketETHToken"
                case "RPL":
                    tokenName = "RPL"
                    tokenContract = "rocketPoolToken"
            }

            // Check balance
            tokenBalanceWei := new(*big.Int)
            if err := p.CM.Contracts[tokenContract].Call(nil, tokenBalanceWei, "balanceOf", nodeAccount.Address); err != nil {
                return nil, errors.New(fmt.Sprintf("Error retrieving node account %s balance: " + err.Error(), tokenName))
            } else if (*tokenBalanceWei).Cmp(sendAmountWei) == -1 {
                response.InsufficientAccountBalance = true
                break
            }

            // Send
            if txor, err := p.AM.GetNodeAccountTransactor(); err != nil {
                return nil, err
            } else {
                if _, err := eth.ExecuteContractTransaction(p.Client, txor, p.CM.Addresses[tokenContract], p.CM.Abis[tokenContract], "transfer", toAddress, sendAmountWei); err != nil {
                    return nil, errors.New(fmt.Sprintf("Error transferring %s to address: " + err.Error(), tokenName))
                } else {
                    response.Success = true
                }
            }

    }

    // Return response
    return response, nil

}

