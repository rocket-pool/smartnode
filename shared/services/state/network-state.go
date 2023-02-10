package state

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/settings/trustednode"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"golang.org/x/sync/errgroup"
)

const (
	threadLimit int = 6
)

type NetworkDetails struct {
	RplPrice                          *big.Int
	MinCollateralFraction             *big.Int
	MaxCollateralFraction             *big.Int
	IntervalDuration                  time.Duration
	IntervalStart                     time.Time
	NodeOperatorRewardsPercent        *big.Int
	TrustedNodeOperatorRewardsPercent *big.Int
	ProtocolDaoRewardsPercent         *big.Int
	PendingRPLRewards                 *big.Int
	RewardIndex                       uint64
	PromotionScrubPeriod              time.Duration
	BondReductionWindowStart          time.Duration
	BondReductionWindowLength         time.Duration
	ScrubPeriod                       time.Duration
	SmoothingPoolAddress              common.Address
	DepositPoolBalance                *big.Int
	DepositPoolExcess                 *big.Int
	QueueCapacity                     minipool.QueueCapacity
	RPLInflationIntervalRate          *big.Int
	RPLTotalSupply                    *big.Int
	PricesBlock                       uint64
	LatestReportablePricesBlock       uint64
	ETHUtilizationRate                float64
	StakingETHBalance                 *big.Int
	RETHExchangeRate                  float64
	TotalETHBalance                   *big.Int
	RETHBalance                       *big.Int
	TotalRETHSupply                   *big.Int
	TotalRPLStake                     *big.Int
	SmoothingPoolBalance              *big.Int
	NodeFee                           float64
}

type NetworkState struct {
	// Block / slot for this state
	ElBlockNumber    uint64
	BeaconSlotNumber uint64
	BeaconConfig     beacon.Eth2Config

	// Network details
	NetworkDetails NetworkDetails

	// Node details
	NodeDetails          []node.NativeNodeDetails
	NodeDetailsByAddress map[common.Address]*node.NativeNodeDetails

	// Minipool details
	MinipoolDetails          []minipool.NativeMinipoolDetails
	MinipoolDetailsByAddress map[common.Address]*minipool.NativeMinipoolDetails
	MinipoolDetailsByNode    map[common.Address][]*minipool.NativeMinipoolDetails

	// Validator details
	ValidatorDetails map[types.ValidatorPubkey]beacon.ValidatorStatus

	// Internal fields
	log *log.ColorLogger
}

