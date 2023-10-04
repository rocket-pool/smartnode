package pdao

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	psettings "github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
)

func getRewardsPercentages(c *cli.Context) (*api.PDAOGetRewardsPercentagesResponse, error) {
	// Get services
	if err := services.RequireEthClientSynced(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.PDAOGetRewardsPercentagesResponse{}

	// Get the percentages
	rewardsPercents, err := psettings.GetRewardsPercentages(rp, nil)
	if err != nil {
		return nil, err
	}

	// Return them
	response.Node = rewardsPercents.NodePercentage
	response.OracleDao = rewardsPercents.OdaoPercentage
	response.ProtocolDao = rewardsPercents.PdaoPercentage
	return &response, nil
}

func canProposeRewardsPercentages(c *cli.Context, node *big.Int, odao *big.Int, pdao *big.Int) (*api.PDAOCanProposeRewardsPercentagesResponse, error) {
	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
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
	response := api.PDAOCanProposeRewardsPercentagesResponse{}

	// Get the account transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Get the latest finalized block number and corresponding pollard
	blockNumber, pollard, encodedPollard, err := createPollard(rp, cfg, bc)
	if err != nil {
		return nil, fmt.Errorf("error creating pollard: %w", err)
	}
	response.BlockNumber = blockNumber
	response.Pollard = encodedPollard

	gasInfo, err := protocol.EstimateProposeSetRewardsPercentageGas(rp, "update RPL rewards distribution", odao, pdao, node, blockNumber, pollard, opts)
	response.GasInfo = gasInfo

	// Return response
	return &response, nil
}

func proposeRewardsPercentages(c *cli.Context, node *big.Int, odao *big.Int, pdao *big.Int, blockNumber uint32, pollard string) (*api.PDAOProposeRewardsPercentagesResponse, error) {
	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
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
	response := api.PDAOProposeRewardsPercentagesResponse{}

	// Get the account transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Decode the pollard
	truePollard, err := decodePollard(pollard)
	if err != nil {
		return nil, fmt.Errorf("error regenerating pollard: %w", err)
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Submit the proposal
	proposalID, hash, err := protocol.ProposeSetRewardsPercentage(rp, "update RPL rewards distribution", odao, pdao, node, blockNumber, truePollard, opts)
	response.ProposalId = proposalID
	response.TxHash = hash

	// Return response
	return &response, nil
}
