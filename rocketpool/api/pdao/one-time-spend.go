package pdao

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
)

func canProposeOneTimeSpend(c *cli.Context, invoiceID string, recipient common.Address, amount *big.Int) (*api.PDAOCanProposeOneTimeSpendResponse, error) {
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
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Try proposing
	message := fmt.Sprintf("one-time spend for invoice %s", invoiceID)
	blockNumber, pollard, err := createPollard(rp, cfg, bc)
	if err != nil {
		return nil, err
	}
	gasInfo, err := protocol.EstimateProposeOneTimeTreasurySpendGas(rp, message, invoiceID, recipient, amount, blockNumber, pollard, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.BlockNumber = blockNumber
	response.GasInfo = gasInfo
	return &response, nil
}

func proposeOneTimeSpend(c *cli.Context, invoiceID string, recipient common.Address, amount *big.Int, blockNumber uint32) (*api.PDAOProposeOneTimeSpendResponse, error) {
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
	response := api.PDAOProposeOneTimeSpendResponse{}

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
	message := fmt.Sprintf("one-time spend for invoice %s", invoiceID)
	pollard, err := getPollard(rp, cfg, bc, blockNumber)
	if err != nil {
		return nil, err
	}
	proposalID, hash, err := protocol.ProposeOneTimeTreasurySpend(rp, message, invoiceID, recipient, amount, blockNumber, pollard, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.ProposalId = proposalID
	response.TxHash = hash
	return &response, nil
}
