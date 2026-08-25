package security

import (
	"net/http"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/dao/security"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getMembers(c *cli.Command) (*api.SecurityMembersResponse, error) {

	// Get services
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.SecurityMembersResponse{}

	// Get members
	members, err := security.GetMembers(rp, nil)
	if err != nil {
		return nil, err
	}
	response.Members = members

	// Return response
	return &response, nil

}

func membersHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := getMembers(c)
		response.WriteResponse(w, resp, err)
	}
}
