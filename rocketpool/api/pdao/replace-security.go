package pdao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/dao/security"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
)

func canProposeReplaceMemberOfSecurityCouncil(c *cli.Context, existingMemberAddress common.Address, newMemberID string, newMemberAddress common.Address) (*api.PDAOCanProposeReplaceMemberOfSecurityCouncilResponse, error) {
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
	response := api.PDAOCanProposeReplaceMemberOfSecurityCouncilResponse{}

	// Get node account
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Get the existing member
	existingID, err := security.GetMemberID(rp, existingMemberAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting ID of existing member: %w", err)
	}

	// Try proposing
	message := fmt.Sprintf("replace %s (%s) on the security council with %s (%s)", existingID, existingMemberAddress.Hex(), newMemberID, newMemberAddress.Hex())
	blockNumber, pollard, err := createPollard(rp, cfg, bc)
	if err != nil {
		return nil, err
	}
	gasInfo, err := protocol.EstimateProposeReplaceSecurityCouncilMemberGas(rp, message, existingMemberAddress, newMemberID, newMemberAddress, blockNumber, pollard, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.BlockNumber = blockNumber
	response.GasInfo = gasInfo
	return &response, nil
}

func proposeReplaceMemberOfSecurityCouncil(c *cli.Context, existingMemberAddress common.Address, newMemberID string, newMemberAddress common.Address, blockNumber uint32) (*api.PDAOProposeReplaceMemberOfSecurityCouncilResponse, error) {
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
	response := api.PDAOProposeReplaceMemberOfSecurityCouncilResponse{}

	// Get node account
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Get the existing member
	existingID, err := security.GetMemberID(rp, existingMemberAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting ID of existing member: %w", err)
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Propose
	message := fmt.Sprintf("replace %s (%s) on the security council with %s (%s)", existingID, existingMemberAddress.Hex(), newMemberID, newMemberAddress.Hex())
	pollard, err := getPollard(rp, cfg, bc, blockNumber)
	if err != nil {
		return nil, err
	}
	proposalID, hash, err := protocol.ProposeReplaceSecurityCouncilMember(rp, message, existingMemberAddress, newMemberID, newMemberAddress, blockNumber, pollard, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.ProposalId = proposalID
	response.TxHash = hash
	return &response, nil
}
