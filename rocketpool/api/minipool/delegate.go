package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canDelegateUpgrade(c *cli.Command, minipoolAddress common.Address) (*api.CanDelegateUpgradeResponse, error) {

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
	response := api.CanDelegateUpgradeResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get latest delegate address
	latestDelegateAddress, err := rp.GetAddress("rocketMinipoolDelegate", nil)
	if err != nil {
		return nil, err
	}
	response.LatestDelegateAddress = *latestDelegateAddress

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := mp.EstimateDelegateUpgradeGas(opts)
	if err == nil {
		response.GasInfo = gasInfo
	}

	// Return response
	return &response, nil

}

func delegateUpgrade(c *cli.Command, minipoolAddress common.Address, opts *bind.TransactOpts) (*api.DelegateUpgradeResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.DelegateUpgradeResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Upgrade
	hash, err := mp.DelegateUpgrade(opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canSetUseLatestDelegate(c *cli.Command, minipoolAddress common.Address) (*api.CanSetUseLatestDelegateResponse, error) {
	setting := true

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
	response := api.CanSetUseLatestDelegateResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	if !setting {
		// Get the version and deposit type
		depositType, err := minipool.GetMinipoolDepositType(rp, minipoolAddress, nil)
		if err != nil {
			return nil, fmt.Errorf("error getting minipool %s deposit type: %w", minipoolAddress.Hex(), err)
		}
		version := mp.GetVersion()
		if depositType == rptypes.Variable && version == 3 {
			// Get the previous delegate
			oldDelegate, err := mp.GetDelegate(nil)
			if err != nil {
				return nil, fmt.Errorf("error getting old delegate for minipool %s: %w", minipoolAddress.Hex(), err)
			}

			// Get the version
			oldDelegateVersion, err := rocketpool.GetContractVersion(rp, oldDelegate, nil)
			if err != nil {
				return nil, fmt.Errorf("error getting version of old delegate %s for minipool %s: %w", oldDelegate.Hex(), minipoolAddress.Hex(), err)
			}

			if oldDelegateVersion == 2 {
				return nil, fmt.Errorf("you cannot unset 'use-latest-delegate' for minipool %s after reducing your ETH bond, as this would revert to the Redstone delegate and render your minipool unable to distribute its balance; please upgrade your minipool's delegate first before unsetting this flag", minipoolAddress.Hex())
			}
		}
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := mp.EstimateSetUseLatestDelegateGas(opts)
	if err == nil {
		response.GasInfo = gasInfo
	}

	// Return response
	return &response, nil

}

func setUseLatestDelegate(c *cli.Command, minipoolAddress common.Address, opts *bind.TransactOpts) (*api.SetUseLatestDelegateResponse, error) {
	setting := true

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.SetUseLatestDelegateResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	if !setting {
		// Get the version and deposit type
		depositType, err := minipool.GetMinipoolDepositType(rp, minipoolAddress, nil)
		if err != nil {
			return nil, fmt.Errorf("error getting minipool %s deposit type: %w", minipoolAddress.Hex(), err)
		}
		version := mp.GetVersion()
		if depositType == rptypes.Variable && version == 3 {
			// Get the previous delegate
			oldDelegate, err := mp.GetDelegate(nil)
			if err != nil {
				return nil, fmt.Errorf("error getting old delegate for minipool %s: %w", minipoolAddress.Hex(), err)
			}

			// Get the version
			oldDelegateVersion, err := rocketpool.GetContractVersion(rp, oldDelegate, nil)
			if err != nil {
				return nil, fmt.Errorf("error getting version of old delegate %s for minipool %s: %w", oldDelegate.Hex(), minipoolAddress.Hex(), err)
			}

			if oldDelegateVersion == 2 {
				return nil, fmt.Errorf("you cannot unset 'use-latest-delegate' for minipool %s after reducing your ETH bond, as this would revert to the Redstone delegate and render your minipool unable to distribute its balance; please upgrade your minipool's delegate first before unsetting this flag", minipoolAddress.Hex())
			}
		}
	}

	// Set the new setting
	hash, err := mp.SetUseLatestDelegate(opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func getUseLatestDelegate(c *cli.Command, minipoolAddress common.Address) (*api.GetUseLatestDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetUseLatestDelegateResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get data
	setting, err := mp.GetUseLatestDelegate(nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting use latest delegate: %w", err)
	}

	// Return response
	response.Setting = setting
	return &response, nil

}

func getDelegate(c *cli.Command, minipoolAddress common.Address) (*api.GetDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetDelegateResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get data
	address, err := mp.GetDelegate(nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting delegate: %w", err)
	}

	// Return response
	response.Address = address
	return &response, nil

}

func getPreviousDelegate(c *cli.Command, minipoolAddress common.Address) (*api.GetPreviousDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetPreviousDelegateResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get data
	address, err := mp.GetPreviousDelegate(nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting delegate: %w", err)
	}

	// Return response
	response.Address = address
	return &response, nil

}

func getEffectiveDelegate(c *cli.Command, minipoolAddress common.Address) (*api.GetEffectiveDelegateResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetEffectiveDelegateResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get data
	address, err := mp.GetEffectiveDelegate(nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting delegate: %w", err)
	}

	// Return response
	response.Address = address
	return &response, nil

}
