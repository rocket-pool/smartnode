package node

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/rewards"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/tokens"
	"github.com/rocket-pool/smartnode/bindings/types"
	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"
	"github.com/rocket-pool/smartnode/shared/math"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	mpApi "github.com/rocket-pool/smartnode/rocketpool/api/minipool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	rprewards "github.com/rocket-pool/smartnode/shared/services/rewards"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Settings
const minipoolBalanceDetailsBatchSize = 20

// Beacon chain balance info for a minipool
type minipoolBalanceDetails struct {
	nodeDeposit *big.Int
	nodeBalance *big.Int
}

// Get the balances of the minipools on the beacon chain
func getBeaconBalances(rp *rocketpool.RocketPool, bc beacon.Client, addresses []common.Address, beaconHead beacon.BeaconHead, opts *bind.CallOpts) ([]minipoolBalanceDetails, error) {

	// Get minipool validator statuses
	validators, err := mpApi.GetMinipoolValidators(rp, bc, addresses, opts, &beacon.ValidatorStatusOptions{Epoch: &beaconHead.Epoch})
	if err != nil {
		return []minipoolBalanceDetails{}, err
	}

	// Load details in batches
	details := make([]minipoolBalanceDetails, len(addresses))
	for bsi := 0; bsi < len(addresses); bsi += minipoolBalanceDetailsBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := min(bsi+minipoolBalanceDetailsBatchSize, len(addresses))

		// Load details
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				address := addresses[mi]
				validator := validators[address]
				mpDetails, err := getMinipoolBalanceDetails(rp, address, opts, validator, beaconHead.Epoch)
				if err == nil {
					details[mi] = mpDetails
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return []minipoolBalanceDetails{}, err
		}

	}

	// Return
	return details, nil
}

// Get minipool balance details
func getMinipoolBalanceDetails(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts, validator beacon.ValidatorStatus, blockEpoch uint64) (minipoolBalanceDetails, error) {

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, opts)
	if err != nil {
		return minipoolBalanceDetails{}, err
	}
	blockBalance := math.GweiToWei(float64(validator.Balance))

	// Data
	var wg errgroup.Group
	var status types.MinipoolStatus
	var nodeDepositBalance *big.Int
	var finalized bool

	// Load data
	wg.Go(func() error {
		var err error
		status, err = mp.GetStatus(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		nodeDepositBalance, err = mp.GetNodeDepositBalance(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		finalized, err = mp.GetFinalised(opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return minipoolBalanceDetails{}, err
	}

	// Deal with pools that haven't received deposits yet so their balance is still 0
	if nodeDepositBalance == nil {
		nodeDepositBalance = big.NewInt(0)
	}

	// Ignore finalized minipools
	if finalized {
		return minipoolBalanceDetails{
			nodeDeposit: big.NewInt(0),
			nodeBalance: big.NewInt(0),
		}, nil
	}

	// Use node deposit balance if initialized or prelaunch
	if status == types.Initialized || status == types.Prelaunch {
		return minipoolBalanceDetails{
			nodeDeposit: nodeDepositBalance,
			nodeBalance: nodeDepositBalance,
		}, nil
	}

	// Use node deposit balance if validator not yet active on beacon chain at block
	if !validator.Exists || validator.ActivationEpoch >= blockEpoch {
		return minipoolBalanceDetails{
			nodeDeposit: nodeDepositBalance,
			nodeBalance: nodeDepositBalance,
		}, nil
	}

	// Get node balance at block
	nodeBalance, err := mp.CalculateNodeShare(blockBalance, opts)
	if err != nil {
		return minipoolBalanceDetails{}, err
	}

	// Return
	return minipoolBalanceDetails{
		nodeDeposit: nodeDepositBalance,
		nodeBalance: nodeBalance,
	}, nil

}

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
	bc, err := services.GetBeaconClient(c)
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
	minipoolDetails, err := getBeaconBalances(rp, bc, addresses, beaconHead, nil)
	if err != nil {
		return nil, err
	}
	for _, minipool := range minipoolDetails {
		totalDepositBalance += math.WeiToEth(minipool.nodeDeposit)
		totalNodeShare += math.WeiToEth(minipool.nodeBalance)
	}
	response.BeaconRewards = totalNodeShare - totalDepositBalance

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
