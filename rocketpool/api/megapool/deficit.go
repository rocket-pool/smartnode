package megapool

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getDeficit(c *cli.Command, address common.Address) (*api.MegapoolDeficitResponse, error) {
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	// Read all inputs at one block so the displayed deficit and selected count
	// describe the same state. Estimation and submission recheck current state.
	header, err := rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	opts := &bind.CallOpts{BlockNumber: header.Number}
	saturn2Deployed, err := state.IsSaturn2Deployed(rp, opts)
	if err != nil {
		return nil, fmt.Errorf("error checking if Saturn 2 is deployed: %w", err)
	}
	if !saturn2Deployed {
		return &api.MegapoolDeficitResponse{Saturn2Deployed: false}, nil
	}
	mp, err := megapool.NewMegapool(rp, address, opts)
	if err != nil {
		return nil, err
	}
	if mp.GetVersion() < 2 {
		return nil, fmt.Errorf("megapool %s must upgrade to delegate version 2 or later to support forced exits", address.Hex())
	}
	nodeAddress, err := mp.GetNodeAddress(opts)
	if err != nil {
		return nil, err
	}
	var debt, bond, queuedBond, refund, creditAndBalance, threshold *big.Int
	var rewards megapool.RewardSplit
	var active, exiting uint32
	var wg errgroup.Group
	wg.Go(func() error {
		var err error
		debt, err = mp.GetDebt(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		bond, err = mp.GetNodeBond(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		queuedBond, err = mp.GetNodeQueuedBond(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		refund, err = mp.GetRefundValue(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		rewards, err = mp.CalculatePendingRewards(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		active, err = mp.GetActiveValidatorCount(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		exiting, err = mp.GetExitingValidatorCount(opts)
		return err
	})
	wg.Go(func() error {
		var err error
		creditAndBalance, err = node.GetNodeCreditAndBalance(rp, nodeAddress, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		threshold, err = protocol.GetMegapoolExitDeficit(rp, opts)
		return err
	})
	if err := wg.Wait(); err != nil {
		return nil, err
	}
	assets := new(big.Int).Add(creditAndBalance, rewards.NodeRewards)
	assets.Add(assets, refund)
	bondRequirement := func(remaining uint32) (*big.Int, error) {
		return node.GetBondRequirement(rp, new(big.Int).SetUint64(uint64(remaining)), opts)
	}
	result, err := planDeficitExits(debt, assets, bond, queuedBond, threshold, active, exiting, bondRequirement)
	if err != nil {
		return nil, err
	}
	result.Saturn2Deployed = true
	if !result.CanExit {
		return result, nil
	}
	ids, err := firstDeficitExitValidators(mp, result.ValidatorsRequired, opts)
	if err != nil {
		return nil, err
	}
	if err := applyEligibleValidators(result, ids, func(additional uint32) (*big.Int, bool, error) {
		return projectDeficitExit(debt, assets, bond, queuedBond, threshold, active, exiting, additional, bondRequirement)
	}); err != nil {
		return nil, err
	}
	return result, nil
}

// planDeficitExits selects a validator count RocketNetworkExit will accept.
// A permissionless caller may exit while the deficit before the last requested
// exit is still at least the exit threshold, so the last exit may cross below
// it. The smallest count that does cross below is also the largest count the
// contract accepts. When no count crosses below, every remaining validator is
// still eligible and is selected.
func planDeficitExits(debt, assets, bond, queuedBond, threshold *big.Int, active, exiting uint32, bondRequirement func(uint32) (*big.Int, error)) (*api.MegapoolDeficitResponse, error) {
	if exiting > active {
		return nil, fmt.Errorf("exiting validator count exceeds active validator count")
	}
	project := func(additional uint32) (*big.Int, bool, error) {
		return projectDeficitExit(debt, assets, bond, queuedBond, threshold, active, exiting, additional, bondRequirement)
	}
	current := new(big.Int).Sub(debt, assets)
	if current.Sign() < 0 {
		current.SetInt64(0)
	}
	result := &api.MegapoolDeficitResponse{
		Deficit: current, ExitDeficit: new(big.Int).Set(threshold), ExistingExits: exiting,
	}
	pending, pendingBelow, err := project(0)
	if err != nil {
		return nil, err
	}
	result.DeficitAfterPendingExits = pending
	result.ProjectedDeficit = pending
	if pendingBelow {
		result.Reason = "No additional exits are permitted; the deficit is already below the amount required for a permissionless exit."
		return result, nil
	}
	maximum := active - exiting
	if maximum == 0 {
		result.Reason = "No active validators remain to exit."
		return result, nil
	}
	lowest, lowestBelow, err := project(maximum)
	if err != nil {
		return nil, err
	}
	if !lowestBelow {
		result.ValidatorsRequired = maximum
		result.ProjectedDeficit = lowest
		result.CanExit = true
		return result, nil
	}
	// Bond requirements are monotonic, so the minimum count that crosses below
	// the threshold can be found without one query per validator.
	low, high := uint32(1), maximum
	for low < high {
		mid := low + (high-low)/2
		_, below, err := project(mid)
		if err != nil {
			return nil, err
		}
		if below {
			high = mid
		} else {
			low = mid + 1
		}
	}
	result.ValidatorsRequired = low
	result.ProjectedDeficit, _, err = project(low)
	if err != nil {
		return nil, err
	}
	result.CanExit = true
	return result, nil
}

// Queued bond is included because the active validator count includes queued
// validators. Release is clamped at zero when underbonded, then capped by
// exiting principal and by the node bond.
func releasedNodeBond(bond, queuedBond, required *big.Int, totalExiting uint32) *big.Int {
	released := new(big.Int).Add(bond, queuedBond)
	released.Sub(released, required)
	if released.Sign() < 0 {
		return new(big.Int)
	}
	principal := new(big.Int).Mul(new(big.Int).SetUint64(uint64(totalExiting)), new(big.Int).Mul(big.NewInt(32), big.NewInt(1e18)))
	if released.Cmp(principal) > 0 {
		released.Set(principal)
	}
	if released.Cmp(bond) > 0 {
		released.Set(bond)
	}
	return released
}

// projectDeficitExit reports the clamped deficit after additional exits and
// whether the unclamped value is strictly under the exit threshold.
func projectDeficitExit(debt, assets, bond, queuedBond, threshold *big.Int, active, exiting, additional uint32, bondRequirement func(uint32) (*big.Int, error)) (*big.Int, bool, error) {
	required, err := bondRequirement(active - (exiting + additional))
	if err != nil {
		return nil, false, err
	}
	raw := new(big.Int).Sub(debt, assets)
	raw.Sub(raw, releasedNodeBond(bond, queuedBond, required, exiting+additional))
	below := raw.Cmp(threshold) < 0
	if raw.Sign() < 0 {
		raw.SetInt64(0)
	}
	return raw, below, nil
}

// applyEligibleValidators shortens the plan when fewer validators can be
// force-exited than the count derived from the active-validator total. A
// shorter positive count still passes the contract's pre-last check unless
// the bond curve is not monotonic.
func applyEligibleValidators(result *api.MegapoolDeficitResponse, ids []uint32, project func(uint32) (*big.Int, bool, error)) error {
	result.ValidatorIds = ids
	eligible := uint32(len(ids))
	if !result.CanExit || eligible >= result.ValidatorsRequired {
		return nil
	}
	if eligible > 0 {
		_, below, err := project(eligible - 1)
		if err != nil {
			return err
		}
		if !below {
			projected, _, err := project(eligible)
			if err != nil {
				return err
			}
			result.ValidatorsRequired = eligible
			result.ProjectedDeficit = projected
			return nil
		}
	}
	result.CanExit = false
	result.Reason = fmt.Sprintf("%d validator exits are required, but only %d validators are eligible for forced exit.", result.ValidatorsRequired, eligible)
	return nil
}

func firstDeficitExitValidators(mp megapool.Megapool, count uint32, opts *bind.CallOpts) ([]uint32, error) {
	total, err := mp.GetValidatorCount(opts)
	if err != nil {
		return nil, err
	}
	ids := []uint32{}
	for id := uint32(0); id < total && uint32(len(ids)) < count; id++ {
		info, err := mp.GetValidatorInfo(id, opts)
		if err != nil {
			return nil, err
		}
		if info.Staked && !info.Dissolved && !info.Exiting && !info.Exited {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func deficitHandler(ctx snroute.Context) {
	address, err := parseDeficitMegapoolAddress(ctx.Request)
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	result, err := getDeficit(ctx.Command(), address)
	response.WriteResponse(ctx.Writer, result, err)
}
