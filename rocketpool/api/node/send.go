package node

import (
    "context"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func runCanNodeSend(c *cli.Context, amountWei *big.Int, token string) {
    response, err := canNodeSend(c, amountWei, token)
    if err != nil {
        api.PrintResponse(&types.CanNodeSendResponse{Error: err.Error()})
    } else {
        api.PrintResponse(response)
    }
}


func runNodeSend(c *cli.Context, amountWei *big.Int, token string, to common.Address) {
    response, err := nodeSend(c, amountWei, token, to)
    if err != nil {
        api.PrintResponse(&types.NodeSendResponse{Error: err.Error()})
    } else {
        api.PrintResponse(response)
    }
}


func canNodeSend(c *cli.Context, amountWei *big.Int, token string) (*types.CanNodeSendResponse, error) {

    // Get services
    if err := services.RequireNodeAccount(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    ec, err := services.GetEthClient(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := types.CanNodeSendResponse{}

    // Get node account
    nodeAccount, _ := am.GetNodeAccount()

    // Handle token type
    switch token {
        case "eth":

            // Check node ETH balance
            ethBalanceWei, err := ec.BalanceAt(context.Background(), nodeAccount.Address, nil)
            if err != nil {
                return nil, err
            }
            response.InsufficientBalance = (amountWei.Cmp(ethBalanceWei) > 0)

        case "neth":

            // Check node nETH balance
            nethBalanceWei, err := tokens.GetNETHBalance(rp, nodeAccount.Address)
            if err != nil {
                return nil, err
            }
            response.InsufficientBalance = (amountWei.Cmp(nethBalanceWei) > 0)

    }

    // Update & return response
    response.CanSend = !response.InsufficientBalance
    return &response, nil

}


func nodeSend(c *cli.Context, amountWei *big.Int, token string, to common.Address) (*types.NodeSendResponse, error) {

    // Get services
    if err := services.RequireNodeAccount(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    //ec, err := services.GetEthClient(c)
    //if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := types.NodeSendResponse{}

    // Get transactor
    opts, err := am.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Handle token type
    switch token {
        case "eth":

            // Transfer ETH
            // TODO: implement

        case "neth":

            // Transfer nETH
            txReceipt, err := tokens.TransferNETH(rp, to, amountWei, opts)
            if err != nil {
                return nil, err
            }
            response.TxHash = txReceipt.TxHash.Hex()

    }

    // Return response
    return &response, nil

}

