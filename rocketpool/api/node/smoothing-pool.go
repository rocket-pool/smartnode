package node

import (
	"context"
	"fmt"
	"time"

	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	rocketpoolapi "github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
	"github.com/urfave/cli"
)

func getSmoothingPoolRegistrationStatus(c *cli.Context) (*api.GetSmoothingPoolRegistrationStatusResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetSmoothingPoolRegistrationStatusResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Check registration status
	response.NodeRegistered, err = node.GetSmoothingPoolRegistrationState(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Get registration time
	regChangeTime, err := node.GetSmoothingPoolRegistrationChanged(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Get the rewards interval
	intervalTime, err := rewards.GetClaimIntervalTime(rp, nil)
	if err != nil {
		return nil, err
	}

	// Get the time the user can next change their opt-in status
	latestBlockTimeUnix, err := services.GetEthClientLatestBlockTimestamp(ec)
	if err != nil {
		return nil, err
	}
	latestBlockTime := time.Unix(int64(latestBlockTimeUnix), 0)
	changeAvailableTime := regChangeTime.Add(intervalTime)
	response.TimeLeftUntilChangeable = changeAvailableTime.Sub(latestBlockTime)

	// Return response
	return &response, nil

}

func canSetSmoothingPoolStatus(c *cli.Context, status bool) (*api.CanSetSmoothingPoolRegistrationStatusResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanSetSmoothingPoolRegistrationStatusResponse{}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := node.EstimateSetSmoothingPoolRegistrationStateGas(rp, status, opts)
	if err == nil {
		response.GasInfo = gasInfo
	}

	return &response, err

}

func setSmoothingPoolStatus(c *cli.Context, status bool) (*api.SetSmoothingPoolRegistrationStatusResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}
	d, err := services.GetDocker(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.SetSmoothingPoolRegistrationStatusResponse{}

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

	// Get node account and distributor address
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// If opting in, change the fee recipient to the Smoothing Pool before submitting the TX so the fee recipient is guaranteed to be non-penalizable at all times
	if status {
		smoothingPoolContract, err := rp.GetContract("rocketSmoothingPool", nil)
		if err != nil {
			return nil, err
		}
		distributor, err := node.GetDistributorAddress(rp, nodeAccount.Address, nil)
		if err != nil {
			return nil, err
		}

		err = rocketpool.UpdateFeeRecipientFile(*smoothingPoolContract.Address, cfg)
		if err != nil {
			return nil, err
		}

		// Restart the VC
		err = validator.RestartValidator(cfg, bc, nil, d)
		if err != nil {
			// Set the fee recipient back to the node distributor
			err2 := rocketpool.UpdateFeeRecipientFile(distributor, cfg)
			if err2 != nil {
				return nil, fmt.Errorf("***WARNING***\nError restarting validator: [%s]\nError setting fee recipient back to your node's distributor: [%w]\nYour node now has the Smoothing Pool as its fee recipient, even though you aren't opted in!\nPlease visit the Rocket Pool Discord server for help with these errors, so it can be set back to your node's distributor.", err.Error(), err2)
			}

			// Restart the VC but don't pay attention to the errors, since a restart error got us here in the first place
			err2 = validator.RestartValidator(cfg, bc, nil, d)
			if err2 != nil {
				return nil, fmt.Errorf("***WARNING***\nError restarting validator: [%s]\nError setting fee recipient back to your node's distributor: [%w]\nYour node now has the Smoothing Pool as its fee recipient, even though you aren't opted in!\nPlease visit the Rocket Pool Discord server for help with these errors, so it can be set back to your node's distributor.", err.Error(), err2)
			}

			return nil, fmt.Errorf("Error restarting validator after updating the fee recipient to the Smoothing Pool: [%w]\nYour fee recipient has been set back to your node's distributor contract.\nYou have not been opted into the Smoothing Pool.", err)
		}
	}

	// Set the registration status
	// NOTE: for opt out, this is done *before* updating the fee recipient to prevent any possibility of errors causing the node to use the distributor when the user hasn't actually opted out yet
	hash, err := node.SetSmoothingPoolRegistrationState(rp, status, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func GetSmoothingPoolBalance(rp *rocketpoolapi.RocketPool, ec *services.ExecutionClientManager) (*api.SmoothingRewardsResponse, error) {
	smoothingPoolContract, err := rp.GetContract("rocketSmoothingPool", nil)
	if err != nil {
		return nil, err
	}

	response := api.SmoothingRewardsResponse{}

	balanceWei, err := ec.BalanceAt(context.Background(), *smoothingPoolContract.Address, nil)
	if err != nil {
		return nil, err
	}
	response.EthBalance = balanceWei

	return &response, nil
}
