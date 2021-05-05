package node

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils"
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

    // Get staking contract address
    rocketNodeStakingAddress, err := rp.GetAddress("rocketNodeStaking")
    if err != nil {
        return nil, err
    }

    // Check RPL balance
    rplBalance, err := tokens.GetRPLBalance(rp, nodeAccount.Address, nil)
    if err != nil {
        return nil, err
    }
    response.InsufficientBalance = (amountWei.Cmp(rplBalance) > 0)

    // Get gas estimates
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }
    approveGasInfo, err := tokens.EstimateApproveRPLGas(rp, *rocketNodeStakingAddress, amountWei, opts)
    if err != nil {
        return nil, err
    }
    stakeGasInfo, err := node.EstimateStakeGas(rp, amountWei, opts)
    if err != nil {
        return nil, err
    }
    response.GasInfo.EstGasLimit = approveGasInfo.EstGasLimit + stakeGasInfo.EstGasLimit
    response.GasInfo.EstGasPrice = approveGasInfo.EstGasPrice

    // Update & return response
    response.CanStake = !response.InsufficientBalance
    return &response, nil

}


func approveRpl(c *cli.Context, amountWei *big.Int) (*api.NodeStakeRplApproveResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.NodeStakeRplApproveResponse{}

    // Get staking contract address
    rocketNodeStakingAddress, err := rp.GetAddress("rocketNodeStaking")
    if err != nil {
        return nil, err
    }

    // Approve RPL allowance
    if opts, err := w.GetNodeAccountTransactor(); err != nil {
        return nil, err
    } else if hash, err := tokens.ApproveRPL(rp, *rocketNodeStakingAddress, amountWei, opts); err != nil {
        return nil, err
    } else {
        response.ApproveTxHash = hash
    }

    // Return response
    return &response, nil

}


func waitForApprovalAndStakeRpl(c *cli.Context, amountWei *big.Int, hash common.Hash) (*api.NodeStakeRplStakeResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }
    
    // Wait for the RPL approval TX to successfully get mined
    _, err = utils.WaitForTransaction(rp.Client, hash)
    if err != nil {
        return nil, err
    }

    // Response
    response := api.NodeStakeRplStakeResponse{}
    
    // Stake RPL
    if opts, err := w.GetNodeAccountTransactor(); err != nil {
        return nil, err
    } else if hash, err := node.StakeRPL(rp, amountWei, opts); err != nil {
        return nil, err
    } else {
        response.StakeTxHash = hash
    }

    // Return response
    return &response, nil

}

