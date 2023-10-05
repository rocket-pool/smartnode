package pdao

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
)

func canProposeRecurringSpend(c *cli.Context, contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, startTime time.Time, numberOfPeriods uint64) (*api.PDAOCanProposeRecurringSpendResponse, error) {
	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
		return nil, err
	}
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
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Try proposing
	message := fmt.Sprintf("recurring payment to %s", contractName)
	blockNumber, pollard, encodedPollard, err := createPollard(rp, cfg, bc)
	if err != nil {
		return nil, err
	}
	gasInfo, err := protocol.EstimateProposeRecurringTreasurySpendGas(rp, message, contractName, recipient, amountPerPeriod, periodLength, startTime, numberOfPeriods, blockNumber, pollard, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.BlockNumber = blockNumber
	response.Pollard = encodedPollard
	response.GasInfo = gasInfo
	return &response, nil
}

func proposeRecurringSpend(c *cli.Context, contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, startTime time.Time, numberOfPeriods uint64, blockNumber uint32, pollard string) (*api.PDAOProposeOneTimeSpendResponse, error) {
	// Get services
	if err := services.RequireNodeTrusted(c); err != nil {
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
	message := fmt.Sprintf("recurring payment to %s", contractName)
	truePollard, err := decodePollard(pollard)
	if err != nil {
		return nil, err
	}
	proposalID, hash, err := protocol.ProposeRecurringTreasurySpend(rp, message, contractName, recipient, amountPerPeriod, periodLength, startTime, numberOfPeriods, blockNumber, truePollard, opts)
	if err != nil {
		return nil, err
	}

	// Update & return response
	response.ProposalId = proposalID
	response.TxHash = hash
	return &response, nil
}
