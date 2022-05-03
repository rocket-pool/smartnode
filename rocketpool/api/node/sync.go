package node

import (
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getSyncProgress(c *cli.Context) (*api.NodeSyncProgressResponse, error) {

	// Response
	response := api.NodeSyncProgressResponse{}

	// Get EC manager
	ecMgr, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}

	// Get status of EC and fallback EC
	status := ecMgr.CheckStatus()
	response.EcStatus = *status

	// Get CC client
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Get CC sync progress
	syncStatus, err := bc.GetSyncStatus()
	if err != nil {
		return nil, err
	}
	if syncStatus.Syncing {
		response.Eth2Progress = syncStatus.Progress
		response.Eth2Synced = false
	} else {
		response.Eth2Progress = 1
		response.Eth2Synced = true
	}

	// Return response
	return &response, nil

}
