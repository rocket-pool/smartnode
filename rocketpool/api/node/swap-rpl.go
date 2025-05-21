package node

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/tokens"
	"github.com/rocket-pool/smartnode/bindings/utils"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canNodeSwapRpl(c *cli.Context, amountWei *big.Int) (*api.CanNodeSwapRplResponse, error) {

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
	gasInfo, err := tokens.EstimateSwapFixedSupplyRPLForRPLGas(rp, amountWei, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	// Update & return response
	response.CanSwap = !response.InsufficientBalance
	return &response, nil

}

func allowanceFsRpl(c *cli.Context) (*api.NodeSwapRplAllowanceResponse, error) {

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

func getSwapApprovalGas(c *cli.Context, amountWei *big.Int) (*api.NodeSwapRplApproveGasResponse, error) {
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
	gasInfo, err := tokens.EstimateApproveFixedSupplyRPLGas(rp, *rocketTokenRPLAddress, amountWei, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo
	return &response, nil
}

func approveFsRpl(c *cli.Context, amountWei *big.Int) (*api.NodeSwapRplApproveResponse, error) {

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
	response := api.NodeSwapRplApproveResponse{}

	// Get RPL contract address
	rocketTokenRPLAddress, err := rp.GetAddress("rocketTokenRPL", nil)
	if err != nil {
		return nil, err
	}

	// Approve fixed-supply RPL allowance
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}
	if hash, err := tokens.ApproveFixedSupplyRPL(rp, *rocketTokenRPLAddress, amountWei, opts); err != nil {
		return nil, err
	} else {
		response.ApproveTxHash = hash
	}

	// Return response
	return &response, nil

}

func waitForApprovalAndSwapFsRpl(c *cli.Context, amountWei *big.Int, hash common.Hash) (*api.NodeSwapRplSwapResponse, error) {

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

	return swapRpl(c, amountWei)

}

func swapRpl(c *cli.Context, amountWei *big.Int) (*api.NodeSwapRplSwapResponse, error) {

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
	response := api.NodeSwapRplSwapResponse{}

	// Swap fixed-supply RPL for RPL
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}
	if hash, err := tokens.SwapFixedSupplyRPLForRPL(rp, amountWei, opts); err != nil {
		return nil, err
	} else {
		response.SwapTxHash = hash
	}

	// Return response
	return &response, nil

}
