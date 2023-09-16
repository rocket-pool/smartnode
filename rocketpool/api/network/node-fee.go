package network

import (
	"fmt"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/settings"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getNodeFee(c *cli.Context) (*api.NetworkNodeFeeData, error) {
	// Get services
	if err := services.RequireEthClientSynced(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NetworkNodeFeeData{}

	// Create bindings
	network, err := network.NewNetworkFees(rp)
	if err != nil {
		return nil, fmt.Errorf("error getting network fees binding: %w", err)
	}
	pSettings, err := settings.NewProtocolDaoSettings(rp)
	if err != nil {
		return nil, fmt.Errorf("error getting protocol DAO settings binding: %w", err)
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		network.GetNodeFee(mc)
		pSettings.GetMinimumNodeFee(mc)
		pSettings.GetTargetNodeFee(mc)
		pSettings.GetMaximumNodeFee(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Update the response
	response.NodeFee = network.Details.NodeFee.Formatted()
	response.MinNodeFee = pSettings.Details.Network.MinimumNodeFee.Formatted()
	response.TargetNodeFee = pSettings.Details.Network.TargetNodeFee.Formatted()
	response.MaxNodeFee = pSettings.Details.Network.MaximumNodeFee.Formatted()
	return &response, nil
}
