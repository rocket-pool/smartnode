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

func sign(c *cli.Context, serializedTx string) (*api.NodeSignResponse, error) {

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

	serializedTx = hexutils.RemovePrefix(serializedTx)
	bytes, err := hex.DecodeString(serializedTx)
	if err != nil {
		return nil, fmt.Errorf("Error parsing TX bytes [%s]: %w", serializedTx, err)
	}

	signedBytes, err := w.Sign(bytes)
	if err != nil {
		return nil, fmt.Errorf("Error signing TX [%s]: %w", serializedTx, err)
	}
	response.SignedData = hexutils.AddPrefix(hex.EncodeToString(signedBytes))

	// Return response
	return &response, nil

}
