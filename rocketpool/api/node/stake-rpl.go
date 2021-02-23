package node

import (
    "math/big"

    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canNodeStakeRpl(c *cli.Context, amountWei *big.Int) (*api.CanNodeStakeRplResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanNodeStakeRplResponse{}

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }

    // Check RPL balance
    rplBalance, err := tokens.GetRPLBalance(rp, nodeAccount.Address, nil)
    if err != nil {
        return nil, err
    }
    response.InsufficientBalance = (amountWei.Cmp(rplBalance) > 0)

    // Update & return response
    response.CanStake = !response.InsufficientBalance
    return &response, nil

}


func nodeStakeRpl(c *cli.Context, amountWei *big.Int) (*api.NodeStakeRplResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.NodeStakeRplResponse{}

    // Get staking contract address
    rocketNodeStakingAddress, err := rp.GetAddress("rocketNodeStaking")
    if err != nil {
        return nil, err
    }

    // Approve RPL allowance
    if opts, err := w.GetNodeAccountTransactor(); err != nil {
        return nil, err
    } else if txReceipt, err := tokens.ApproveRPL(rp, *rocketNodeStakingAddress, amountWei, opts); err != nil {
        return nil, err
    } else {
        response.ApproveTxHash = txReceipt.TxHash
    }

    // Stake RPL
    if opts, err := w.GetNodeAccountTransactor(); err != nil {
        return nil, err
    } else if txReceipt, err := node.StakeRPL(rp, amountWei, opts); err != nil {
        return nil, err
    } else {
        response.StakeTxHash = txReceipt.TxHash
    }

    // Return response
    return &response, nil

}

