package node

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canNodeUnstakeLegacyRpl(c *cli.Context, amountWei *big.Int) (*api.CanNodeUnstakeLegacyRplResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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

	// Response
	response := api.CanNodeUnstakeLegacyRplResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Data
	var wg errgroup.Group
	var legacyRplStake *big.Int
	nodeRplLocked := big.NewInt(0)
	var isRPLWithdrawalAddressSet bool
	var rplWithdrawalAddress common.Address
	rplStakeThreshold := big.NewInt(0)

	// Get RPL stake
	wg.Go(func() error {
		var err error
		legacyRplStake, err = node.GetNodeLegacyStakedRPL(rp, nodeAccount.Address, nil)
		return err
	})

	// Check if the RPL withdrawal address is set
	wg.Go(func() error {
		var err error
		isRPLWithdrawalAddressSet, err = node.GetNodeRPLWithdrawalAddressIsSet(rp, nodeAccount.Address, nil)
		return err
	})

	// Get the RPL withdrawal address
	wg.Go(func() error {
		var err error
		rplWithdrawalAddress, err = node.GetNodeRPLWithdrawalAddress(rp, nodeAccount.Address, nil)
		return err
	})

	// Get RPL locked on node
	wg.Go(func() error {
		var err error
		nodeRplLocked, err = node.GetNodeLockedRPL(rp, nodeAccount.Address, nil)
		return err
	})

	// Get the minimum amount of legacy staked RPL a node must have after unstaking
	wg.Go(func() error {
		var err error
		rplStakeThreshold, err = node.GetNodeMinimumLegacyRPLStake(rp, nodeAccount.Address, nil)
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasInfo, err := node.EstimateUnstakeLegacyRPLGas(rp, amountWei, opts)
		if err == nil {
			response.GasInfo = gasInfo
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Check data
	var remainingLegacyRplStake big.Int
	remainingLegacyRplStake.Sub(legacyRplStake, amountWei)
	remainingLegacyRplStake.Sub(&remainingLegacyRplStake, nodeRplLocked)
	response.InsufficientBalance = (amountWei.Cmp(legacyRplStake) > 0)
	response.HasDifferentRPLWithdrawalAddress = (isRPLWithdrawalAddressSet && nodeAccount.Address != rplWithdrawalAddress)
	response.BelowMaxRPLStake = (remainingLegacyRplStake.Cmp(rplStakeThreshold) < 0)

	// Update & return response
	response.CanUnstake = !(response.InsufficientBalance || response.HasDifferentRPLWithdrawalAddress)
	return &response, nil

}

func nodeUnstakeLegacyRpl(c *cli.Context, amountWei *big.Int) (*api.NodeUnstakeLegacyRplResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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

	// Response
	response := api.NodeUnstakeLegacyRplResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}
	var hash common.Hash
	// Unstake legacy RPL
	hash, err = node.UnstakeLegacyRPL(rp, amountWei, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
