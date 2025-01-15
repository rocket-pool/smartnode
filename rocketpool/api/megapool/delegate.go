package megapool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/megapool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
)

// func canDelegateUpgrade()

// func delegateUpgrade()

func getUseLatestDelegate(c *cli.Context, megapoolAddress common.Address) (*api.MegapoolGetUseLatestDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolGetUseLatestDelegateResponse{}

	// Create megapool
	mega, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get data
	setting, err := mega.GetUseLatestDelegate(nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting use latest delegate: %w", err)
	}

	//Return response
	response.Setting = setting
	return &response, nil

}

func canSetUseLatestDelegate(c *cli.Context, megapoolAddress common.Address, setting bool) (*api.MegapoolCanSetUseLatestDelegateResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolCanSetUseLatestDelegateResponse{}

	// Create Megapool
	mega, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	gasInfo, err := mega.EstimateSetUseLatestDelegateGas(setting, opts)
	if err == nil {
		response.GasInfo = gasInfo
	}

	// Return response
	return &response, nil

}

func setUseLatestDelegate(c *cli.Context, megapoolAddress common.Address, setting bool) (*api.MegapoolSetUseLatestDelegateResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolSetUseLatestDelegateResponse{}

	// Create megapool
	mega, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
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

	// Set the new setting
	hash, err := mega.SetUseLatestDelegate(setting, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func getDelegate(c *cli.Context, megapoolAddress common.Address) (*api.MegapoolGetDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolGetDelegateResponse{}

	// Create megapool
	mega, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get data
	address, err := mega.GetDelegate(nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting delegate: %w", err)
	}
	response.Address = address

	// Return response
	return &response, nil

}

func getEffectiveDelegate(c *cli.Context, megapoolAddress common.Address) (*api.MegapoolGetEffectiveDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolGetEffectiveDelegateResponse{}

	// Create Megapool
	mega, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get data
	address, err := mega.GetEffectiveDelegate(nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting delegate: %w", err)
	}

	// Return response
	response.Address = address
	return &response, nil
}
