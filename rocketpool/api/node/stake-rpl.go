package node

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	node131 "github.com/rocket-pool/rocketpool-go/legacy/v1.3.1/node"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	updateCheck "github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canNodeStakeRpl(c *cli.Context, amountWei *big.Int) (*api.CanNodeStakeRplResponse, error) {

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
	response := api.CanNodeStakeRplResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Check if Saturn is already deployed
	saturnDeployed, err := updateCheck.IsSaturnDeployed(rp, nil)
	if err != nil {
		return nil, err
	}

	// Check RPL balance
	rplBalance, err := tokens.GetRPLBalance(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.InsufficientBalance = (amountWei.Cmp(rplBalance) > 0)

	if !saturnDeployed {
		// Get the min RPL stake
		minRplStake, err := node131.GetNodeMinimumRPLStake(rp, nodeAccount.Address, nil)
		if err != nil {
			return nil, err
		}
		response.MinimumRplStake = minRplStake
	}

	// Get the max stake fraction
	maxStakeFraction, err := protocol.GetMaximumPerMinipoolStake(rp, nil)
	if err != nil {
		return nil, err
	}
	response.MaximumStakeFraction = maxStakeFraction

	// Get gas estimates
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := node.EstimateStakeGas(rp, amountWei, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo

	// Update & return response
	response.CanStake = !(response.InsufficientBalance)
	return &response, nil

}

func getStakeApprovalGas(c *cli.Context, amountWei *big.Int) (*api.NodeStakeRplApproveGasResponse, error) {
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
	response := api.NodeStakeRplApproveGasResponse{}

	// Get staking contract address
	rocketNodeStakingAddress, err := rp.GetAddress("rocketNodeStaking", nil)
	if err != nil {
		return nil, err
	}

	// Get gas estimates
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := tokens.EstimateApproveRPLGas(rp, *rocketNodeStakingAddress, amountWei, opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo
	return &response, nil
}

func allowanceRpl(c *cli.Context) (*api.NodeStakeRplAllowanceResponse, error) {

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
	response := api.NodeStakeRplAllowanceResponse{}

	// Get staking contract address
	rocketNodeStakingAddress, err := rp.GetAddress("rocketNodeStaking", nil)
	if err != nil {
		return nil, err
	}

	// Get node account
	account, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get node's RPL allowance
	allowance, err := tokens.GetRPLAllowance(rp, account.Address, *rocketNodeStakingAddress, nil)
	if err != nil {
		return nil, err
	}

	response.Allowance = allowance

	return &response, nil
}

func approveRpl(c *cli.Context, amountWei *big.Int) (*api.NodeStakeRplApproveResponse, error) {

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
	response := api.NodeStakeRplApproveResponse{}

	// Get staking contract address
	rocketNodeStakingAddress, err := rp.GetAddress("rocketNodeStaking", nil)
	if err != nil {
		return nil, err
	}

	// Approve RPL allowance
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}
	hash, err := tokens.ApproveRPL(rp, *rocketNodeStakingAddress, amountWei, opts)
	if err != nil {
		return nil, err
	}

	response.ApproveTxHash = hash

	// Return response
	return &response, nil

}

func waitForApprovalAndStakeRpl(c *cli.Context, amountWei *big.Int, hash common.Hash) (*api.NodeStakeRplStakeResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Wait for the RPL approval TX to successfully get included in a block
	_, err = utils.WaitForTransaction(rp.Client, hash)
	if err != nil {
		return nil, err
	}

	// Perform the stake
	return stakeRpl(c, amountWei)

}

func stakeRpl(c *cli.Context, amountWei *big.Int) (*api.NodeStakeRplStakeResponse, error) {

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
	response := api.NodeStakeRplStakeResponse{}

	// Stake RPL
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}
	hash, err := node.StakeRPL(rp, amountWei, opts)
	if err != nil {
		return nil, err
	}

	response.StakeTxHash = hash

	// Return response
	return &response, nil

}
