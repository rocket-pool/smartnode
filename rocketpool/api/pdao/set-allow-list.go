package pdao

import (
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
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

func proposeAllowListedControllers(c *cli.Command, addressList []common.Address, blockNumber uint32, opts *bind.TransactOpts) (*api.PDAOProposeAllowListedControllersResponse, error) {
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

func canProposeAllowListedControllersHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addressList := paramVal(r, "addressList")
		addresses, err := parseAddressList(r, "addressList")
		if err != nil {
			// Fall back to the raw comma-separated string if address parsing fails
			addresses = parseRawAddressList(addressList)
		}
		resp, err := canProposeAllowListedControllers(c, addresses)
		response.WriteResponse(w, resp, err)
	}
}

func proposeAllowListedControllersHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addressList := paramVal(r, "addressList")
		addresses, err := parseAddressList(r, "addressList")
		if err != nil {
			addresses = parseRawAddressList(addressList)
		}
		blockNumber, err := parseUint32Param(r, "blockNumber")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := proposeAllowListedControllers(c, addresses, blockNumber, opts)
		response.WriteResponse(w, resp, err)
	}
}
