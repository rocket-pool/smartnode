package minipool

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canRefundMinipool(c *cli.Command, minipoolAddress common.Address) (*api.CanRefundMinipoolResponse, error) {

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
	response := api.CanRefundMinipoolResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Validate minipool owner
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	if err := validateMinipoolOwner(mp, nodeAccount.Address); err != nil {
		return nil, err
	}

	// Check node refund balance
	refundBalance, err := mp.GetNodeRefundBalance(nil)
	if err != nil {
		return nil, err
	}
	response.InsufficientRefundBalance = (refundBalance.Cmp(big.NewInt(0)) == 0)

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasLimits, err := mp.EstimateRefundGas(opts)
	if err == nil {
		response.GasLimits = gasLimits
	}

	// Update & return response
	response.CanRefund = !response.InsufficientRefundBalance
	return &response, nil

}

func refundMinipool(c *cli.Command, minipoolAddress common.Address, t *snroute.TransactOpts) (*api.RefundMinipoolResponse, error) {
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
	response := api.RefundMinipoolResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Refund
	hash, err := mp.Refund(opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canRefundHandler(ctx snroute.Context) {
	addr, err := parseAddress(ctx.Request, "address")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canRefundMinipool(ctx.Command(), addr)
	response.WriteResponse(ctx.Writer, resp, err)
}

func refundHandler(ctx snroute.WriteContext) {
	addr, err := parseAddress(ctx.Request, "address")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := refundMinipool(ctx.Command(), addr, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
