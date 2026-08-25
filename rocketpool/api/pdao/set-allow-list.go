package pdao

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canProposeAllowListedControllers(c *cli.Command, addressList []common.Address) (*api.PDAOACanProposeAllowListedControllersResponse, error) {
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
	response := api.PDAOACanProposeAllowListedControllersResponse{}

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

	// Get node account
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Try proposing
	blockNumber, pollard, err := createPollard(rp, cfg, bc)
	if err != nil {
		return nil, err
	}
	gasLimits, err := protocol.EstimateProposeAllowListedControllersGas(rp, addressList, blockNumber, pollard, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.BlockNumber = blockNumber
	response.GasLimits = gasLimits
	return &response, nil
}

func proposeAllowListedControllers(c *cli.Command, addressList []common.Address, blockNumber uint32, t *snroute.TransactOpts) (*api.PDAOProposeAllowListedControllersResponse, error) {
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
	response := api.PDAOProposeAllowListedControllersResponse{}

	// Propose
	pollard, err := getPollard(rp, cfg, bc, blockNumber)
	if err != nil {
		return nil, err
	}

	proposalID, hash, err := protocol.ProposeAllowListedControllers(rp, addressList, blockNumber, pollard, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.ProposalId = proposalID
	response.TxHash = hash
	return &response, nil
}

func canProposeAllowListedControllersHandler(ctx snroute.Context) {
	addressList := paramVal(ctx.Request, "addressList")
	addresses, err := parseAddressList(ctx.Request, "addressList")
	if err != nil {
		// Fall back to the raw comma-separated string if address parsing fails
		addresses = parseRawAddressList(addressList)
	}
	resp, err := canProposeAllowListedControllers(ctx.Command(), addresses)
	response.WriteResponse(ctx.Writer, resp, err)
}

func proposeAllowListedControllersHandler(ctx snroute.WriteContext) {
	addressList := paramVal(ctx.Request, "addressList")
	addresses, err := parseAddressList(ctx.Request, "addressList")
	if err != nil {
		addresses = parseRawAddressList(addressList)
	}
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
	resp, err := proposeAllowListedControllers(ctx.Command(), addresses, blockNumber, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
