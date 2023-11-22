package security

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/security"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canProposeKick(c *cli.Context, memberAddress common.Address) (*api.SecurityCanProposeKickResponse, error) {

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
	response := api.SecurityCanProposeKickResponse{}

	// Sync
	var wg errgroup.Group

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		memberId, err := security.GetMemberID(rp, memberAddress, nil)
		if err != nil {
			return err
		}
		message := fmt.Sprintf("kick %s (%s)", memberId, memberAddress.Hex())
		gasInfo, err := security.EstimateProposeKickGas(rp, message, memberAddress, opts)
		if err == nil {
			response.GasInfo = gasInfo
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Update & return response
	response.CanPropose = true
	return &response, nil

}

func proposeKick(c *cli.Context, memberAddress common.Address) (*api.SecurityProposeKickResponse, error) {

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
	response := api.SecurityProposeKickResponse{}

	// Data
	var wg errgroup.Group
	var memberId string

	// Get member details
	wg.Go(func() error {
		var err error
		memberId, err = security.GetMemberID(rp, memberAddress, nil)
		return err
	})
	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

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
	message := fmt.Sprintf("kick %s (%s)", memberId, memberAddress.Hex())
	proposalId, hash, err := security.ProposeKick(rp, message, memberAddress, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil

}
