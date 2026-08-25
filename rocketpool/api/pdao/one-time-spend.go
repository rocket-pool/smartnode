package pdao

import (
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canProposeOneTimeSpend(c *cli.Command, invoiceID string, recipient common.Address, amount *big.Int, customMessage string) (*api.PDAOCanProposeOneTimeSpendResponse, error) {
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
	response := api.PDAOCanProposeOneTimeSpendResponse{}

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

	// Try proposing
	blockNumber, pollard, err := createPollard(rp, cfg, bc)
	if err != nil {
		return nil, err
	}
	gasLimits, err := protocol.EstimateProposeOneTimeTreasurySpendGas(rp, customMessage, invoiceID, recipient, amount, blockNumber, pollard, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.BlockNumber = blockNumber
	response.GasLimits = gasLimits
	return &response, nil
}

func proposeOneTimeSpend(c *cli.Command, invoiceID string, recipient common.Address, amount *big.Int, blockNumber uint32, customMessage string, opts *bind.TransactOpts) (*api.PDAOProposeOneTimeSpendResponse, error) {
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
	response := api.PDAOProposeOneTimeSpendResponse{}

	// Propose
	pollard, err := getPollard(rp, cfg, bc, blockNumber)
	if err != nil {
		return nil, err
	}
	proposalID, hash, err := protocol.ProposeOneTimeTreasurySpend(rp, customMessage, invoiceID, recipient, amount, blockNumber, pollard, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.ProposalId = proposalID
	response.TxHash = hash
	return &response, nil
}

func canProposeOneTimeSpendHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		invoiceID, recipient, amount, customMessage, err := parseOneTimeSpendParams(r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeOneTimeSpend(c, invoiceID, recipient, amount, customMessage)
		response.WriteResponse(w, resp, err)
	}
}

func proposeOneTimeSpendHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		invoiceID, recipient, amount, customMessage, err := parseOneTimeSpendParams(r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
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
		resp, err := proposeOneTimeSpend(c, invoiceID, recipient, amount, blockNumber, customMessage, opts)
		response.WriteResponse(w, resp, err)
	}
}
