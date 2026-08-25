package pdao

import (
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canProposeRecurringSpendUpdate(c *cli.Command, contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, numberOfPeriods uint64, customMessage string) (*api.PDAOCanProposeRecurringSpendResponse, error) {
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
	response := api.PDAOCanProposeRecurringSpendResponse{}

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
	gasLimits, err := protocol.EstimateProposeRecurringTreasurySpendUpdateGas(rp, customMessage, contractName, recipient, amountPerPeriod, periodLength, numberOfPeriods, blockNumber, pollard, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.BlockNumber = blockNumber
	response.GasLimits = gasLimits
	return &response, nil
}

func proposeRecurringSpendUpdate(c *cli.Command, contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, numberOfPeriods uint64, blockNumber uint32, customMessage string, opts *bind.TransactOpts) (*api.PDAOProposeOneTimeSpendResponse, error) {
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
	proposalID, hash, err := protocol.ProposeRecurringTreasurySpendUpdate(rp, customMessage, contractName, recipient, amountPerPeriod, periodLength, numberOfPeriods, blockNumber, pollard, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.ProposalId = proposalID
	response.TxHash = hash
	return &response, nil
}

func canProposeRecurringSpendUpdateHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contractName, recipient, amountPerPeriod, periodLength, _, numberOfPeriods, customMessage, err := parseRecurringSpendParams(r, true)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeRecurringSpendUpdate(c, contractName, recipient, amountPerPeriod, periodLength, numberOfPeriods, customMessage)
		response.WriteResponse(w, resp, err)
	}
}

func proposeRecurringSpendUpdateHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contractName, recipient, amountPerPeriod, periodLength, _, numberOfPeriods, customMessage, err := parseRecurringSpendParams(r, true)
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
		resp, err := proposeRecurringSpendUpdate(c, contractName, recipient, amountPerPeriod, periodLength, numberOfPeriods, blockNumber, customMessage, opts)
		response.WriteResponse(w, resp, err)
	}
}
