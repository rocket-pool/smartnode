package megapool

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func getStatus(c *cli.Context) (*api.MegapoolStatusResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	if err := services.RequireBeaconClientSynced(c); err != nil {
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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolStatusResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	details, err := services.GetNodeMegapoolDetails(rp, bc, nodeAccount.Address)
	if err != nil {
		return nil, err
	}
	response.Megapool = details

	// Get latest delegate address
	delegate, err := rp.GetContract("rocketMegapoolDelegate", nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting latest minipool delegate contract: %w", err)
	}
	response.LatestDelegate = *delegate.Address

	// Return response
	return &response, nil
}

func calculateRewards(c *cli.Context, amount *big.Int) (*api.MegapoolRewardSplitResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	if err := services.RequireBeaconClientSynced(c); err != nil {
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
	response := api.MegapoolRewardSplitResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Calculate the rewards split for a given amount
	response, err = services.CalculateRewards(rp, amount, nodeAccount.Address)
	if err != nil {
		return nil, fmt.Errorf("Error getting rewards split for amount %s: %w", amount, err)
	}

	//Return response
	return &response, nil
}

func calculatePendingRewards(c *cli.Context) (*api.MegapoolRewardSplitResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	if err := services.RequireBeaconClientSynced(c); err != nil {
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
	response := api.MegapoolRewardSplitResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get the megapool address
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Load the megapool
	mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Calculate the rewards split for a given amount
	pendingRewards, err := mp.CalculatePendingRewards(nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting pending rewards: %w", err)
	}
	response.RewardSplit = pendingRewards

	//Return response
	return &response, nil
}

// Get a map of the node's validator states, the total beacon balance, and the node share of beacon balance
func getValidatorMapAndBalances(c *cli.Context) (*api.MegapoolValidatorMapAndRewardsResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	if err := services.RequireBeaconClientSynced(c); err != nil {
		return nil, err
	}

	status, err := getStatus(c)
	if err != nil {
		return nil, fmt.Errorf("Error getting the megapool status")
	}

	// Response
	response := api.MegapoolValidatorMapAndRewardsResponse{}

	statusValidators := map[string][]api.MegapoolValidatorDetails{
		"Staking":     {},
		"Exited":      {},
		"Initialized": {},
		"Prelaunch":   {},
		"Dissolved":   {},
		"Exiting":     {},
		"Locked":      {},
	}

	var totalBeaconBalance uint64
	var totalEffectiveBeaconBalance uint64
	// Iterate over the validators and append them based on their statuses
	for _, validator := range status.Megapool.Validators {
		if validator.Staked && !validator.Exited && !validator.Exiting {
			statusValidators["Staking"] = append(statusValidators["Staking"], validator)
			if validator.Activated {
				totalBeaconBalance += validator.BeaconStatus.Balance
				totalEffectiveBeaconBalance += validator.BeaconStatus.EffectiveBalance
			}
		}
		if validator.Exited {
			statusValidators["Exited"] = append(statusValidators["Exited"], validator)
		}
		if validator.InQueue {
			statusValidators["Initialized"] = append(statusValidators["Initialized"], validator)
		}
		if validator.InPrestake {
			statusValidators["Prelaunch"] = append(statusValidators["Prelaunch"], validator)
		}
		if validator.Dissolved {
			statusValidators["Dissolved"] = append(statusValidators["Dissolved"], validator)
		}
		if validator.Exiting {
			statusValidators["Exiting"] = append(statusValidators["Exiting"], validator)
		}
		if validator.Locked {
			statusValidators["Locked"] = append(statusValidators["Locked"], validator)
		}
	}
	// Store map in the api response
	response.MegapoolValidatorMap = statusValidators

	weiPerGwei := big.NewInt(int64(eth.WeiPerGwei))
	totalBeaconBalanceWei := new(big.Int).SetUint64(totalBeaconBalance)
	totalEffectiveBeaconBalanceWei := new(big.Int).SetUint64(totalEffectiveBeaconBalance)
	totalBeaconBalanceWei = totalBeaconBalanceWei.Mul(totalBeaconBalanceWei, weiPerGwei)
	totalEffectiveBeaconBalanceWei = totalEffectiveBeaconBalanceWei.Mul(totalEffectiveBeaconBalanceWei, weiPerGwei)

	// Get the node share of CL rewards
	nodeShareOfCLBalance := big.NewInt(0)
	if totalBeaconBalanceWei.Cmp(totalEffectiveBeaconBalanceWei) <= 0 {
		nodeShareOfCLBalance = big.NewInt(0)
	} else {
		toBeSkimmed := new(big.Int).Sub(totalBeaconBalanceWei, totalEffectiveBeaconBalanceWei)
		rewards, err := calculateRewards(c, toBeSkimmed)
		if err != nil {
			return &response, fmt.Errorf("Error calculating the rewards split for amount %s: %w", toBeSkimmed.String(), err)
		}
		nodeShareOfCLBalance = nodeShareOfCLBalance.Add(rewards.RewardSplit.NodeRewards, status.Megapool.NodeBond)
	}
	response.TotalBeaconBalance = totalBeaconBalanceWei
	response.NodeShareOfCLBalance = nodeShareOfCLBalance
	response.NodeBond = status.Megapool.NodeBond

	// Return response
	return &response, nil

}
