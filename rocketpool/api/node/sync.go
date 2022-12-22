package node

import (
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getSyncProgress(c *cli.Context) (*api.NodeSyncProgressResponse, error) {

	// Response
	response := api.NodeSyncProgressResponse{}

	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Get the EC manager
	ecMgr, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}

	// Get the status of the EC and fallback EC
	ecStatus := ecMgr.CheckStatus(cfg)
	response.EcStatus = *ecStatus

	// Get the BC manager
	bcMgr, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Get the status of the BC and fallback BC
	bcStatus := bcMgr.CheckStatus()
	response.BcStatus = *bcStatus

	// Return response
	return &response, nil

}
