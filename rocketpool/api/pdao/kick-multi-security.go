package pdao

import (
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canProposeKickMultiFromSecurityCouncil(c *cli.Command, addresses []common.Address) (*api.PDAOCanProposeKickMultiFromSecurityCouncilResponse, error) {
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
	response := api.PDAOCanProposeKickMultiFromSecurityCouncilResponse{}

	// Get node account
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Try proposing
	message := "kick multiple members from the security council"
	blockNumber, pollard, err := createPollard(rp, cfg, bc)
	if err != nil {
		return nil, err
	}
	gasLimits, err := protocol.EstimateProposeKickMultiFromSecurityCouncilGas(rp, message, addresses, blockNumber, pollard, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.BlockNumber = blockNumber
	response.GasLimits = gasLimits
	return &response, nil
}

func proposeKickMultiFromSecurityCouncil(c *cli.Command, addresses []common.Address, blockNumber uint32, opts *bind.TransactOpts) (*api.PDAOProposeKickMultiFromSecurityCouncilResponse, error) {
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
	response := api.PDAOProposeKickMultiFromSecurityCouncilResponse{}

	// Propose
	message := "kick multiple members from the security council"
	pollard, err := getPollard(rp, cfg, bc, blockNumber)
	if err != nil {
		return nil, err
	}
	proposalID, hash, err := protocol.ProposeKickMultiFromSecurityCouncil(rp, message, addresses, blockNumber, pollard, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.ProposalId = proposalID
	response.TxHash = hash
	return &response, nil
}

func canProposeKickMultiFromSecurityCouncilHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addresses, err := parseAddressList(r, "addresses")
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := canProposeKickMultiFromSecurityCouncil(c, addresses)
		response.WriteResponse(w, resp, err)
	}
}

func proposeKickMultiFromSecurityCouncilHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addresses, err := parseAddressList(r, "addresses")
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
		resp, err := proposeKickMultiFromSecurityCouncil(c, addresses, blockNumber, opts)
		response.WriteResponse(w, resp, err)
	}
}
