package node

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	hexutils "github.com/rocket-pool/smartnode/shared/hex"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
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

func signMessageHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		message := r.FormValue("message")
		resp, err := signMessage(c, message)
		response.WriteResponse(w, resp, err)
	}
}
