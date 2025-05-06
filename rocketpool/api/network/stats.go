package network

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/deposit"
	node131 "github.com/rocket-pool/rocketpool-go/legacy/v1.3.1/node"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	rpstate "github.com/rocket-pool/rocketpool-go/utils/state"
	updateCheck "github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getStats(c *cli.Context) (*api.NetworkStatsResponse, error) {

	// Get services
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
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
	response := api.NetworkStatsResponse{}

	// Sync
	var wg errgroup.Group

	// Get the deposit pool balance
	wg.Go(func() error {
		balance, err := deposit.GetBalance(rp, nil)
		if err == nil {
			response.DepositPoolBalance = eth.WeiToEth(balance)
		}
		return err
	})

	// Get the total minipool capacity
	wg.Go(func() error {
		minipoolQueueCapacity, err := minipool.GetQueueCapacity(rp, nil)
		if err == nil {
			response.MinipoolCapacity = eth.WeiToEth(minipoolQueueCapacity.Total)
		}
		return err
	})

	// Get the ETH utilization rate
	wg.Go(func() error {
		stakerUtilization, err := network.GetETHUtilizationRate(rp, nil)
		if err == nil {
			response.StakerUtilization = stakerUtilization
		}
		return err
	})

	// Get node fee
	wg.Go(func() error {
		nodeFee, err := network.GetNodeFee(rp, nil)
		if err == nil {
			response.NodeFee = nodeFee
		}
		return err
	})

	// Get node count
	wg.Go(func() error {
		nodeCount, err := node.GetNodeCount(rp, nil)
		if err == nil {
			response.NodeCount = nodeCount
		}
		return err
	})

	// Get minipool counts
	wg.Go(func() error {
		minipoolCounts, err := minipool.GetMinipoolCountPerStatus(rp, nil)
		if err != nil {
			return err
		}
		response.InitializedMinipoolCount = minipoolCounts.Initialized.Uint64()
		response.PrelaunchMinipoolCount = minipoolCounts.Prelaunch.Uint64()
		response.StakingMinipoolCount = minipoolCounts.Staking.Uint64()
		response.WithdrawableMinipoolCount = minipoolCounts.Withdrawable.Uint64()
		response.DissolvedMinipoolCount = minipoolCounts.Dissolved.Uint64()

		finalizedCount, err := minipool.GetFinalisedMinipoolCount(rp, nil)
		if err != nil {
			return err
		}
		response.FinalizedMinipoolCount = finalizedCount

		return nil
	})

	// Get RPL price
	wg.Go(func() error {
		rplPrice, err := network.GetRPLPrice(rp, nil)
		if err == nil {
			response.RplPrice = eth.WeiToEth(rplPrice)
		}
		return err
	})

	if saturnDeployed {
		// Get total RPL staked
		wg.Go(func() error {
			totalStaked, err := node.GetTotalStakedRPL(rp, nil)
			if err == nil {
				response.TotalRplStaked = eth.WeiToEth(totalStaked)
			}
			return err
		})

		// Get RPL staked on megapools
		wg.Go(func() error {
			megapoolStaked, err := node.GetTotalMegapoolStakedRPL(rp, nil)
			if err == nil {
				response.TotalMegapoolRplStaked = eth.WeiToEth(megapoolStaked)
			}
			return err
		})

		// Get legacy RPL staked
		wg.Go(func() error {
			legacyStaked, err := node.GetTotalLegacyStakedRPL(rp, nil)
			if err == nil {
				response.TotalLegacyRplStaked = eth.WeiToEth(legacyStaked)
			}
			return err
		})

	} else {
		wg.Go(func() error {
			totalStaked, err := node131.GetTotalRPLStake(rp, nil)
			if err == nil {
				response.TotalRplStaked = eth.WeiToEth(totalStaked)
			}
			return err
		})
	}

	// Get total effective RPL staked
	wg.Go(func() error {
		multicallerAddress := common.HexToAddress(cfg.Smartnode.GetMulticallAddress())
		balanceBatcherAddress := common.HexToAddress(cfg.Smartnode.GetBalanceBatcherAddress())
		contracts, err := rpstate.NewNetworkContracts(rp, saturnDeployed, multicallerAddress, balanceBatcherAddress, nil)
		if err != nil {
			return fmt.Errorf("error getting network contracts: %w", err)
		}
		totalEffectiveStake, err := rpstate.GetTotalEffectiveRplStake(rp, contracts)
		if err != nil {
			return fmt.Errorf("error getting total effective stake: %w", err)
		}
		response.EffectiveRplStaked = eth.WeiToEth(totalEffectiveStake)
		return nil
	})

	// Get rETH price
	wg.Go(func() error {
		rethPrice, err := tokens.GetRETHExchangeRate(rp, nil)
		if err == nil {
			response.RethPrice = rethPrice
		}
		return err
	})

	// Get smoothing pool status
	wg.Go(func() error {
		smoothingPoolNodes, err := node.GetSmoothingPoolRegisteredNodeCount(rp, nil)
		if err == nil {
			response.SmoothingPoolNodes = smoothingPoolNodes
		}
		return err
	})

	// Get smoothing pool balance
	wg.Go(func() error {
		// Get the Smoothing Pool contract's balance
		smoothingPoolContract, err := rp.GetContract("rocketSmoothingPool", nil)
		if err != nil {
			return fmt.Errorf("error getting smoothing pool contract: %w", err)
		}
		response.SmoothingPoolAddress = *smoothingPoolContract.Address

		smoothingPoolBalance, err := rp.Client.BalanceAt(context.Background(), *smoothingPoolContract.Address, nil)
		if err != nil {
			return fmt.Errorf("error getting smoothing pool balance: %w", err)
		}

		response.SmoothingPoolBalance = eth.WeiToEth(smoothingPoolBalance)
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Get the TVL
	activeMinipools := response.InitializedMinipoolCount +
		response.PrelaunchMinipoolCount +
		response.StakingMinipoolCount +
		response.WithdrawableMinipoolCount +
		response.DissolvedMinipoolCount
	tvl := float64(activeMinipools)*32 + response.DepositPoolBalance + response.MinipoolCapacity + (response.TotalRplStaked * response.RplPrice)
	response.TotalValueLocked = tvl

	// Return response
	return &response, nil

}
