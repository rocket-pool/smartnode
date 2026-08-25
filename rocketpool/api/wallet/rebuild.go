package wallet

import (
	"net/http"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func rebuildWallet(c *cli.Command) (*api.RebuildWalletResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.RebuildWalletResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Recover validator keys
	response.ValidatorKeys, err = recoverNodeKeys(c, rp, bc, nodeAccount.Address, w, false)
	if err != nil {
		return nil, err
	}

	// Save wallet
	if err := w.Save(); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}

func rebuildHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := withRecoveryLock("wallet rebuild", func() (*api.RebuildWalletResponse, error) {
			return rebuildWallet(c)
		})
		response.WriteResponse(w, resp, err)
	}
}
