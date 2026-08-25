package service

import (
	"net/http"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Gets the status of the configured Execution clients
func getClientStatus(c *cli.Command) (*api.ClientStatusResponse, error) {

	// Get services
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ClientStatusResponse{}

	// Get the EC manager status
	ecMgrStatus := ec.CheckStatus(cfg)
	response.EcManagerStatus = *ecMgrStatus

	// Get the BC manager status
	bcMgrStatus := bc.CheckStatus()
	response.BcManagerStatus = *bcMgrStatus

	// Return response
	return &response, nil

}

func getClientStatusHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := getClientStatus(c)
		response.WriteResponse(w, resp, err)
	}
}
