package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
	"github.com/urfave/cli/v3"
)

// Restarts the Validator client
func restartVc(c *cli.Command) (*api.RestartVcResponse, error) {

	// Get services
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	d, err := services.GetDocker(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.RestartVcResponse{}

	if err := validator.RestartValidator(cfg, bc, nil, d); err != nil {
		return nil, fmt.Errorf("error restarting validator client: %w", err)
	}

	// Return response
	return &response, nil

}
