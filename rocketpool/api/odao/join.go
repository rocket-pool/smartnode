package odao

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	tndao "github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	tnsettings "github.com/rocket-pool/smartnode/bindings/settings/trustednode"
	"github.com/rocket-pool/smartnode/bindings/tokens"
	"github.com/rocket-pool/smartnode/bindings/utils"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canJoin(c *cli.Command) (*api.CanJoinTNDAOResponse, error) {

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
	response := api.CanJoinTNDAOResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Data
	var wg errgroup.Group
	var nodeRplBalance *big.Int
	var rplBondAmount *big.Int

	// Check proposal actionable status
	wg.Go(func() error {
		proposalActionable, err := getProposalIsActionable(rp, nodeAccount.Address, "invited")
		if err == nil {
			response.ProposalExpired = !proposalActionable
		}
		return err
	})

	// Check if already a member
	wg.Go(func() error {
		isMember, err := tndao.GetMemberExists(rp, nodeAccount.Address, nil)
		if err == nil {
			response.AlreadyMember = isMember
		}
		return err
	})

	// Get node RPL balance
	wg.Go(func() error {
		var err error
		nodeRplBalance, err = tokens.GetRPLBalance(rp, nodeAccount.Address, nil)
		return err
	})

	// Get RPL bond amount
	wg.Go(func() error {
		var err error
		rplBondAmount, err = tnsettings.GetRPLBond(rp, nil)
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		rocketDAONodeTrustedActionsAddress, err := rp.GetAddress("rocketDAONodeTrustedActions", nil)
		if err != nil {
			return err
		}
		rplBondAmount, err := tnsettings.GetRPLBond(rp, nil)
		if err != nil {
			return err
		}
		approveGasLimits, err := tokens.EstimateApproveRPLGas(rp, *rocketDAONodeTrustedActionsAddress, rplBondAmount, opts)
		if err != nil {
			return err
		}
		//joinGasInfo, err := tndao.EstimateJoinGas(rp, opts)
		if err == nil {
			response.GasLimits = approveGasLimits
			//response.GasLimits.EstGasLimit += joinGasInfo.EstGasLimit
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Check data
	response.InsufficientRplBalance = (nodeRplBalance.Cmp(rplBondAmount) < 0)

	// Update & return response
	response.CanJoin = !response.ProposalExpired && !response.AlreadyMember && !response.InsufficientRplBalance
	return &response, nil

}

func approveRpl(c *cli.Command, t *snroute.TransactOpts) (*api.JoinTNDAOApproveResponse, error) {
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
	response := api.JoinTNDAOApproveResponse{}

	// Data
	var wg errgroup.Group
	var rocketDAONodeTrustedActionsAddress *common.Address
	var rplBondAmount *big.Int

	// Get oracle node actions contract address
	wg.Go(func() error {
		var err error
		rocketDAONodeTrustedActionsAddress, err = rp.GetAddress("rocketDAONodeTrustedActions", nil)
		return err
	})

	// Get RPL bond amount
	wg.Go(func() error {
		var err error
		rplBondAmount, err = tnsettings.GetRPLBond(rp, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Approve RPL allowance
	hash, err := tokens.ApproveRPL(rp, *rocketDAONodeTrustedActionsAddress, rplBondAmount, opts)
	if err != nil {
		return nil, err
	}

	response.ApproveTxHash = hash

	// Return response
	return &response, nil

}

func waitForApprovalAndJoin(c *cli.Command, hash common.Hash, t *snroute.TransactOpts) (*api.JoinTNDAOJoinResponse, error) {
	opts := t.Opts()

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

	// Response
	response := api.JoinTNDAOJoinResponse{}

	// Join
	joinHash, err := tndao.Join(rp, opts)
	if err != nil {
		return nil, err
	}

	response.JoinTxHash = joinHash

	// Return response
	return &response, nil

}

func canJoinHandler(ctx snroute.Context) {
	resp, err := canJoin(ctx.Command())
	response.WriteResponse(ctx.Writer, resp, err)
}

func joinApproveRplHandler(ctx snroute.WriteContext) {
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := approveRpl(ctx.Command(), opts)
	response.WriteResponse(ctx.Writer, resp, err)
}

func joinHandler(ctx snroute.WriteContext) {
	hashStr := ctx.Request.FormValue("approvalTxHash")
	if hashStr == "" {
		response.WriteErrorResponse(ctx.Writer, fmt.Errorf("missing required parameter: approvalTxHash"))
		return
	}
	hash := common.HexToHash(hashStr)
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := waitForApprovalAndJoin(ctx.Command(), hash, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
