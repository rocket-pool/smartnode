package odao

import (
	"fmt"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canProposeLeave(c *cli.Command) (*api.CanProposeTNDAOLeaveResponse, error) {

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
	response := api.CanProposeTNDAOLeaveResponse{}

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

	// Check if members can leave the oracle DAO
	wg.Go(func() error {
		membersCanLeave, err := getMembersCanLeave(rp)
		if err == nil {
			response.InsufficientMembers = !membersCanLeave
		}
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		nodeAccount, err := w.GetNodeAccount()
		if err != nil {
			return err
		}
		nodeMemberId, err := trustednode.GetMemberID(rp, nodeAccount.Address, nil)
		if err != nil {
			return err
		}
		nodeMemberUrl, err := trustednode.GetMemberUrl(rp, nodeAccount.Address, nil)
		if err != nil {
			return err
		}
		message := fmt.Sprintf("%s (%s) leaves", nodeMemberId, nodeMemberUrl)
		gasLimits, err := trustednode.EstimateProposeMemberLeaveGas(rp, message, nodeAccount.Address, opts)
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
	response.CanPropose = !response.ProposalCooldownActive && !response.InsufficientMembers
	return &response, nil

}

func proposeLeave(c *cli.Command, t *snroute.TransactOpts) (*api.ProposeTNDAOLeaveResponse, error) {
	opts := t.Opts()

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
	response := api.ProposeTNDAOLeaveResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Data
	var wg errgroup.Group
	var nodeMemberId string
	var nodeMemberUrl string

	// Get node member details
	wg.Go(func() error {
		var err error
		nodeMemberId, err = trustednode.GetMemberID(rp, nodeAccount.Address, nil)
		return err
	})
	wg.Go(func() error {
		var err error
		nodeMemberUrl, err = trustednode.GetMemberUrl(rp, nodeAccount.Address, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Submit proposal
	message := fmt.Sprintf("%s (%s) leaves", nodeMemberId, nodeMemberUrl)
	proposalId, hash, err := trustednode.ProposeMemberLeave(rp, message, nodeAccount.Address, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canProposeLeaveHandler(ctx snroute.Context) {
	resp, err := canProposeLeave(ctx.Command())
	response.WriteResponse(ctx.Writer, resp, err)
}

func proposeLeaveHandler(ctx snroute.WriteContext) {
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := proposeLeave(ctx.Command(), opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
