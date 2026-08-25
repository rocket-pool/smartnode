package wallet

import (
	"encoding/hex"
	"net/http"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func exportWallet(c *cli.Command) (*api.ExportWalletResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	pm, err := services.GetPasswordManager(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ExportWalletResponse{}

	// Get password
	password, err := pm.GetPassword()
	if err != nil {
		return nil, err
	}
	response.Password = password

	// Serialize wallet
	wallet, err := w.String()
	if err != nil {
		return nil, err
	}
	response.Wallet = wallet

	// Get account private key
	privateKey, err := w.GetNodePrivateKeyBytes()
	if err != nil {
		return nil, err
	}
	response.AccountPrivateKey = hex.EncodeToString(privateKey)

	// Return response
	return &response, nil

}

func exportHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := exportWallet(c)
		response.WriteResponse(w, resp, err)
	}
}
