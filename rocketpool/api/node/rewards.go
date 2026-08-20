package node

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/rewards"
	"github.com/rocket-pool/smartnode/bindings/tokens"
	"github.com/rocket-pool/smartnode/bindings/types"
	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"
	"github.com/rocket-pool/smartnode/shared/math"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

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

	var totalEffectiveStake *big.Int
	var totalRplSupply *big.Int
	var inflationInterval *big.Int
	var odaoSize uint64
	var nodeOperatorRewardsPercent float64
	var trustedNodeOperatorRewardsPercent float64
	var totalDepositBalance float64
	var totalNodeShare float64
	var networkState *state.NetworkState

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
		unclaimedRplRewardsWei := big.NewInt(0)
		rplRewards := big.NewInt(0)
		unclaimedEthRewardsWei := big.NewInt(0)
		ethRewards := big.NewInt(0)

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
			rplRewards.Add(rplRewards, &intervalInfo.CollateralRplAmount.Int)
			ethRewards.Add(ethRewards, &intervalInfo.TotalEthAmount.Int)
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
				unclaimedRplRewardsWei.Add(unclaimedRplRewardsWei, &intervalInfo.CollateralRplAmount.Int)
				unclaimedEthRewardsWei.Add(unclaimedEthRewardsWei, &intervalInfo.TotalEthAmount.Int)
			}
		}

		if err == nil {
			response.CumulativeRplRewards = math.WeiToEth(rplRewards)
			response.UnclaimedRplRewards = math.WeiToEth(unclaimedRplRewardsWei)
			response.CumulativeEthRewards = math.WeiToEth(ethRewards)
			response.UnclaimedEthRewards = math.WeiToEth(unclaimedEthRewardsWei)
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
			response.TotalRplStake = math.WeiToEth(stake)
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
		nodeOperatorRewardsPercent = math.WeiToEth(nodeOperatorRewardsPercentRaw)
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
		if nodeDeposit == nil {
			nodeDeposit = big.NewInt(0)
		}

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

		totalDepositBalance += math.WeiToEth(nodeDeposit)
		totalNodeShare += math.WeiToEth(nodeShare)
	}
	response.BeaconRewards = totalNodeShare - totalDepositBalance

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

			megapoolReward := big.NewInt(0)
			if totalBeaconBalance > totalEffectiveBeaconBalance {
				weiPerGwei := big.NewInt(int64(math.WeiPerGwei))
				totalBeaconBalanceWei := new(big.Int).Mul(new(big.Int).SetUint64(totalBeaconBalance), weiPerGwei)
				totalEffectiveBeaconBalanceWei := new(big.Int).Mul(new(big.Int).SetUint64(totalEffectiveBeaconBalance), weiPerGwei)
				toBeSkimmed := new(big.Int).Sub(totalBeaconBalanceWei, totalEffectiveBeaconBalanceWei)

				rewardSplit, err := services.CalculateRewards(rp, toBeSkimmed, nodeAccount.Address)
				if err != nil {
					return nil, fmt.Errorf("Error calculating megapool rewards split for amount %s: %w", toBeSkimmed.String(), err)
				}
				megapoolReward = rewardSplit.RewardSplit.NodeRewards
			}
			response.BeaconRewards += math.WeiToEth(megapoolReward)
		}
	}

	// Calculate the estimated rewards
	rewardsIntervalDays := response.RewardsInterval.Seconds() / (60 * 60 * 24)
	inflationPerDay := math.WeiToEth(inflationInterval)
	totalRplAtNextCheckpoint := (math.Pow(inflationPerDay, float64(rewardsIntervalDays)) - 1) * math.WeiToEth(totalRplSupply)
	if totalRplAtNextCheckpoint < 0 {
		totalRplAtNextCheckpoint = 0
	}

	if totalEffectiveStake.Cmp(big.NewInt(0)) == 1 {
		response.EstimatedRewards = response.EffectiveRplStake / math.WeiToEth(totalEffectiveStake) * totalRplAtNextCheckpoint * nodeOperatorRewardsPercent
	}

	if response.Trusted {

		var wg2 errgroup.Group

		// Get cumulative ODAO rewards
		wg2.Go(func() error {
			// Legacy rewards
			unclaimedRplRewardsWei := big.NewInt(0)
			rplRewards := big.NewInt(0)

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
				rplRewards.Add(rplRewards, &intervalInfo.ODaoRplAmount.Int)
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
					unclaimedRplRewardsWei.Add(unclaimedRplRewardsWei, &intervalInfo.ODaoRplAmount.Int)
				}
			}

			if err == nil {
				response.CumulativeTrustedRplRewards = math.WeiToEth(rplRewards)
				response.UnclaimedTrustedRplRewards = math.WeiToEth(unclaimedRplRewardsWei)
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
			trustedNodeOperatorRewardsPercent = math.WeiToEth(trustedNodeOperatorRewardsPercentRaw)
			if err != nil {
				return err
			}
			return nil
		})

		// Get the node's oDAO RPL stake
		wg2.Go(func() error {
			bond, err := trustednode.GetMemberRPLBondAmount(rp, nodeAccount.Address, nil)
			if err == nil {
				response.TrustedRplBond = math.WeiToEth(bond)
			}
			return err
		})

		// Wait for data
		if err := wg2.Wait(); err != nil {
			return nil, err
		}

		response.EstimatedTrustedRplRewards = totalRplAtNextCheckpoint * trustedNodeOperatorRewardsPercent / float64(odaoSize)

	}

	// Return response
	return &response, nil

}
