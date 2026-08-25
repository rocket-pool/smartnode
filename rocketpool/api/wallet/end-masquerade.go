package wallet

import (
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func endMasquerade(c *cli.Command) (*api.EndMasqueradeResponse, error) {

	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	err = w.EndMasquerade()
	if err != nil {
		return nil, fmt.Errorf("Error ending masquerade mode")
	}

	response := api.EndMasqueradeResponse{}

	return &response, nil
}

func endMasqueradeHandler(ctx snroute.WriteContext) {
	resp, err := endMasquerade(ctx.Command())
	response.WriteResponse(ctx.Writer, resp, err)
}
