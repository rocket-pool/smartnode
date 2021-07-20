package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)


func canDestroyMinipool(c *cli.Context, minipoolAddress common.Address) (*api.CanDestroyMinipoolResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanDestroyMinipoolResponse{}

    // Create minipool
    mp, err := minipool.NewMinipool(rp, minipoolAddress)
    if err != nil {
        return nil, err
    }

    // Get gas estimate
    opts, err := w.GetNodeAccountTransactor()
    if err != nil { 
        return nil, err 
    }
    gasInfo, err := mp.EstimateDestroyGas(opts)
    if err == nil {
        response.GasInfo = gasInfo
    }

    // Update & return response
    return &response, nil

}


func destroyMinipool(c *cli.Context, minipoolAddress common.Address) (*api.DestroyMinipoolResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.DestroyMinipoolResponse{}

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

    // Close
    hash, err := mp.Destroy(opts)
    if err != nil {
        return nil, err
    }
    response.TxHash = hash

    // Return response
    return &response, nil

}

