package auction

import (
    "math/big"

    "github.com/rocket-pool/rocketpool-go/auction"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canBidOnLot(c *cli.Context, lotIndex uint64) (*api.CanBidOnLotResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanBidOnLotResponse{}



    // Update & return response
    response.CanBid = !(response.DoesNotExist || response.BiddingEnded || response.RPLExhausted || response.BidOnLotDisabled)
    return &response, nil

}


func bidOnLot(c *cli.Context, lotIndex uint64, amountWei *big.Int) (*api.BidOnLotResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.BidOnLotResponse{}



    // Return response
    return &response, nil

}

