package odao

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/settings/trustednode"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)


func getMemberSettings(c *cli.Context) (*api.GetTNDAOSettingMembersResponse, error) {

    // Get services
    if err := services.RequireNodeTrusted(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.GetTNDAOSettingMembersResponse{}

    quorum, err := trustednode.GetQuorum(rp, nil)
    if(err != nil) {
        return nil, fmt.Errorf("Error getting quorum: %w", err)
    }

    rplBond, err := trustednode.GetRPLBond(rp, nil)
    if(err != nil) {
        return nil, fmt.Errorf("Error getting RPL Bond: %w", err)
    }

    minipoolUnbondedMax, err := trustednode.GetMinipoolUnbondedMax(rp, nil)
    if(err != nil) {
        return nil, fmt.Errorf("Error getting minipool unbonded max: %w", err)
    }

    response.Quorum = quorum
    response.RPLBond = rplBond
    response.MinipoolUnbondedMax = minipoolUnbondedMax
    
    // Return response
    return &response, nil
}