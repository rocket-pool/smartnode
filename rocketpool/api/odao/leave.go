package odao

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/dao/trustednode"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canLeave(c *cli.Command) (*api.CanLeaveTNDAOResponse, error) {

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
	response := api.CanLeaveTNDAOResponse{}

	// Sync
	var wg errgroup.Group

	// Check proposal actionable status
	wg.Go(func() error {
		nodeAccount, err := w.GetNodeAccount()
		if err != nil {
			return err
		}
		proposalActionable, err := getProposalIsActionable(rp, nodeAccount.Address, "leave")
		if err == nil {
			response.ProposalExpired = !proposalActionable
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
		gasLimits, err := trustednode.EstimateLeaveGas(rp, opts.From, opts)
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
	response.CanLeave = !response.ProposalExpired && !response.InsufficientMembers
	return &response, nil

}

func leave(c *cli.Command, bondRefundAddress common.Address, opts *bind.TransactOpts) (*api.LeaveTNDAOResponse, error) {

	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.LeaveTNDAOResponse{}

	// Leave
	hash, err := trustednode.Leave(rp, bondRefundAddress, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canLeaveHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := canLeave(c)
		response.WriteResponse(w, resp, err)
	}
}

func leaveHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bondRefundStr := r.FormValue("bondRefundAddress")
		if bondRefundStr == "" {
			response.WriteErrorResponse(w, fmt.Errorf("missing required parameter: bondRefundAddress"))
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := leave(c, common.HexToAddress(bondRefundStr), opts)
		response.WriteResponse(w, resp, err)
	}
}
