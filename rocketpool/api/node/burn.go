package node

import (
    "context"
    "math/big"

    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func runCanNodeBurn(c *cli.Context, amountWei *big.Int, token string) {
    response, err := canNodeBurn(c, amountWei, token)
    if err != nil {
        api.PrintResponse(&types.CanNodeBurnResponse{Error: err.Error()})
    } else {
        api.PrintResponse(response)
    }
}


func runNodeBurn(c *cli.Context, amountWei *big.Int, token string) {
    response, err := nodeBurn(c, amountWei, token)
    if err != nil {
        api.PrintResponse(&types.NodeBurnResponse{Error: err.Error()})
    } else {
        api.PrintResponse(response)
    }
}


func canNodeBurn(c *cli.Context, amountWei *big.Int, token string) (*types.CanNodeBurnResponse, error) {

    // Get services
    if err := services.RequireNodeAccount(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    ec, err := services.GetEthClient(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := types.CanNodeBurnResponse{}

    // Sync
    var wg errgroup.Group

    // Check node balance
    wg.Go(func() error {
        switch token {
            case "neth":

                // Check node nETH balance
                nodeAccount, _ := am.GetNodeAccount()
                nethBalanceWei, err := tokens.GetNETHBalance(rp, nodeAccount.Address)
                if err != nil {
                    return err
                }
                response.InsufficientBalance = (amountWei.Cmp(nethBalanceWei) > 0)

        }
        return nil
    })

    // Check token contract collateral
    wg.Go(func() error {
        switch token {
            case "neth":

                // Check nETH contract balance
                nethContractAddress, err := rp.GetAddress("rocketNodeETHToken")
                if err != nil {
                    return err
                }
                nethContractEthBalanceWei, err := ec.BalanceAt(context.Background(), *nethContractAddress, nil)
                if err != nil {
                    return err
                }
                response.InsufficientCollateral = (amountWei.Cmp(nethContractEthBalanceWei) > 0)

        }
        return nil
    })

    // Update & return response
    response.CanBurn = !(response.InsufficientBalance || response.InsufficientCollateral)
    return &response, nil

}


func nodeBurn(c *cli.Context, amountWei *big.Int, token string) (*types.NodeBurnResponse, error) {

    // Get services
    if err := services.RequireNodeAccount(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := types.NodeBurnResponse{}

    // Get transactor
    opts, err := am.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Handle token type
    switch token {
        case "neth":

            // Burn nETH
            txReceipt, err := tokens.BurnNETH(rp, amountWei, opts)
            if err != nil {
                return nil, err
            }
            response.TxHash = txReceipt.TxHash.Hex()

    }

    // Return response
    return &response, nil

}

