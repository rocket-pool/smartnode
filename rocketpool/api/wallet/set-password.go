package wallet

import (
	"errors"
	"net/http"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func setPassword(c *cli.Command, password string) (*api.SetPasswordResponse, error) {

	// Get services
	pm, err := services.GetPasswordManager(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.SetPasswordResponse{}

	// Check if password is already set
	if pm.IsPasswordSet() {
		return nil, errors.New("The node password is already set")
	}

	// Set password
	if err := pm.SetPassword(password); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}

func setPasswordHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		password := r.FormValue("password")
		resp, err := setPassword(c, password)
		response.WriteResponse(w, resp, err)
	}
}
