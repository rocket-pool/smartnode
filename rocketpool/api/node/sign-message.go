package node

import (
	"encoding/hex"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	hexutils "github.com/rocket-pool/smartnode/shared/utils/hex"
)

func signMessage(c *cli.Command, message string) (*api.NodeSignResponse, error) {
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeSignResponse{}
	signedBytes, err := w.SignMessage(message)
	if err != nil {
		return nil, fmt.Errorf("Error signing message [%s]: %w", message, err)
	}
	response.SignedData = hexutils.AddPrefix(hex.EncodeToString(signedBytes))

	// Return response
	return &response, nil

}
