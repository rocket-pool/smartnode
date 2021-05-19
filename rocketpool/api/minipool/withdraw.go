package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)


func canWithdrawMinipool(c *cli.Context, minipoolAddress common.Address) (*api.CanWithdrawMinipoolResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanWithdrawMinipoolResponse{}

    // Create minipool
    mp, err := minipool.NewMinipool(rp, minipoolAddress)
    if err != nil {
        return nil, err
    }

    // Validate minipool owner
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }
    if err := validateMinipoolOwner(mp, nodeAccount.Address); err != nil {
        return nil, err
    }

    // Check minipool status
    status, err := mp.GetStatus(nil)
    if err != nil {
        return nil, err
    }
    response.InvalidStatus = (status != types.Withdrawable)

    // Get gas estimate
    opts, err := w.GetNodeAccountTransactor()
    if err != nil { 
        return nil, err 
    }
    gasInfo, err := mp.EstimateWithdrawGas(opts)
    if err == nil {
        response.GasInfo = gasInfo
    }

    // Update & return response
    response.CanWithdraw = !response.InvalidStatus
    return &response, nil

}


func withdrawMinipool(c *cli.Context, minipoolAddress common.Address) (*api.WithdrawMinipoolResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.WithdrawMinipoolResponse{}

    // Create minipool
    mp, err := minipool.NewMinipool(rp, minipoolAddress)
    if err != nil {
        return nil, err
    }

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Override the last pending TX if requested 
    err = eth1.CheckForNonceOverride(c, opts)
    if err != nil {
        return nil, fmt.Errorf("Error checking for nonce override: %w", err)
    }

    // Withdraw
    hash, err := mp.Withdraw(opts)
    if err != nil {
        return nil, err
    }
    response.TxHash = hash

    // Return response
    return &response, nil

}

