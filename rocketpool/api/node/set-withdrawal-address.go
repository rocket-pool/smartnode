package node

import (
    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func setWithdrawalAddress(c *cli.Context, withdrawalAddress common.Address) (*api.SetNodeWithdrawalAddressResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.SetNodeWithdrawalAddressResponse{}

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Set withdrawal address
    txReceipt, err := node.SetWithdrawalAddress(rp, withdrawalAddress, opts)
    if err != nil {
        return nil, err
    }
    response.TxHash = txReceipt.TxHash

    // Return response
    return &response, nil

}

