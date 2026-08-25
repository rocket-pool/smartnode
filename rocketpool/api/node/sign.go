package node

import (
	"encoding/hex"
	"fmt"
	"net/http"
	_ "time/tzdata" // Must be imported somewhere for the embedded tz data to load

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	hexutils "github.com/rocket-pool/smartnode/shared/hex"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func sign(c *cli.Command, serializedTx string) (*api.NodeSignResponse, error) {

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

func signHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serializedTx := r.FormValue("serializedTx")
		resp, err := sign(c, serializedTx)
		response.WriteResponse(w, resp, err)
	}
}
