package rp

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	node131 "github.com/rocket-pool/smartnode/bindings/legacy/v1.3.1/node"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	tnsettings "github.com/rocket-pool/smartnode/bindings/settings/trustednode"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"golang.org/x/sync/errgroup"
)

func GetNodeValidatorIndices(rp *rocketpool.RocketPool, ec rocketpool.ExecutionClient, bc beacon.Client, nodeAddress common.Address) ([]string, error) {
	// Get current block number so all subsequent queries are done at same point in time
	blockNumber, err := ec.BlockNumber(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Error getting block number: %w", err)
	}

	// Setup call opts
	blockNumberBig := big.NewInt(0).SetUint64(blockNumber)
	callOpts := bind.CallOpts{BlockNumber: blockNumberBig}

	// Get list of pubkeys for this given node
	pubkeys, err := minipool.GetNodeValidatingMinipoolPubkeys(rp, nodeAddress, &callOpts)
	if err != nil {
		return nil, err
	}

	// Remove zero pubkeys
	zeroPubkey := types.ValidatorPubkey{}
	filteredPubkeys := []types.ValidatorPubkey{}
	for _, pubkey := range pubkeys {
		if !bytes.Equal(pubkey[:], zeroPubkey[:]) {
			filteredPubkeys = append(filteredPubkeys, pubkey)
		}
	}
	pubkeys = filteredPubkeys

	// Get validator statuses by pubkeys
	statuses, err := bc.GetValidatorStatuses(pubkeys, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting validator statuses: %w", err)
	}

	// Enumerate validators statuses and fill indices array
	validatorIndices := make([]string, len(statuses)+1)

	i := 0
	for _, status := range statuses {
		validatorIndices[i] = status.Index
		i++
	}

	return validatorIndices, nil
}

// Checks the given node's current borrowed ETH, its limit on borrowed ETH, and how much ETH is preparing to be borrowed by pending bond reductions
func CheckCollateral(saturnDeployed bool, rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (ethBorrowed *big.Int, ethBorrowedLimit *big.Int, pendingBorrowAmount *big.Int, err error) {
	// Get the node's minipool addresses
	addresses, err := minipool.GetNodeMinipoolAddresses(rp, nodeAddress, opts)
	if err != nil {
		err = fmt.Errorf("error getting minipool addresses for node %s: %w", nodeAddress.Hex(), err)
		return
	}

	latestBlockHeader, err := rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting latest block header: %w", err)
	}
	blockTime := time.Unix(int64(latestBlockHeader.Time), 0)
	var reductionWindowStart uint64
	var reductionWindowLength uint64

	// Data
	var wg1 errgroup.Group

	wg1.Go(func() error {
		var err error
		reductionWindowStart, err = tnsettings.GetBondReductionWindowStart(rp, nil)
		return err
	})
	wg1.Go(func() error {
		var err error
		reductionWindowLength, err = tnsettings.GetBondReductionWindowLength(rp, nil)
		return err
	})

	// Wait for data
	if err = wg1.Wait(); err != nil {
		return nil, nil, nil, err
	}

	reductionWindowEnd := time.Duration(reductionWindowStart+reductionWindowLength) * time.Second

	// Data
	var wg errgroup.Group
	deltas := make([]*big.Int, len(addresses))
	zeroTime := time.Unix(0, 0)

	if saturnDeployed {
		wg.Go(func() error {
			var err error
			ethBorrowed, err = node.GetNodeETHBorrowed(rp, nodeAddress, opts)
			if err != nil {
				return fmt.Errorf("error getting node's borrowed ETH amount: %w", err)
			}
			return nil
		})
	} else {
		wg.Go(func() error {
			var err error
			ethBorrowed, err = node131.GetNodeEthMatched(rp, nodeAddress, opts)
			if err != nil {
				return fmt.Errorf("error getting node's borrowed ETH amount: %w", err)
			}
			return nil
		})
		wg.Go(func() error {
			var err error
			// Matched is renamed borrowed in Saturn v1.4
			ethBorrowedLimit, err = node131.GetNodeEthMatchedLimit(rp, nodeAddress, opts)
			if err != nil {
				return fmt.Errorf("error getting how much ETH the node is able to borrow: %w", err)
			}
			return nil
		})
	}

	for i, address := range addresses {
		wg.Go(func() error {
			reduceBondTime, err := minipool.GetReduceBondTime(rp, address, opts)
			if err != nil {
				return fmt.Errorf("error getting bond reduction time for minipool %s: %w", address.Hex(), err)
			}

			reduceBondCancelled, err := minipool.GetReduceBondCancelled(rp, address, nil)
			if err != nil {
				return fmt.Errorf("error getting bond reduction cancel status for minipool %s: %w", address.Hex(), err)
			}

			// Ignore minipools that don't have a bond reduction pending
			timeSinceReductionStart := blockTime.Sub(reduceBondTime)
			if reduceBondTime == zeroTime ||
				reduceBondCancelled ||
				timeSinceReductionStart > reductionWindowEnd {
				deltas[i] = big.NewInt(0)
				return nil
			}

			// Get the old and new (pending) bonds
			mp, err := minipool.NewMinipool(rp, address, opts)
			if err != nil {
				return fmt.Errorf("error creating binding for minipool %s: %w", address.Hex(), err)
			}
			oldBond, err := mp.GetNodeDepositBalance(opts)
			if err != nil {
				return fmt.Errorf("error getting node deposit balance for minipool %s: %w", address.Hex(), err)
			}
			newBond, err := minipool.GetReduceBondValue(rp, address, opts)
			if err != nil {
				return fmt.Errorf("error getting pending bond reduced balance for minipool %s: %w", address.Hex(), err)
			}

			// Delta = old - new
			deltas[i] = big.NewInt(0).Sub(oldBond, newBond)
			return nil
		})
	}

	// Wait for data
	if err = wg.Wait(); err != nil {
		return
	}

	// Get the total pending borrow amount
	totalDelta := big.NewInt(0)
	for _, delta := range deltas {
		totalDelta.Add(totalDelta, delta)
	}
	pendingBorrowAmount = totalDelta

	return
}
