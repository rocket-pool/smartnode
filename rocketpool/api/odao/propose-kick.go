package odao

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canProposeKick(c *cli.Command, memberAddress common.Address, fineAmountWei *big.Int) (*api.CanProposeTNDAOKickResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
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
	response := api.CanProposeTNDAOKickResponse{}

	// Sync
	var wg errgroup.Group

	// Check if proposal cooldown is active
	wg.Go(func() error {
		nodeAccount, err := w.GetNodeAccount()
		if err != nil {
			return err
		}
		proposalCooldownActive, err := getProposalCooldownActive(rp, nodeAccount.Address)
		if err == nil {
			response.ProposalCooldownActive = proposalCooldownActive
		}
		return err
	})

	// Check member's RPL bond amount
	wg.Go(func() error {
		rplBondAmount, err := trustednode.GetMemberRPLBondAmount(rp, memberAddress, nil)
		if err == nil {
			response.InsufficientRplBond = (fineAmountWei.Cmp(rplBondAmount) > 0)
		}
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		memberId, err := trustednode.GetMemberID(rp, memberAddress, nil)
		if err != nil {
			return err
		}
		memberUrl, err := trustednode.GetMemberUrl(rp, memberAddress, nil)
		if err != nil {
			return err
		}
		message := fmt.Sprintf("kick %s (%s) with %.6f RPL fine", memberId, memberUrl, math.RoundDown(math.WeiToEth(fineAmountWei), 6))
		gasLimits, err := trustednode.EstimateProposeKickMemberGas(rp, message, memberAddress, fineAmountWei, opts)
		if err == nil {
			response.GasLimits = gasLimits
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Update & return response
	response.CanPropose = !response.ProposalCooldownActive && !response.InsufficientRplBond
	return &response, nil

}

func proposeKick(c *cli.Command, memberAddress common.Address, fineAmountWei *big.Int, t *snroute.TransactOpts) (*api.ProposeTNDAOKickResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProposeTNDAOKickResponse{}

	// Data
	var wg errgroup.Group
	var memberId string
	var memberUrl string

	// Get member details
	wg.Go(func() error {
		var err error
		memberId, err = trustednode.GetMemberID(rp, memberAddress, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		memberUrl, err = trustednode.GetMemberUrl(rp, memberAddress, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Submit proposal
	message := fmt.Sprintf("kick %s (%s) with %.6f RPL fine", memberId, memberUrl, math.RoundDown(math.WeiToEth(fineAmountWei), 6))
	proposalId, hash, err := trustednode.ProposeKickMember(rp, message, memberAddress, fineAmountWei, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canProposeKickHandler(ctx snroute.Context) {
	addr, fine, err := parseKickParams(ctx.Request)
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canProposeKick(ctx.Command(), addr, fine)
	response.WriteResponse(ctx.Writer, resp, err)
}

func proposeKickHandler(ctx snroute.WriteContext) {
	addr, fine, err := parseKickParams(ctx.Request)
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := proposeKick(ctx.Command(), addr, fine, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
