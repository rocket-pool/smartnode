package node

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/rewards"
	"github.com/rocket-pool/smartnode/bindings/tokens"
	"github.com/rocket-pool/smartnode/bindings/types"
	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"
	"github.com/rocket-pool/smartnode/shared/units"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getRewards(c *cli.Command) (*api.NodeRewardsResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	if err := services.RequireEthClientSynced(c); err != nil {
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
	stateProvider, err := services.GetNetworkStateProvider(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeRewardsResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	var totalEffectiveStake units.Wei
	var totalRplSupply units.Wei
	var inflationInterval units.Wei
	var odaoSize uint64
	var nodeOperatorRewardsPercent units.Eth
	var trustedNodeOperatorRewardsPercent units.Eth
	var totalDepositBalance units.Eth
	var totalNodeShare units.Eth
	var networkState *state.NetworkStateIndex

	// Sync
	var wg errgroup.Group

	// Check if the node is registered or not
	wg.Go(func() error {
		exists, err := node.GetNodeExists(rp, nodeAccount.Address, nil)
		if err == nil {
			response.Registered = exists
		}
		return err
	})

	// Get the node registration time
	wg.Go(func() error {
		var time time.Time
		var err error
		time, err = node.GetNodeRegistrationTime(rp, nodeAccount.Address, nil)

		if err == nil {
			response.NodeRegistrationTime = time
		}
		return err
	})

	// Get node trusted status
	wg.Go(func() error {
		trusted, err := trustednode.GetMemberExists(rp, nodeAccount.Address, nil)
		if err == nil {
			response.Trusted = trusted
		}
		return err
	})

	// Get claimed and pending rewards
	wg.Go(func() error {
		// Legacy rewards
		unclaimedRplRewardsWei := units.Wei{}
		rplRewards := units.Wei{}
		unclaimedEthRewardsWei := units.Wei{}
		ethRewards := units.Wei{}

		// Get the claimed and unclaimed intervals
		unclaimed, claimed, err := rprewards.GetClaimStatus(rp, nodeAccount.Address)
		if err != nil {
			return err
		}

		// Get the info for each claimed interval
		for _, claimedInterval := range claimed {
			intervalInfo, err := rprewards.GetIntervalInfo(rp, cfg, nodeAccount.Address, claimedInterval, nil)
			if err != nil {
				return err
			}
			if !intervalInfo.TreeFileExists {
				return fmt.Errorf("Error calculating lifetime node rewards: rewards file %s doesn't exist but interval %d was claimed", intervalInfo.TreeFilePath, claimedInterval)
			}
			rplRewards = rplRewards.Add(units.NewWei(&intervalInfo.CollateralRplAmount.Int))
			ethRewards = ethRewards.Add(units.NewWei(&intervalInfo.TotalEthAmount.Int))
		}

		// Get the unclaimed rewards
		for _, unclaimedInterval := range unclaimed {
			intervalInfo, err := rprewards.GetIntervalInfo(rp, cfg, nodeAccount.Address, unclaimedInterval, nil)
			if err != nil {
				return err
			}
			if !intervalInfo.TreeFileExists {
				return fmt.Errorf("Error calculating lifetime node rewards: rewards file %s doesn't exist and interval %d is unclaimed", intervalInfo.TreeFilePath, unclaimedInterval)
			}
			if intervalInfo.NodeExists {
				unclaimedRplRewardsWei = unclaimedRplRewardsWei.Add(units.NewWei(&intervalInfo.CollateralRplAmount.Int))
				unclaimedEthRewardsWei = unclaimedEthRewardsWei.Add(units.NewWei(&intervalInfo.TotalEthAmount.Int))
			}
		}

		if err == nil {
			response.CumulativeRplRewards = rplRewards.ToEth()
			response.UnclaimedRplRewards = unclaimedRplRewardsWei.ToEth()
			response.CumulativeEthRewards = ethRewards.ToEth()
			response.UnclaimedEthRewards = unclaimedEthRewardsWei.ToEth()
		}
		return err
	})

	// Get the start of the rewards checkpoint
	wg.Go(func() error {
		lastCheckpoint, err := rewards.GetClaimIntervalTimeStart(rp, nil)
		if err == nil {
			response.LastCheckpoint = lastCheckpoint
		}
		return err
	})

	// Get the rewards checkpoint interval
	wg.Go(func() error {
		rewardsInterval, err := rewards.GetClaimIntervalTime(rp, nil)
		if err == nil {
			response.RewardsInterval = rewardsInterval
		}
		return err
	})

	// Get the node's total stake
	wg.Go(func() error {
		stake, err := node.GetNodeStakedRPL(rp, nodeAccount.Address, nil)
		if err == nil {
			response.TotalRplStake = stake.ToEth()
		}
		return err
	})

	// Get the total network effective stake
	wg.Go(func() error {
		multicallerAddress := common.HexToAddress(cfg.Smartnode.GetMulticallAddress())
		balanceBatcherAddress := common.HexToAddress(cfg.Smartnode.GetBalanceBatcherAddress())
		contracts, err := rpstate.NewNetworkContracts(rp, multicallerAddress, balanceBatcherAddress, nil)
		if err != nil {
			return fmt.Errorf("error creating network contract binding: %w", err)
		}
		totalEffectiveStake, err = rpstate.GetTotalEffectiveRplStake(rp, contracts)
		if err != nil {
			return fmt.Errorf("error getting total effective RPL stake: %w", err)
		}
		return nil
	})

	// Get the total RPL supply
	wg.Go(func() error {
		var err error
		totalRplSupply, err = tokens.GetRPLTotalSupply(rp, nil)
		if err != nil {
			return err
		}
		return nil
	})

	// Get the RPL inflation interval
	wg.Go(func() error {
		var err error
		inflationInterval, err = tokens.GetRPLInflationIntervalRate(rp, nil)
		if err != nil {
			return err
		}
		return nil
	})

	// Get the node operator rewards percent
	wg.Go(func() error {
		nodeOperatorRewardsPercentRaw, err := rewards.GetNodeOperatorRewardsPercent(rp, nil)
		nodeOperatorRewardsPercent = nodeOperatorRewardsPercentRaw.ToEth()
		if err != nil {
			return err
		}
		return nil
	})

	// Get the network state, filtered to this node's validators
	wg.Go(func() error {
		_networkState, err := stateProvider.GetHeadStateForNode(nodeAccount.Address)
		if err != nil {
			return fmt.Errorf("Error getting network state: %w", err)
		}
		networkState = _networkState
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Calculate the total deposits and corresponding beacon chain balance share
	intervalEndEpoch := networkState.BeaconSlotNumber / networkState.BeaconConfig.SlotsPerEpoch
	for _, mpd := range networkState.MinipoolDetailsByNode[nodeAccount.Address] {
		if mpd.Finalised {
			// Finalized minipools contribute nothing
			continue
		}

		nodeDeposit := mpd.NodeDepositBalance

		// Default to the deposit balance until the validator is confirmed active on Beacon
		nodeShare := nodeDeposit
		if mpd.Status != types.Initialized && mpd.Status != types.Prelaunch {
			validator, exists := networkState.MinipoolValidatorDetails[mpd.Pubkey]
			if exists && validator.Exists && validator.ActivationEpoch < intervalEndEpoch {
				if validator.ExitEpoch <= intervalEndEpoch {
					// Exited but not finalized -- funds already swept to the minipool's own balance
					nodeShare = mpd.NodeShareOfBalanceIncludingBeacon
				} else {
					nodeShare = mpd.NodeShareOfBeaconBalance
				}
			}
		}

		totalDepositBalance = totalDepositBalance.Add(nodeDeposit.ToEth())
		totalNodeShare = totalNodeShare.Add(nodeShare.ToEth())
	}
	response.BeaconRewards = totalNodeShare.Sub(totalDepositBalance)

	// Add the megapool's unskimmed CL rewards
	nodeDetails, exists := networkState.NodeDetailsByAddress[nodeAccount.Address]
	if exists && nodeDetails.MegapoolDeployed {
		megapoolDetails, mdExists := networkState.MegapoolDetails[nodeDetails.MegapoolAddress]
		if mdExists && !megapoolDetails.DelegateExpired {
			var totalBeaconBalance, totalEffectiveBeaconBalance uint64
			for _, pubkey := range networkState.MegapoolToPubkeysMap[nodeDetails.MegapoolAddress] {
				info, infoExists := networkState.GetMegapoolValidatorInfo(nodeDetails.MegapoolAddress, pubkey)
				if !infoExists {
					continue
				}
				vi := info.ValidatorInfo
				if !vi.Staked || vi.Exited || vi.Exiting {
					continue
				}
				beaconStatus, statusExists := networkState.MegapoolValidatorDetails[pubkey]
				if statusExists && beaconStatus.Exists && intervalEndEpoch > beaconStatus.ActivationEpoch {
					totalBeaconBalance += beaconStatus.Balance
					totalEffectiveBeaconBalance += beaconStatus.EffectiveBalance
				}
			}

			megapoolReward := units.Wei{}
			if totalBeaconBalance > totalEffectiveBeaconBalance {
				totalBeaconBalanceWei := units.NewGwei(totalBeaconBalance).ToWei()
				totalEffectiveBeaconBalanceWei := units.NewGwei(totalEffectiveBeaconBalance).ToWei()
				toBeSkimmed := totalBeaconBalanceWei.Sub(totalEffectiveBeaconBalanceWei)

				rewardSplit, err := services.CalculateRewards(rp, toBeSkimmed, nodeAccount.Address)
				if err != nil {
					return nil, fmt.Errorf("Error calculating megapool rewards split for amount %s: %w", toBeSkimmed.String(), err)
				}
				megapoolReward = rewardSplit.RewardSplit.NodeRewards
			}
			response.BeaconRewards = response.BeaconRewards.Add(megapoolReward.ToEth())
		}
	}

	// Calculate the estimated rewards
	rewardsIntervalDays := response.RewardsInterval.Seconds() / (60 * 60 * 24)
	inflationPerDay := inflationInterval.ToEth()
	totalRplAtNextCheckpoint := (inflationPerDay.Pow(rewardsIntervalDays).Sub(units.NewEth(1))).Mul(totalRplSupply.ToEth())
	if totalRplAtNextCheckpoint.IsNegative() {
		totalRplAtNextCheckpoint = units.Eth{}
	}

	if totalEffectiveStake.IsPositive() {
		response.EstimatedRewards = response.EffectiveRplStake.Div(totalEffectiveStake.ToEth())
		response.EstimatedRewards = response.EstimatedRewards.Mul(totalRplAtNextCheckpoint)
		response.EstimatedRewards = response.EstimatedRewards.Mul(nodeOperatorRewardsPercent)
	}

	if response.Trusted {

		var wg2 errgroup.Group

		// Get cumulative ODAO rewards
		wg2.Go(func() error {
			// Legacy rewards
			unclaimedRplRewardsWei := units.Wei{}
			rplRewards := units.Wei{}

			// Get the claimed and unclaimed intervals
			unclaimed, claimed, err := rprewards.GetClaimStatus(rp, nodeAccount.Address)
			if err != nil {
				return err
			}

			// Get the info for each claimed interval
			for _, claimedInterval := range claimed {
				intervalInfo, err := rprewards.GetIntervalInfo(rp, cfg, nodeAccount.Address, claimedInterval, nil)
				if err != nil {
					return err
				}
				if !intervalInfo.TreeFileExists {
					return fmt.Errorf("Error calculating lifetime node rewards: rewards file %s doesn't exist but interval %d was claimed", intervalInfo.TreeFilePath, claimedInterval)
				}
				rplRewards = rplRewards.Add(units.NewWei(&intervalInfo.ODaoRplAmount.Int))
			}

			// Get the unclaimed rewards
			for _, unclaimedInterval := range unclaimed {
				intervalInfo, err := rprewards.GetIntervalInfo(rp, cfg, nodeAccount.Address, unclaimedInterval, nil)
				if err != nil {
					return err
				}
				if !intervalInfo.TreeFileExists {
					return fmt.Errorf("Error calculating lifetime node rewards: rewards file %s doesn't exist and interval %d is unclaimed", intervalInfo.TreeFilePath, unclaimedInterval)
				}
				if intervalInfo.NodeExists {
					unclaimedRplRewardsWei = unclaimedRplRewardsWei.Add(units.NewWei(&intervalInfo.ODaoRplAmount.Int))
				}
			}

			if err == nil {
				response.CumulativeTrustedRplRewards = rplRewards.ToEth()
				response.UnclaimedTrustedRplRewards = unclaimedRplRewardsWei.ToEth()
			}
			return err
		})

		// Get the ODAO member count
		wg2.Go(func() error {
			var err error
			odaoSize, err = trustednode.GetMemberCount(rp, nil)
			if err != nil {
				return err
			}
			return nil
		})

		// Get the trusted node operator rewards percent
		wg2.Go(func() error {
			trustedNodeOperatorRewardsPercentRaw, err := rewards.GetTrustedNodeOperatorRewardsPercent(rp, nil)
			trustedNodeOperatorRewardsPercent = trustedNodeOperatorRewardsPercentRaw.ToEth()
			if err != nil {
				return err
			}
			return nil
		})

		// Get the node's oDAO RPL stake
		wg2.Go(func() error {
			bond, err := trustednode.GetMemberRPLBondAmount(rp, nodeAccount.Address, nil)
			if err == nil {
				response.TrustedRplBond = bond.ToEth()
			}
			return err
		})

		// Wait for data
		if err := wg2.Wait(); err != nil {
			return nil, err
		}

		response.EstimatedTrustedRplRewards = totalRplAtNextCheckpoint.Mul(trustedNodeOperatorRewardsPercent).Div(units.NewEth(odaoSize))

	}

	// Return response
	return &response, nil

}

func rewardsHandler(ctx snroute.Context) {
	resp, err := getRewards(ctx.Command())
	response.WriteResponse(ctx.Writer, resp, err)
}