func CreateNetworkState(cfg *config.RocketPoolConfig, rp *rocketpool.RocketPool, ec rocketpool.ExecutionClient, bc beacon.Client, log *log.ColorLogger, slotNumber uint64, beaconConfig beacon.Eth2Config) (*NetworkState, error) {
	// Get the execution block for the given slot
	beaconBlock, exists, err := bc.GetBeaconBlock(fmt.Sprintf("%d", slotNumber))
	if err != nil {
		return nil, fmt.Errorf("error getting Beacon block for slot %d: %w", slotNumber, err)
	}
	if !exists {
		return nil, fmt.Errorf("slot %d did not have a Beacon block", slotNumber)
	}

	// Get the corresponding block on the EL
	elBlockNumber := beaconBlock.ExecutionBlockNumber
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(elBlockNumber),
	}

	// Create the state wrapper
	state := &NetworkState{
		NodeDetailsByAddress:     map[common.Address]*node.NativeNodeDetails{},
		MinipoolDetailsByAddress: map[common.Address]*minipool.NativeMinipoolDetails{},
		MinipoolDetailsByNode:    map[common.Address][]*minipool.NativeMinipoolDetails{},
		BeaconConfig:             beaconConfig,
		log:                      log,
	}

	// Network details
	var wg errgroup.Group
	wg.SetLimit(threadLimit)

	wg.Go(func() error {
		state.NetworkDetails.RplPrice, err = network.GetRPLPrice(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting RPL price ratio: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.MinCollateralFraction, err = protocol.GetMinimumPerMinipoolStakeRaw(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting minimum per minipool stake: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.MaxCollateralFraction, err = protocol.GetMaximumPerMinipoolStakeRaw(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting maximum per minipool stake: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		rewardIndex, err := rewards.GetRewardIndex(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting reward index: %w", err)
		}
		state.NetworkDetails.RewardIndex = rewardIndex.Uint64()
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.IntervalDuration, err = GetClaimIntervalTime(cfg, state.NetworkDetails.RewardIndex, rp, opts)
		if err != nil {
			return fmt.Errorf("error getting interval duration: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.IntervalStart, err = rewards.GetClaimIntervalTimeStart(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting interval start: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.NodeOperatorRewardsPercent, err = GetNodeOperatorRewardsPercent(cfg, state.NetworkDetails.RewardIndex, rp, opts)
		if err != nil {
			return fmt.Errorf("error getting node operator rewards percent")
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.TrustedNodeOperatorRewardsPercent, err = GetTrustedNodeOperatorRewardsPercent(cfg, state.NetworkDetails.RewardIndex, rp, opts)
		if err != nil {
			return fmt.Errorf("error getting trusted node operator rewards percent")
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.ProtocolDaoRewardsPercent, err = GetProtocolDaoRewardsPercent(cfg, state.NetworkDetails.RewardIndex, rp, opts)
		if err != nil {
			return fmt.Errorf("error getting protocol DAO rewards percent")
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.PendingRPLRewards, err = GetPendingRPLRewards(cfg, state.NetworkDetails.RewardIndex, rp, opts)
		if err != nil {
			return fmt.Errorf("error getting pending RPL rewards")
		}
		return nil
	})

	wg.Go(func() error {
		promotionScrubPeriodSeconds, err := trustednode.GetPromotionScrubPeriod(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting promotion scrub period: %w", err)
		}
		state.NetworkDetails.PromotionScrubPeriod = time.Duration(promotionScrubPeriodSeconds) * time.Second
		return nil
	})

	wg.Go(func() error {
		windowStartRaw, err := trustednode.GetBondReductionWindowStart(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting bond reduction window start: %w", err)
		}
		state.NetworkDetails.BondReductionWindowStart = time.Duration(windowStartRaw) * time.Second
		return nil
	})

	wg.Go(func() error {
		windowLengthRaw, err := trustednode.GetBondReductionWindowLength(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting bond reduction window length: %w", err)
		}
		state.NetworkDetails.BondReductionWindowLength = time.Duration(windowLengthRaw) * time.Second
		return nil
	})

	wg.Go(func() error {
		scrubPeriodSeconds, err := trustednode.GetScrubPeriod(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting scrub period: %w", err)
		}
		state.NetworkDetails.ScrubPeriod = time.Duration(scrubPeriodSeconds) * time.Second
		return nil
	})

	wg.Go(func() error {
		smoothingPoolContract, err := rp.GetContract("rocketSmoothingPool", opts)
		if err != nil {
			return fmt.Errorf("error getting smoothing pool contract: %w", err)
		}
		state.NetworkDetails.SmoothingPoolAddress = *smoothingPoolContract.Address

		state.NetworkDetails.SmoothingPoolBalance, err = rp.Client.BalanceAt(context.Background(), *smoothingPoolContract.Address, opts.BlockNumber)
		if err != nil {
			return fmt.Errorf("error getting smoothing pool balance: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.DepositPoolBalance, err = deposit.GetBalance(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting deposit pool balance: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.DepositPoolExcess, err = deposit.GetExcessBalance(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting deposit pool excess: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.QueueCapacity, err = minipool.GetQueueCapacity(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting minipool queue capacity: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.RPLInflationIntervalRate, err = tokens.GetRPLInflationIntervalRate(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting RPL inflation interval: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.RPLTotalSupply, err = tokens.GetRPLTotalSupply(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting total RPL supply: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.PricesBlock, err = network.GetPricesBlock(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting ETH1 prices block: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		latestReportableBlock, err := network.GetLatestReportablePricesBlock(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting ETH1 latest reportable block: %w", err)
		}
		state.NetworkDetails.LatestReportablePricesBlock = latestReportableBlock.Uint64()
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.ETHUtilizationRate, err = network.GetETHUtilizationRate(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting ETH utilization rate: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.StakingETHBalance, err = network.GetStakingETHBalance(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting total ETH staking balance: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.RETHExchangeRate, err = tokens.GetRETHExchangeRate(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting ETH-rETH exchange rate: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.TotalETHBalance, err = network.GetTotalETHBalance(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting total ETH balance (TVL): %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		rethAddress := cfg.Smartnode.GetRethAddress()
		state.NetworkDetails.RETHBalance, err = rp.Client.BalanceAt(context.Background(), rethAddress, opts.BlockNumber)
		if err != nil {
			return fmt.Errorf("error getting ETH balance of rETH staking contract: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.TotalRETHSupply, err = tokens.GetRETHTotalSupply(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting total rETH supply: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.TotalRPLStake, err = node.GetTotalRPLStake(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting total amount of RPL staked on the network: %w", err)
		}
		return nil
	})

	wg.Go(func() error {
		state.NetworkDetails.NodeFee, err = network.GetNodeFee(rp, opts)
		if err != nil {
			return fmt.Errorf("error getting current node fee for new minipools: %w", err)
		}
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Node details
	state.logLine("Getting network state for EL block %d, Beacon slot %d", elBlockNumber, slotNumber)
	start := time.Now()

	nodeDetails, err := node.GetAllNativeNodeDetails(rp, opts)
	if err != nil {
		return nil, err
	}
	state.NodeDetails = nodeDetails
	state.logLine("1/4 - Retrieved node details (%s so far)", time.Since(start))

	// Minipool details
	minipoolDetails, err := minipool.GetAllNativeMinipoolDetails(rp, opts)
	if err != nil {
		return nil, err
	}
	state.logLine("2/4 - Retrieved minipool details (%s so far)", time.Since(start))

	// Create the node lookup
	for _, details := range nodeDetails {
		state.NodeDetailsByAddress[details.NodeAddress] = &details
	}

	// Create the minipool lookups
	pubkeys := make([]types.ValidatorPubkey, 0, len(minipoolDetails))
	emptyPubkey := types.ValidatorPubkey{}
	for _, details := range minipoolDetails {
		state.MinipoolDetailsByAddress[details.MinipoolAddress] = &details
		if details.Pubkey != emptyPubkey {
			pubkeys = append(pubkeys, details.Pubkey)
		}

		// The map of nodes to minipools
		nodeList, exists := state.MinipoolDetailsByNode[details.NodeAddress]
		if !exists {
			nodeList = []*minipool.NativeMinipoolDetails{}
		}
		nodeList = append(nodeList, &details)
		state.MinipoolDetailsByNode[details.NodeAddress] = nodeList
	}
	state.logLine("3/4 - Created lookups (%s so far)", time.Since(start))

	// Get the validator stats from Beacon
	statusMap, err := bc.GetValidatorStatuses(pubkeys, &beacon.ValidatorStatusOptions{
		Slot: &slotNumber,
	})
	if err != nil {
		return nil, err
	}
	state.ValidatorDetails = statusMap
	state.logLine("4/4 - Retrieved validator details (total time: %s)", time.Since(start))

	return state, nil
}

// Logs a line if the logger is specified
func (s *NetworkState) logLine(format string, v ...interface{}) {
	if s.log != nil {
		s.log.Printlnf(format, v...)
	}
}

// Calculate the effective stakes of all nodes in the state
func (s *NetworkState) CalculateEffectiveStakes(scaleByParticipation bool) (map[common.Address]*big.Int, *big.Int, error) {
	effectiveStakes := make(map[common.Address]*big.Int, len(s.NodeDetails))
	totalEffectiveStake := big.NewInt(0)
	intervalDurationBig := big.NewInt(int64(s.NetworkDetails.IntervalDuration.Seconds()))
	slotTime := time.Unix(int64(s.BeaconConfig.GenesisTime), 0).Add(time.Duration(s.BeaconSlotNumber*s.BeaconConfig.SecondsPerSlot) * time.Second)

	nodeCount := uint64(len(s.NodeDetails))
	effectiveStakeSlice := make([]*big.Int, nodeCount)

	// Get the effective stake for each node
	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	for i, node := range s.NodeDetails {
		i := i
		node := node
		wg.Go(func() error {
			eligibleBorrowedEth := big.NewInt(0)
			eligibleBondedEth := big.NewInt(0)
			for _, mpd := range s.MinipoolDetailsByNode[node.NodeAddress] {
				// It must exist and be staking
				if mpd.Exists && mpd.Status == types.Staking {
					// Doesn't exist on Beacon yet
					validatorStatus, exists := s.ValidatorDetails[mpd.Pubkey]
					if !exists {
						s.logLine("NOTE: minipool %s (pubkey %s) didn't exist, ignoring it in effective RPL calculation", mpd.MinipoolAddress.Hex(), mpd.Pubkey.Hex())
						continue
					}

					// Starts too late
					intervalEndEpoch := s.BeaconSlotNumber / s.BeaconConfig.SlotsPerEpoch
					if validatorStatus.ActivationEpoch > intervalEndEpoch {
						s.logLine("NOTE: Minipool %s starts on epoch %d which is after interval epoch %d so it's not eligible for RPL rewards", mpd.MinipoolAddress.Hex(), validatorStatus.ActivationEpoch, intervalEndEpoch)
						continue
					}

					// Already exited
					if validatorStatus.ExitEpoch <= intervalEndEpoch {
						s.logLine("NOTE: Minipool %s exited on epoch %d which is not after interval epoch %d so it's not eligible for RPL rewards", mpd.MinipoolAddress.Hex(), validatorStatus.ExitEpoch, intervalEndEpoch)
						continue
					}
					// It's eligible, so add up the borrowed and bonded amounts
					eligibleBorrowedEth.Add(eligibleBorrowedEth, mpd.UserDepositBalance)
					eligibleBondedEth.Add(eligibleBondedEth, mpd.NodeDepositBalance)
				}
			}

			// minCollateral := borrowedEth * minCollateralFraction / ratio
			// NOTE: minCollateralFraction and ratio are both percentages, but multiplying and dividing by them cancels out the need for normalization by eth.EthToWei(1)
			minCollateral := big.NewInt(0).Mul(eligibleBorrowedEth, s.NetworkDetails.MinCollateralFraction)
			minCollateral.Div(minCollateral, s.NetworkDetails.RplPrice)

			// maxCollateral := bondedEth * maxCollateralFraction / ratio
			// NOTE: maxCollateralFraction and ratio are both percentages, but multiplying and dividing by them cancels out the need for normalization by eth.EthToWei(1)
			maxCollateral := big.NewInt(0).Mul(eligibleBondedEth, s.NetworkDetails.MaxCollateralFraction)
			maxCollateral.Div(maxCollateral, s.NetworkDetails.RplPrice)

			// Calculate the effective stake
			nodeStake := big.NewInt(0).Set(node.RplStake)
			if nodeStake.Cmp(minCollateral) == -1 {
				// Under min collateral
				nodeStake.SetUint64(0)
			} else if nodeStake.Cmp(maxCollateral) == 1 {
				// Over max collateral
				nodeStake.Set(maxCollateral)
			}

			// Scale the effective stake by the participation in the current interval
			if scaleByParticipation {
				// Get the timestamp of the node's registration
				regTimeBig := node.RegistrationTime
				regTime := time.Unix(regTimeBig.Int64(), 0)

				// Get the actual effective stake, scaled based on participation
				eligibleDuration := slotTime.Sub(regTime)
				if eligibleDuration < s.NetworkDetails.IntervalDuration {
					eligibleSeconds := big.NewInt(int64(eligibleDuration / time.Second))
					nodeStake.Mul(nodeStake, eligibleSeconds)
					nodeStake.Div(nodeStake, intervalDurationBig)
				}
			}

			effectiveStakeSlice[i] = nodeStake
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, nil, err
	}

	// Tally everything up and make the node stake map
	for i, nodeStake := range effectiveStakeSlice {
		node := s.NodeDetails[i]
		effectiveStakes[node.NodeAddress] = nodeStake
		totalEffectiveStake.Add(totalEffectiveStake, nodeStake)
	}

	return effectiveStakes, totalEffectiveStake, nil

}
