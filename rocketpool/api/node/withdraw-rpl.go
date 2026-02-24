package node

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	node131 "github.com/rocket-pool/smartnode/bindings/legacy/v1.3.1/node"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canNodeWithdrawRpl(c *cli.Context) (*api.CanNodeWithdrawRplResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanNodeWithdrawRplResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Data
	var wg errgroup.Group
	var rplUnstaking *big.Int
	var currentTime uint64
	var rplLastUnstakedTime uint64
	var unstakingPeriod time.Duration
	var isRPLWithdrawalAddressSet bool
	var rplWithdrawalAddress common.Address

	// Get RPL stake
	wg.Go(func() error {
		var err error
		rplUnstaking, err = node.GetNodeUnstakingRPL(rp, nodeAccount.Address, nil)
		return err
	})

	// Get current block
	wg.Go(func() error {
		header, err := ec.HeaderByNumber(context.Background(), nil)
		if err == nil {
			currentTime = header.Time
		}
		return err
	})

	// Get RPL staked time
	wg.Go(func() error {
		var err error
		rplLastUnstakedTime, err = node.GetNodeLastUnstakeTime(rp, nodeAccount.Address, nil)
		return err
	})

	// Get withdrawal delay
	wg.Go(func() error {
		var err error
		unstakingPeriod, err = protocol.GetRewardsClaimIntervalTime(rp, nil)
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

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasInfo, err := node.EstimateWithdrawRPLGas(rp, opts)
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

	response.InsufficientBalance = (rplUnstaking.Cmp(big.NewInt(0)) > 0)
	response.UnstakingPeriodActive = ((currentTime - rplLastUnstakedTime) < uint64(unstakingPeriod.Seconds()))
	response.HasDifferentRPLWithdrawalAddress = (isRPLWithdrawalAddressSet && nodeAccount.Address != rplWithdrawalAddress)

	// Update & return response
	response.CanWithdraw = !(response.InsufficientBalance || response.UnstakingPeriodActive || response.HasDifferentRPLWithdrawalAddress)
	return &response, nil

}

func nodeWithdrawRpl(c *cli.Context) (*api.NodeWithdrawRplResponse, error) {

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
	response := api.NodeWithdrawRplResponse{}

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
	// Withdraw RPL
	hash, err = node.WithdrawRPL(rp, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

// Used if saturn is not deployed (v1.3.1)
func canNodeWithdrawRplv1_3_1(c *cli.Context, amountWei *big.Int) (*api.CanNodeWithdrawRplv1_3_1Response, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanNodeWithdrawRplv1_3_1Response{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Data
	var wg errgroup.Group
	var rplStake *big.Int
	var minimumRplStake *big.Int
	var maximumRplStake *big.Int
	nodeRplLocked := big.NewInt(0)
	var currentTime uint64
	var rplStakedTime uint64
	var withdrawalDelay time.Duration
	var isRPLWithdrawalAddressSet bool
	var rplWithdrawalAddress common.Address

	// Get RPL stake
	wg.Go(func() error {
		var err error
		rplStake, err = node131.GetNodeRPLStake(rp, nodeAccount.Address, nil)
		return err
	})

	// Get minimum RPL stake
	wg.Go(func() error {
		var err error
		minimumRplStake, err = node131.GetNodeMinimumRPLStake(rp, nodeAccount.Address, nil)
		return err
	})

	// Get maximum RPL stake
	wg.Go(func() error {
		var err error
		maximumRplStake, err = node131.GetNodeMaximumRPLStake(rp, nodeAccount.Address, nil)
		return err
	})

	// Get current block
	wg.Go(func() error {
		header, err := ec.HeaderByNumber(context.Background(), nil)
		if err == nil {
			currentTime = header.Time
		}
		return err
	})

	// Get RPL staked time
	wg.Go(func() error {
		var err error
		rplStakedTime, err = node131.GetNodeRPLStakedTime(rp, nodeAccount.Address, nil)
		return err
	})

	// Get withdrawal delay
	wg.Go(func() error {
		var err error
		withdrawalDelay, err = protocol.GetRewardsClaimIntervalTime(rp, nil)
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
		nodeRplLocked, err = node131.GetNodeRPLLocked(rp, nodeAccount.Address, nil)
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasInfo, err := node131.EstimateWithdrawRPLGas(rp, nodeAccount.Address, amountWei, opts)
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
	var remainingRplStake big.Int
	remainingRplStake.Sub(rplStake, amountWei)
	remainingRplStake.Sub(&remainingRplStake, nodeRplLocked)
	response.InsufficientBalance = (amountWei.Cmp(rplStake) > 0)
	response.BelowMaxRPLStake = (remainingRplStake.Cmp(maximumRplStake) < 0)
	response.MinipoolsUndercollateralized = (remainingRplStake.Cmp(minimumRplStake) < 0)
	response.WithdrawalDelayActive = ((currentTime - rplStakedTime) < uint64(withdrawalDelay.Seconds()))
	response.HasDifferentRPLWithdrawalAddress = (isRPLWithdrawalAddressSet && nodeAccount.Address != rplWithdrawalAddress)

	// Update & return response
	response.CanWithdraw = !(response.InsufficientBalance || response.MinipoolsUndercollateralized || response.WithdrawalDelayActive || response.HasDifferentRPLWithdrawalAddress || response.BelowMaxRPLStake)
	return &response, nil
}

// Used if saturn is not deployed (v1.3.1)
func nodeWithdrawRplv1_3_1(c *cli.Context, amountWei *big.Int) (*api.NodeWithdrawRplResponse, error) {

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

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeWithdrawRplResponse{}

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
	// Withdraw RPL
	hash, err = node131.WithdrawRPL(rp, nodeAccount.Address, amountWei, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil
}
