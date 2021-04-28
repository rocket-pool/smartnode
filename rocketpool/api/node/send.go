package node

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)


func canNodeSend(c *cli.Context, amountWei *big.Int, token string) (*api.CanNodeSendResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    ec, err := services.GetEthClient(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanNodeSendResponse{}

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }

    // Handle token type
    switch token {
        case "eth":

            // Check node ETH balance
            ethBalanceWei, err := ec.BalanceAt(context.Background(), nodeAccount.Address, nil)
            if err != nil {
                return nil, err
            }
            response.InsufficientBalance = (amountWei.Cmp(ethBalanceWei) > 0)

        case "rpl":

            // Check node RPL balance
            rplBalanceWei, err := tokens.GetRPLBalance(rp, nodeAccount.Address, nil)
            if err != nil {
                return nil, err
            }
            response.InsufficientBalance = (amountWei.Cmp(rplBalanceWei) > 0)

        case "fsrpl":

            // Check node fixed-supply RPL balance
            fixedSupplyRplBalanceWei, err := tokens.GetFixedSupplyRPLBalance(rp, nodeAccount.Address, nil)
            if err != nil {
                return nil, err
            }
            response.InsufficientBalance = (amountWei.Cmp(fixedSupplyRplBalanceWei) > 0)

        case "neth":

            // Check node nETH balance
            nethBalanceWei, err := tokens.GetNETHBalance(rp, nodeAccount.Address, nil)
            if err != nil {
                return nil, err
            }
            response.InsufficientBalance = (amountWei.Cmp(nethBalanceWei) > 0)

        case "reth":

            // Check node rETH balance
            rethBalanceWei, err := tokens.GetRETHBalance(rp, nodeAccount.Address, nil)
            if err != nil {
                return nil, err
            }
            response.InsufficientBalance = (amountWei.Cmp(rethBalanceWei) > 0)

    }

    // Update & return response
    response.CanSend = !response.InsufficientBalance
    return &response, nil

}


func nodeSend(c *cli.Context, amountWei *big.Int, token string, to common.Address) (*api.NodeSendResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    ec, err := services.GetEthClient(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.NodeSendResponse{}

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Handle token type
    switch token {
        case "eth":

            // Transfer ETH
            opts.Value = amountWei
            hash, err := eth.SendTransaction(ec, to, opts)
            if err != nil {
                return nil, err
            }
            response.TxHash = hash

        case "rpl":

            // Transfer RPL
            hash, err := tokens.TransferRPL(rp, to, amountWei, opts)
            if err != nil {
                return nil, err
            }
            response.TxHash = hash

        case "fsrpl":

            // Transfer fixed-supply RPL
            hash, err := tokens.TransferFixedSupplyRPL(rp, to, amountWei, opts)
            if err != nil {
                return nil, err
            }
            response.TxHash = hash

        case "neth":

            // Transfer nETH
            hash, err := tokens.TransferNETH(rp, to, amountWei, opts)
            if err != nil {
                return nil, err
            }
            response.TxHash = hash

        case "reth":

            // Transfer rETH
            hash, err := tokens.TransferRETH(rp, to, amountWei, opts)
            if err != nil {
                return nil, err
            }
            response.TxHash = hash

    }

    // Return response
    return &response, nil

}

