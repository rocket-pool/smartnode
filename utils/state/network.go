package state

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/multicall"
)

type NetworkDetails struct {
	// Redstone
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
	BalancesBlock                     *big.Int
	LatestReportableBalancesBlock     *big.Int
	SubmitBalancesEnabled             bool

	// Atlas
	PromotionScrubPeriod      time.Duration
	BondReductionWindowStart  time.Duration
	BondReductionWindowLength time.Duration
	DepositPoolUserBalance    *big.Int
}

// TODO: Finish this, involves porting e.g. GetClaimIntervalTime() over
func _getNetworkDetailsFast(rp *rocketpool.RocketPool, multicallerAddress common.Address, contracts *NetworkContracts, opts *bind.CallOpts) (*NetworkDetails, error) {

	mc, err := multicall.NewMultiCaller(rp.Client, multicallerAddress)
	if err != nil {
		return nil, err
	}

	details := &NetworkDetails{}

	var rewardIndex *big.Int
	mc.AddCall(contracts.RocketNetworkPrices, &details.RplPrice, "getRPLPrice")
	mc.AddCall(contracts.RocketDAOProtocolSettingsNode, &details.MinCollateralFraction, "getMinimumPerMinipoolStake")
	mc.AddCall(contracts.RocketDAOProtocolSettingsNode, &details.MaxCollateralFraction, "getMaximumPerMinipoolStake")
	mc.AddCall(contracts.RocketRewardsPool, &rewardIndex, "getRewardIndex")

	details.RewardIndex = rewardIndex.Uint64()

	/*
		wg.Go(func() error {
			var err error
			state.NetworkDetails.IntervalDuration, err = GetClaimIntervalTime(cfg, state.NetworkDetails.RewardIndex, rp, opts)
			if err != nil {
				return fmt.Errorf("error getting interval duration: %w", err)
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.IntervalStart, err = rewards.GetClaimIntervalTimeStart(rp, opts)
			if err != nil {
				return fmt.Errorf("error getting interval start: %w", err)
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.NodeOperatorRewardsPercent, err = GetNodeOperatorRewardsPercent(cfg, state.NetworkDetails.RewardIndex, rp, opts)
			if err != nil {
				return fmt.Errorf("error getting node operator rewards percent")
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.TrustedNodeOperatorRewardsPercent, err = GetTrustedNodeOperatorRewardsPercent(cfg, state.NetworkDetails.RewardIndex, rp, opts)
			if err != nil {
				return fmt.Errorf("error getting trusted node operator rewards percent")
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.ProtocolDaoRewardsPercent, err = GetProtocolDaoRewardsPercent(cfg, state.NetworkDetails.RewardIndex, rp, opts)
			if err != nil {
				return fmt.Errorf("error getting protocol DAO rewards percent")
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.PendingRPLRewards, err = GetPendingRPLRewards(cfg, state.NetworkDetails.RewardIndex, rp, opts)
			if err != nil {
				return fmt.Errorf("error getting pending RPL rewards")
			}
			return nil
		})

		if isAtlasDeployed {
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
		}

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
			var err error
			state.NetworkDetails.DepositPoolBalance, err = deposit.GetBalance(rp, opts)
			if err != nil {
				return fmt.Errorf("error getting deposit pool balance: %w", err)
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.DepositPoolExcess, err = deposit.GetExcessBalance(rp, opts)
			if err != nil {
				return fmt.Errorf("error getting deposit pool excess: %w", err)
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.QueueCapacity, err = minipool.GetQueueCapacity(rp, opts)
			if err != nil {
				return fmt.Errorf("error getting minipool queue capacity: %w", err)
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.RPLInflationIntervalRate, err = tokens.GetRPLInflationIntervalRate(rp, opts)
			if err != nil {
				return fmt.Errorf("error getting RPL inflation interval: %w", err)
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.RPLTotalSupply, err = tokens.GetRPLTotalSupply(rp, opts)
			if err != nil {
				return fmt.Errorf("error getting total RPL supply: %w", err)
			}
			return nil
		})

		wg.Go(func() error {
			var err error
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
			var err error
			state.NetworkDetails.ETHUtilizationRate, err = network.GetETHUtilizationRate(rp, opts)
			if err != nil {
				return fmt.Errorf("error getting ETH utilization rate: %w", err)
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.StakingETHBalance, err = network.GetStakingETHBalance(rp, opts)
			if err != nil {
				return fmt.Errorf("error getting total ETH staking balance: %w", err)
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.RETHExchangeRate, err = tokens.GetRETHExchangeRate(rp, opts)
			if err != nil {
				return fmt.Errorf("error getting ETH-rETH exchange rate: %w", err)
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.TotalETHBalance, err = network.GetTotalETHBalance(rp, opts)
			if err != nil {
				return fmt.Errorf("error getting total ETH balance (TVL): %w", err)
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			rethAddress := cfg.Smartnode.GetRethAddress()
			state.NetworkDetails.RETHBalance, err = rp.Client.BalanceAt(context.Background(), rethAddress, opts.BlockNumber)
			if err != nil {
				return fmt.Errorf("error getting ETH balance of rETH staking contract: %w", err)
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.TotalRETHSupply, err = tokens.GetRETHTotalSupply(rp, opts)
			if err != nil {
				return fmt.Errorf("error getting total rETH supply: %w", err)
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.TotalRPLStake, err = node.GetTotalRPLStake(rp, opts)
			if err != nil {
				return fmt.Errorf("error getting total amount of RPL staked on the network: %w", err)
			}
			return nil
		})

		wg.Go(func() error {
			var err error
			state.NetworkDetails.NodeFee, err = network.GetNodeFee(rp, opts)
			if err != nil {
				return fmt.Errorf("error getting current node fee for new minipools: %w", err)
			}
			return nil
		})

		// Wait for data
		if err := wg.Wait(); err != nil {
			return err
		}

		return nil
	*/

	return details, nil
}
