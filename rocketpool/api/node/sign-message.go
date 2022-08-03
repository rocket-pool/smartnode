package node

import (
	"encoding/hex"
	"fmt"
	_ "time/tzdata"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	hexutils "github.com/rocket-pool/smartnode/shared/utils/hex"
)

func signMessage(c *cli.Context) (*api.NodeSignResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeSignResponse{}
	signedBytes, err := w.SignMessage(c.String("message"))
	if err != nil {
		return nil, fmt.Errorf("Error signing message [%s]: %w", message, err)
	}
	response.SignedData = hexutils.AddPrefix(hex.EncodeToString(signedBytes))

	// Return response
	return &response, nil

}
