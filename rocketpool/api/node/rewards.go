package node

import (
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/rewards"
	"github.com/rocket-pool/smartnode/bindings/tokens"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"
	updateCheck "github.com/rocket-pool/smartnode/shared/services/state"

	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth2"
)

func getRewards(c *cli.Context) (*api.NodeRewardsResponse, error) {

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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Check if Saturn is already deployed
	saturnDeployed, err := updateCheck.IsSaturnDeployed(rp, nil)
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

	// Get the event log interval
	/*eventLogInterval, err := cfg.GetEventLogInterval()
	if err != nil {
		return nil, err
	}*/

	// Legacy contract addresses
	//legacyRocketRewardsAddress := cfg.Smartnode.GetLegacyRewardsPoolAddress()
	//legacyClaimNodeAddress := cfg.Smartnode.GetLegacyClaimNodeAddress()
	//legacyClaimTrustedNodeAddress := cfg.Smartnode.GetLegacyClaimTrustedNodeAddress()

	var totalEffectiveStake *big.Int
	var totalRplSupply *big.Int
	var inflationInterval *big.Int
	var odaoSize uint64
	var nodeOperatorRewardsPercent float64
	var trustedNodeOperatorRewardsPercent float64
	var totalDepositBalance float64
	var totalNodeShare float64
	var addresses []common.Address
	var beaconHead beacon.BeaconHead

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
		// TEMP removal of the legacy rewards crawler for now, TODO performance improvements here
		/*
			rplRewards, err := legacyrewards.CalculateLifetimeNodeRewards(rp, nodeAccount.Address, big.NewInt(int64(eventLogInterval)), nil, &legacyRocketRewardsAddress, &legacyClaimNodeAddress)*/
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
			ethRewards.Add(ethRewards, &intervalInfo.SmoothingPoolEthAmount.Int)
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
				unclaimedEthRewardsWei.Add(unclaimedEthRewardsWei, &intervalInfo.SmoothingPoolEthAmount.Int)
			}
		}

		if err == nil {
			response.CumulativeRplRewards = eth.WeiToEth(rplRewards)
			response.UnclaimedRplRewards = eth.WeiToEth(unclaimedRplRewardsWei)
			response.CumulativeEthRewards = eth.WeiToEth(ethRewards)
			response.UnclaimedEthRewards = eth.WeiToEth(unclaimedEthRewardsWei)
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

	// Get the node's effective stake
	wg.Go(func() error {
		effectiveStake, err := node.GetNodeEffectiveRPLStake(rp, nodeAccount.Address, nil)
		if err == nil {
			response.EffectiveRplStake = eth.WeiToEth(effectiveStake)
		}
		return err
	})

	// Get the node's total stake
	wg.Go(func() error {
		stake, err := node.GetNodeRPLStake(rp, nodeAccount.Address, nil)
		if err == nil {
			response.TotalRplStake = eth.WeiToEth(stake)
		}
		return err
	})

	// Get the total network effective stake
	wg.Go(func() error {
		multicallerAddress := common.HexToAddress(cfg.Smartnode.GetMulticallAddress())
		balanceBatcherAddress := common.HexToAddress(cfg.Smartnode.GetBalanceBatcherAddress())
		contracts, err := rpstate.NewNetworkContracts(rp, saturnDeployed, multicallerAddress, balanceBatcherAddress, nil)
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
		nodeOperatorRewardsPercent = eth.WeiToEth(nodeOperatorRewardsPercentRaw)
		if err != nil {
			return err
		}
		return nil
	})

	// Get the list of minipool addresses for this node
	wg.Go(func() error {
		_addresses, err := minipool.GetNodeMinipoolAddresses(rp, nodeAccount.Address, nil)
		if err != nil {
			return fmt.Errorf("Error getting node minipool addresses: %w", err)
		}
		addresses = _addresses
		return nil
	})

	// Get the beacon head
	wg.Go(func() error {
		_beaconHead, err := bc.GetBeaconHead()
		if err != nil {
			return fmt.Errorf("Error getting beacon chain head: %w", err)
		}
		beaconHead = _beaconHead
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Calculate the total deposits and corresponding beacon chain balance share
	minipoolDetails, err := eth2.GetBeaconBalances(rp, bc, addresses, beaconHead, nil)
	if err != nil {
		return nil, err
	}
	for _, minipool := range minipoolDetails {
		totalDepositBalance += eth.WeiToEth(minipool.NodeDeposit)
		totalNodeShare += eth.WeiToEth(minipool.NodeBalance)
	}
	response.BeaconRewards = totalNodeShare - totalDepositBalance

	// Calculate the estimated rewards
	rewardsIntervalDays := response.RewardsInterval.Seconds() / (60 * 60 * 24)
	inflationPerDay := eth.WeiToEth(inflationInterval)
	totalRplAtNextCheckpoint := (math.Pow(inflationPerDay, float64(rewardsIntervalDays)) - 1) * eth.WeiToEth(totalRplSupply)
	if totalRplAtNextCheckpoint < 0 {
		totalRplAtNextCheckpoint = 0
	}

	if totalEffectiveStake.Cmp(big.NewInt(0)) == 1 {
		response.EstimatedRewards = response.EffectiveRplStake / eth.WeiToEth(totalEffectiveStake) * totalRplAtNextCheckpoint * nodeOperatorRewardsPercent
	}

	if response.Trusted {

		var wg2 errgroup.Group

		// Get cumulative ODAO rewards
		wg2.Go(func() error {
			// Legacy rewards
			unclaimedRplRewardsWei := big.NewInt(0)
			rplRewards := big.NewInt(0)
			// TODO: PERFORMANCE IMPROVEMENTS
			//rplRewards, err := legacyrewards.CalculateLifetimeTrustedNodeRewards(rp, nodeAccount.Address, big.NewInt(int64(eventLogInterval)), nil, &legacyRocketRewardsAddress, &legacyClaimTrustedNodeAddress)

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
				response.CumulativeTrustedRplRewards = eth.WeiToEth(rplRewards)
				response.UnclaimedTrustedRplRewards = eth.WeiToEth(unclaimedRplRewardsWei)
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
			trustedNodeOperatorRewardsPercent = eth.WeiToEth(trustedNodeOperatorRewardsPercentRaw)
			if err != nil {
				return err
			}
			return nil
		})

		// Get the node's oDAO RPL stake
		wg2.Go(func() error {
			bond, err := trustednode.GetMemberRPLBondAmount(rp, nodeAccount.Address, nil)
			if err == nil {
				response.TrustedRplBond = eth.WeiToEth(bond)
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
