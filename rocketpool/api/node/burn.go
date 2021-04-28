package node

import (
	"math/big"

	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)


func canNodeBurn(c *cli.Context, amountWei *big.Int, token string) (*api.CanNodeBurnResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanNodeBurnResponse{}

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }

    // Sync
    var wg errgroup.Group

    // Check node balance
    wg.Go(func() error {
        switch token {
            case "neth":

                // Check node nETH balance
                nethBalanceWei, err := tokens.GetNETHBalance(rp, nodeAccount.Address, nil)
                if err != nil {
                    return err
                }
                response.InsufficientBalance = (amountWei.Cmp(nethBalanceWei) > 0)

            case "reth":

                // Check node rETH balance
                rethBalanceWei, err := tokens.GetRETHBalance(rp, nodeAccount.Address, nil)
                if err != nil {
                    return err
                }
                response.InsufficientBalance = (amountWei.Cmp(rethBalanceWei) > 0)

        }
        return nil
    })

    // Check token contract collateral
    wg.Go(func() error {
        switch token {
            case "neth":

                // Check nETH collateral
                nethContractEthBalanceWei, err := tokens.GetNETHContractETHBalance(rp, nil)
                if err != nil {
                    return err
                }
                response.InsufficientCollateral = (amountWei.Cmp(nethContractEthBalanceWei) > 0)

            case "reth":

                // Check rETH collateral
                rethTotalCollateral, err := tokens.GetRETHTotalCollateral(rp, nil)
                if err != nil {
                    return err
                }
                response.InsufficientCollateral = (amountWei.Cmp(rethTotalCollateral) > 0)

        }
        return nil
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Update & return response
    response.CanBurn = !(response.InsufficientBalance || response.InsufficientCollateral)
    return &response, nil

}


func nodeBurn(c *cli.Context, amountWei *big.Int, token string) (*api.NodeBurnResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.NodeBurnResponse{}

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Handle token type
    switch token {
        case "neth":

            // Burn nETH
            hash, err := tokens.BurnNETH(rp, amountWei, opts)
            if err != nil {
                return nil, err
            }
            response.TxHash = hash

        case "reth":

            // Burn rETH
            hash, err := tokens.BurnRETH(rp, amountWei, opts)
            if err != nil {
                return nil, err
            }
            response.TxHash = hash

    }

    // Return response
    return &response, nil

}

