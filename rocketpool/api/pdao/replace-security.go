package pdao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/dao/security"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canProposeReplaceMemberOfSecurityCouncil(c *cli.Command, existingMemberAddress common.Address, newMemberID string, newMemberAddress common.Address) (*api.PDAOCanProposeReplaceMemberOfSecurityCouncilResponse, error) {
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
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Sync
	var isRplLockingAllowed bool

	// Get is RPL locking allowed
	isRplLockingAllowed, err = node.GetRPLLockedAllowed(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.IsRplLockingDisallowed = !isRplLockingAllowed

	// return if proposing is not possible
	response.CanPropose = !response.IsRplLockingDisallowed
	if !response.CanPropose {
		return &response, nil
	}

	// Get the account transactor
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
	gasLimits, err := protocol.EstimateProposeReplaceSecurityCouncilMemberGas(rp, message, existingMemberAddress, newMemberID, newMemberAddress, blockNumber, pollard, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.BlockNumber = blockNumber
	response.GasLimits = gasLimits
	return &response, nil
}

func proposeReplaceMemberOfSecurityCouncil(c *cli.Command, existingMemberAddress common.Address, newMemberID string, newMemberAddress common.Address, blockNumber uint32, t *snroute.TransactOpts) (*api.PDAOProposeReplaceMemberOfSecurityCouncilResponse, error) {
	opts := t.Opts()

	// Get services
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
	response := api.PDAOProposeReplaceMemberOfSecurityCouncilResponse{}

	// Get node account
	// Get the existing member
	existingID, err := security.GetMemberID(rp, existingMemberAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting ID of existing member: %w", err)
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

func canProposeReplaceMemberOfSecurityCouncilHandler(ctx snroute.Context) {
	existing := common.HexToAddress(paramVal(ctx.Request, "existingAddress"))
	newID := paramVal(ctx.Request, "newId")
	newAddr := common.HexToAddress(paramVal(ctx.Request, "newAddress"))
	resp, err := canProposeReplaceMemberOfSecurityCouncil(ctx.Command(), existing, newID, newAddr)
	response.WriteResponse(ctx.Writer, resp, err)
}

func proposeReplaceMemberOfSecurityCouncilHandler(ctx snroute.WriteContext) {
	existing := common.HexToAddress(paramVal(ctx.Request, "existingAddress"))
	newID := paramVal(ctx.Request, "newId")
	newAddr := common.HexToAddress(paramVal(ctx.Request, "newAddress"))
	blockNumber, err := parseUint32Param(ctx.Request, "blockNumber")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := proposeReplaceMemberOfSecurityCouncil(ctx.Command(), existing, newID, newAddr, blockNumber, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
