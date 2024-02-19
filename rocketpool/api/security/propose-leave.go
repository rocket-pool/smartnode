package security

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/dao/security"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canProposeLeave(c *cli.Context) (*api.SecurityCanProposeLeaveResponse, error) {

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
	gasInfo, err := security.EstimateRequestLeaveGas(rp, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.GasInfo = gasInfo
	return &response, nil

}

func proposeLeave(c *cli.Context) (*api.SecurityProposeLeaveResponse, error) {

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
	response := api.SecurityProposeLeaveResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit proposal
	hash, err := security.RequestLeave(rp, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
