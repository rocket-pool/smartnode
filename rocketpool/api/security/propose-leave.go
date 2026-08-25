package security

import (
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/dao/security"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canProposeLeave(c *cli.Command) (*api.SecurityCanProposeLeaveResponse, error) {

	// Get services
	if err := services.RequireNodeSecurityMember(c); err != nil {
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
	response := api.SecurityCanProposeLeaveResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Check if the member exists
	exists, err := security.GetMemberExists(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	response.MemberDoesntExist = !exists

	// Check validity
	response.CanPropose = !(response.MemberDoesntExist)
	if !response.CanPropose {
		return &response, nil
	}

	// Simulate the tx
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasLimits, err := security.EstimateRequestLeaveGas(rp, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.GasLimits = gasLimits
	return &response, nil

}

func proposeLeave(c *cli.Command, opts *bind.TransactOpts) (*api.SecurityProposeLeaveResponse, error) {

	// Get services
	if err := services.RequireNodeSecurityMember(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.SecurityProposeLeaveResponse{}

	// Submit proposal
	hash, err := security.RequestLeave(rp, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canProposeLeaveHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := canProposeLeave(c)
		response.WriteResponse(w, resp, err)
	}
}

func proposeLeaveHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeLeave(c, opts)
		response.WriteResponse(w, resp, err)
	}
}
