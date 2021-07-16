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


func canProcessWithdrawalMinipool(c *cli.Context, minipoolAddress common.Address) (*api.CanProcessWithdrawalResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanProcessWithdrawalResponse{}

    // Create minipool
    mp, err := minipool.NewMinipool(rp, minipoolAddress)
    if err != nil {
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
    gasInfo, err := mp.EstimateProcessWithdrawalGas(opts)
    if err == nil {
        response.GasInfo = gasInfo
    }

    // Update & return response
    response.CanWithdraw = !response.InvalidStatus
    return &response, nil

}


func processWithdrawalMinipool(c *cli.Context, minipoolAddress common.Address) (*api.ProcessWithdrawalResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.ProcessWithdrawalResponse{}

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

    // Override the provided pending TX if requested 
    err = eth1.CheckForNonceOverride(c, opts)
    if err != nil {
        return nil, fmt.Errorf("Error checking for nonce override: %w", err)
    }

    // Withdraw
    hash, err := mp.ProcessWithdrawal(opts)
    if err != nil {
        return nil, err
    }
    response.TxHash = hash

    // Return response
    return &response, nil

}


func canProcessWithdrawalAndDestroyMinipool(c *cli.Context, minipoolAddress common.Address) (*api.CanProcessWithdrawalAndDestroyResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanProcessWithdrawalAndDestroyResponse{}

    // Create minipool
    mp, err := minipool.NewMinipool(rp, minipoolAddress)
    if err != nil {
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
    gasInfo, err := mp.EstimateProcessWithdrawalAndDestroyGas(opts)
    if err == nil {
        response.GasInfo = gasInfo
    }

    // Update & return response
    response.CanWithdraw = !response.InvalidStatus
    return &response, nil

}


func processWithdrawalAndDestroyMinipool(c *cli.Context, minipoolAddress common.Address) (*api.ProcessWithdrawalAndDestroyResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.ProcessWithdrawalAndDestroyResponse{}

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

    // Override the provided pending TX if requested 
    err = eth1.CheckForNonceOverride(c, opts)
    if err != nil {
        return nil, fmt.Errorf("Error checking for nonce override: %w", err)
    }

    // Withdraw
    hash, err := mp.ProcessWithdrawalAndDestroy(opts)
    if err != nil {
        return nil, err
    }
    response.TxHash = hash

    // Return response
    return &response, nil

}

