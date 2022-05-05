package service

import (
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Gets the status of the configured execution clients
func getExecutionClientStatus(c *cli.Context) (*api.ExecutionClientStatusResponse, error) {

	// Get services
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ExecutionClientStatusResponse{}

	// Get the EC manager status
	mgrStatus := ec.CheckStatus()
	response.ManagerStatus = *mgrStatus

	// Return response
	return &response, nil

}
