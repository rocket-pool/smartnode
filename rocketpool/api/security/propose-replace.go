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

func canProposeReplaceMember(c *cli.Context, existingAddress common.Address, newID string, newAddress common.Address) (*api.SecurityCanProposeReplaceResponse, error) {
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
	response := api.SecurityCanProposeReplaceResponse{}

	// Sync
	var wg errgroup.Group
	var oldMemberExists bool
	var oldMemberID string
	var newMemberExists bool

	// Check if old member exists
	wg.Go(func() error {
		var err error
		oldMemberExists, err = security.GetMemberExists(rp, existingAddress, nil)
		return err
	})

	// Get the member ID
	wg.Go(func() error {
		var err error
		oldMemberID, err = security.GetMemberID(rp, existingAddress, nil)
		return err
	})

	// Check if new member exists
	wg.Go(func() error {
		var err error
		newMemberExists, err = security.GetMemberExists(rp, newAddress, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	response.OldMemberDoesntExist = !oldMemberExists
	response.NewMemberAlreadyExists = newMemberExists
	response.CanPropose = oldMemberExists && !newMemberExists
	if !response.CanPropose {
		return &response, nil
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	message := fmt.Sprintf("replace %s (%s) with %s (%s)", oldMemberID, existingAddress.Hex(), newID, newAddress.Hex())
	gasInfo, err := security.EstimateProposeReplaceGas(rp, message, existingAddress, newID, newAddress, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.GasInfo = gasInfo
	return &response, nil
}

func proposeReplaceMember(c *cli.Context, existingAddress common.Address, newID string, newAddress common.Address) (*api.SecurityProposeInviteResponse, error) {
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
	response := api.SecurityProposeInviteResponse{}

	// Get the old ID
	oldMemberID, err := security.GetMemberID(rp, existingAddress, nil)
	if err != nil {
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
	message := fmt.Sprintf("replace %s (%s) with %s (%s)", oldMemberID, existingAddress.Hex(), newID, newAddress.Hex())
	proposalId, hash, err := security.ProposeReplace(rp, message, existingAddress, newID, newAddress, opts)
	if err != nil {
		return nil, err
	}
	response.ProposalId = proposalId
	response.TxHash = hash

	// Return response
	return &response, nil
}
