package auction

import (
    "github.com/rocket-pool/rocketpool-go/auction"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canClaimFromLot(c *cli.Context, lotIndex uint64) (*api.CanClaimFromLotResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanClaimFromLotResponse{}



    // Update & return response
    response.CanClaim = !(response.DoesNotExist || response.NoBidFromAddress || response.NotCleared)
    return &response, nil

}


func claimFromLot(c *cli.Context, lotIndex uint64) (*api.ClaimFromLotResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.ClaimFromLotResponse{}



    // Return response
    return &response, nil

}

