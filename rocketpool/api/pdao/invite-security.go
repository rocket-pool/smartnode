package pdao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/dao/security"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

func canProposeInviteToSecurityCouncil(c *cli.Context, id string, address common.Address) (*api.PDAOCanProposeInviteToSecurityCouncilResponse, error) {
	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.PDAOCanProposeInviteToSecurityCouncilResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Sync
	var isRplLockingAllowed bool
	var wg errgroup.Group

	// Get is RPL locking allowed
	wg.Go(func() error {
		var err error
		isRplLockingAllowed, err = node.GetRPLLockedAllowed(rp, nodeAccount.Address, nil)
		return err
	})

	// Check if the member exists
	wg.Go(func() error {
		var err error
		response.MemberAlreadyExists, err = security.GetMemberExists(rp, address, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Update & return response
	response.IsRplLockingDisallowed = !isRplLockingAllowed

	// return if proposing is not possible
	response.CanPropose = !(response.MemberAlreadyExists || response.IsRplLockingDisallowed)
	if !response.CanPropose {
		return &response, nil
	}

	// Get the account transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Try proposing
	message := fmt.Sprintf("invite %s (%s) to the security council", id, address.Hex())
	blockNumber, pollard, err := createPollard(rp, cfg, bc)
	if err != nil {
		return nil, err
	}
	gasInfo, err := protocol.EstimateProposeInviteToSecurityCouncilGas(rp, message, id, address, blockNumber, pollard, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.BlockNumber = blockNumber
	response.GasInfo = gasInfo
	return &response, nil
}

func proposeInviteToSecurityCouncil(c *cli.Context, id string, address common.Address, blockNumber uint32) (*api.PDAOProposeInviteToSecurityCouncilResponse, error) {
	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.PDAOProposeInviteToSecurityCouncilResponse{}

	// Get node account
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Propose
	message := fmt.Sprintf("invite %s (%s) to the security council", id, address.Hex())
	pollard, err := getPollard(rp, cfg, bc, blockNumber)
	if err != nil {
		return nil, err
	}
	proposalID, hash, err := protocol.ProposeInviteToSecurityCouncil(rp, message, id, address, blockNumber, pollard, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.ProposalId = proposalID
	response.TxHash = hash
	return &response, nil
}
