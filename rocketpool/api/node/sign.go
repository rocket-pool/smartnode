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

func sign(c *cli.Context, data string) (*api.NodeSignResponse, error) {

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

	data = hexutils.RemovePrefix(data)
	bytes, err := hex.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("Error parsing data [%s]: %w", data, err)
	}

	signedBytes, err := w.Sign(bytes)
	if err != nil {
		return nil, fmt.Errorf("Error signing data [%s]: %w", data, err)
	}
	response.SignedData = signedBytes

	// Return response
	return &response, nil

}
