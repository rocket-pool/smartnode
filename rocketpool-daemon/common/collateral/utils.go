package collateral

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
)

const (
	minipoolReduceDetailsBatchSize int = 200
)

type CollateralAmounts struct {
	EthMatched         *big.Int
	EthMatchedLimit    *big.Int
	PendingMatchAmount *big.Int
}

// Checks the given node's current matched ETH, its limit on matched ETH, and how much ETH is preparing to be matched by pending bond reductions
func CheckCollateral(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*CollateralAmounts, error) {
	// Create the bindings
	node, err := node.NewNode(rp, nodeAddress)
	if err != nil {
		return nil, fmt.Errorf("error getting node %s binding: %w", nodeAddress.Hex(), err)
	}
	mpMgr, err := minipool.NewMinipoolManager(rp)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool manager binding: %w", err)
	}

	// Get the minipool count
	err = rp.Query(nil, opts, node.MinipoolCount)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool count: %w", err)
	}

	// Get the minipool addresses
	addresses, err := node.GetMinipoolAddresses(node.MinipoolCount.Formatted(), opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool addresses: %w", err)
	}

	// Create the minipool bindings
	mps, err := mpMgr.CreateMinipoolsFromAddresses(addresses, false, opts)
	if err != nil {
		return nil, fmt.Errorf("error creating minipool bindings: %w", err)
	}

	// Get the minipool details
	err = rp.BatchQuery(len(addresses), minipoolReduceDetailsBatchSize, func(mc *batch.MultiCaller, i int) error {
		mpv3, isMpv3 := minipool.GetMinipoolAsV3(mps[i])
		if isMpv3 {
			eth.AddQueryablesToMulticall(mc,
				mpv3.ReduceBondTime,
				mpv3.IsBondReduceCancelled,
				mpv3.NodeDepositBalance,
				mpv3.ReduceBondValue,
			)
		}
		return nil
	}, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool details: %w", err)
	}

	return CheckCollateralWithMinipoolCache(rp, nodeAddress, mps, opts)
}

// Checks the given node's current matched ETH, its limit on matched ETH, and how much ETH is preparing to be matched by pending bond reductions
func CheckCollateralWithMinipoolCache(rp *rocketpool.RocketPool, nodeAddress common.Address, minipools []minipool.IMinipool, opts *bind.CallOpts) (*CollateralAmounts, error) {
	// Get the relevant header
	var blockHeader *ethtypes.Header
	var err error
	if opts != nil {
		blockHeader, err = rp.Client.HeaderByNumber(context.Background(), opts.BlockNumber)
	} else {
		blockHeader, err = rp.Client.HeaderByNumber(context.Background(), nil)
	}
	if err != nil {
		return nil, fmt.Errorf("error getting latest block header: %w", err)
	}

	// Get the time and set up opts from the header
	blockTime := time.Unix(int64(blockHeader.Time), 0)
	if opts == nil {
		opts = &bind.CallOpts{
			BlockNumber: blockHeader.Number,
		}
	}

	// Create the bindings
	node, err := node.NewNode(rp, nodeAddress)
	if err != nil {
		return nil, fmt.Errorf("error getting node %s binding: %w", nodeAddress.Hex(), err)
	}
	oMgr, err := oracle.NewOracleDaoManager(rp)
	if err != nil {
		return nil, fmt.Errorf("error getting oracle DAO manager binding: %w", err)
	}
	oSettings := oMgr.Settings

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		eth.AddQueryablesToMulticall(mc,
			node.EthMatched,
			node.EthMatchedLimit,
			oSettings.Minipool.BondReductionWindowStart,
			oSettings.Minipool.BondReductionWindowLength,
		)
		return nil
	}, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	reductionWindowStart := oSettings.Minipool.BondReductionWindowStart.Formatted()
	reductionWindowLength := oSettings.Minipool.BondReductionWindowLength.Formatted()
	reductionWindowEnd := reductionWindowStart + reductionWindowLength

	// Calculate the deltas
	totalDelta := big.NewInt(0)
	zeroTime := time.Unix(0, 0)
	for _, mp := range minipools {
		mpv3, isMpv3 := minipool.GetMinipoolAsV3(mp)
		if !isMpv3 {
			continue
		}
		mpCommon := mp.Common()
		reduceBondTime := mpv3.ReduceBondTime.Formatted()
		reduceBondCancelled := mpv3.IsBondReduceCancelled.Get()

		// Ignore minipools that don't have a bond reduction pending
		timeSinceReductionStart := blockTime.Sub(reduceBondTime)
		if reduceBondTime == zeroTime ||
			reduceBondCancelled ||
			timeSinceReductionStart > reductionWindowEnd {
			continue
		}

		// Calculate the bond delta from the pending reduction
		oldBond := mpCommon.NodeDepositBalance.Get()
		newBond := mpv3.ReduceBondValue.Get()
		mpDelta := big.NewInt(0).Sub(oldBond, newBond)
		totalDelta.Add(totalDelta, mpDelta)
	}

	return &CollateralAmounts{
		EthMatched:         node.EthMatched.Get(),
		EthMatchedLimit:    node.EthMatchedLimit.Get(),
		PendingMatchAmount: totalDelta,
	}, nil
}
