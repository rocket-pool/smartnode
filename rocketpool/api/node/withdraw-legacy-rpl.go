package node

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/units"
)

func canNodeUnstakeLegacyRpl(c *cli.Command, amountWei units.Wei) (*api.CanNodeUnstakeLegacyRplResponse, error) {

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
	var legacyRplStake units.Wei
	var nodeRplLocked units.Wei
	var isRPLWithdrawalAddressSet bool
	var rplWithdrawalAddress common.Address
	var rplStakeThreshold units.Wei

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
		gasLimits, err := node.EstimateUnstakeLegacyRPLGas(rp, amountWei, opts)
		if err == nil {
			response.GasLimits = gasLimits
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Check data
	var remainingLegacyRplStake units.Wei
	remainingLegacyRplStake = legacyRplStake.Sub(amountWei)
	remainingLegacyRplStake = remainingLegacyRplStake.Sub(nodeRplLocked)
	response.InsufficientBalance = (amountWei.Cmp(legacyRplStake) > 0)
	response.HasDifferentRPLWithdrawalAddress = (isRPLWithdrawalAddressSet && nodeAccount.Address != rplWithdrawalAddress)
	response.BelowMaxRPLStake = (remainingLegacyRplStake.Cmp(rplStakeThreshold) < 0)

	// Update & return response
	response.CanUnstake = !response.InsufficientBalance && !response.HasDifferentRPLWithdrawalAddress
	return &response, nil

}

func nodeUnstakeLegacyRpl(c *cli.Command, amountWei units.Wei, t *snroute.TransactOpts) (*api.NodeUnstakeLegacyRplResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeUnstakeLegacyRplResponse{}

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

func canUnstakeLegacyRplHandler(ctx snroute.Context) {
	amountWei, err := parseNodeWei(ctx.Request, "amountWei")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canNodeUnstakeLegacyRpl(ctx.Command(), amountWei)
	response.WriteResponse(ctx.Writer, resp, err)
}

func unstakeLegacyRplHandler(ctx snroute.WriteContext) {
	amountWei, err := parseNodeWei(ctx.Request, "amountWei")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := nodeUnstakeLegacyRpl(ctx.Command(), amountWei, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
