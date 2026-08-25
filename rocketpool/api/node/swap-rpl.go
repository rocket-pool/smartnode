package node

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/tokens"
	"github.com/rocket-pool/smartnode/bindings/utils"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canNodeSwapRpl(c *cli.Command, amountWei *big.Int) (*api.CanNodeSwapRplResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
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
	response := api.CanNodeSwapRplResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Check node fixed-supply RPL balance
	fixedSupplyRplBalance, err := tokens.GetFixedSupplyRPLBalance(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.InsufficientBalance = (amountWei.Cmp(fixedSupplyRplBalance) > 0)

	// Get gas estimates
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasLimits, err := tokens.EstimateSwapFixedSupplyRPLForRPLGas(rp, amountWei, opts)
	if err != nil {
		return nil, err
	}
	response.GasLimits = gasLimits

	// Update & return response
	response.CanSwap = !response.InsufficientBalance
	return &response, nil

}

func allowanceFsRpl(c *cli.Command) (*api.NodeSwapRplAllowanceResponse, error) {

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
	response := api.NodeSwapRplAllowanceResponse{}

	// Get new RPL contract address
	rocketTokenRPLAddress, err := rp.GetAddress("rocketTokenRPL", nil)
	if err != nil {
		return nil, err
	}

	// Get node account
	account, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get node's FSRPL allowance
	allowance, err := tokens.GetFixedSupplyRPLAllowance(rp, account.Address, *rocketTokenRPLAddress, nil)
	if err != nil {
		return nil, err
	}

	response.Allowance = allowance

	return &response, nil
}

func getSwapApprovalGas(c *cli.Command, amountWei *big.Int) (*api.NodeSwapRplApproveGasResponse, error) {
	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
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
	response := api.NodeSwapRplApproveGasResponse{}

	// Get RPL contract address
	rocketTokenRPLAddress, err := rp.GetAddress("rocketTokenRPL", nil)
	if err != nil {
		return nil, err
	}

	// Get gas estimates
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasLimits, err := tokens.EstimateApproveFixedSupplyRPLGas(rp, *rocketTokenRPLAddress, amountWei, opts)
	if err != nil {
		return nil, err
	}
	response.GasLimits = gasLimits
	return &response, nil
}

func approveFsRpl(c *cli.Command, amountWei *big.Int, t *snroute.TransactOpts) (*api.NodeSwapRplApproveResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeSwapRplApproveResponse{}

	// Get RPL contract address
	rocketTokenRPLAddress, err := rp.GetAddress("rocketTokenRPL", nil)
	if err != nil {
		return nil, err
	}

	// Approve fixed-supply RPL allowance
	if response.ApproveTxHash, err = tokens.ApproveFixedSupplyRPL(rp, *rocketTokenRPLAddress, amountWei, opts); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}

func waitForApprovalAndSwapFsRpl(c *cli.Command, amountWei *big.Int, hash common.Hash, t *snroute.TransactOpts) (*api.NodeSwapRplSwapResponse, error) {
	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Wait for the fixed-supply RPL approval TX to successfully get included in a block
	_, err = utils.WaitForTransaction(rp.Client, hash)
	if err != nil {
		return nil, err
	}

	return swapRpl(c, amountWei, t)

}

func swapRpl(c *cli.Command, amountWei *big.Int, t *snroute.TransactOpts) (*api.NodeSwapRplSwapResponse, error) {
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
	response := api.NodeSwapRplSwapResponse{}

	// Swap fixed-supply RPL for RPL
	if response.SwapTxHash, err = tokens.SwapFixedSupplyRPLForRPL(rp, amountWei, opts); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}

func swapRplAllowanceHandler(ctx snroute.Context) {
	resp, err := allowanceFsRpl(ctx.Command())
	response.WriteResponse(ctx.Writer, resp, err)
}

func canSwapRplHandler(ctx snroute.Context) {
	amountWei, err := parseNodeBigInt(ctx.Request, "amountWei")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canNodeSwapRpl(ctx.Command(), amountWei)
	response.WriteResponse(ctx.Writer, resp, err)
}

func getSwapRplApprovalGasHandler(ctx snroute.Context) {
	amountWei, err := parseNodeBigInt(ctx.Request, "amountWei")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := getSwapApprovalGas(ctx.Command(), amountWei)
	response.WriteResponse(ctx.Writer, resp, err)
}

func swapRplApproveRplHandler(ctx snroute.WriteContext) {
	amountWei, err := parseNodeBigInt(ctx.Request, "amountWei")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := approveFsRpl(ctx.Command(), amountWei, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}

func waitAndSwapRplHandler(ctx snroute.WriteContext) {
	amountWei, err := parseNodeBigInt(ctx.Request, "amountWei")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	hash := common.HexToHash(ctx.Request.FormValue("approvalTxHash"))
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := waitForApprovalAndSwapFsRpl(ctx.Command(), amountWei, hash, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}

func swapRplHandler(ctx snroute.WriteContext) {
	amountWei, err := parseNodeBigInt(ctx.Request, "amountWei")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := swapRpl(ctx.Command(), amountWei, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
