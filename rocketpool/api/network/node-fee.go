package network

import (
    "github.com/rocket-pool/rocketpool-go/network"
    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)

// Config
const SuggestedNodeFeeDelta = -0.01 // 1% below current

func getNodeFee(c *cli.Context) (*api.NodeFeeResponse, error) {

    // Get services
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.NodeFeeResponse{}

    // Sync
    var wg errgroup.Group

    // Get data
    wg.Go(func() error {
        nodeFee, err := network.GetNodeFee(rp, nil)
        if err == nil {
            response.NodeFee = nodeFee
        }
        return err
    })
    wg.Go(func() error {
        minNodeFee, err := settings.GetMinimumNodeFee(rp, nil)
        if err == nil {
            response.MinNodeFee = minNodeFee
        }
        return err
    })
    wg.Go(func() error {
        targetNodeFee, err := settings.GetTargetNodeFee(rp, nil)
        if err == nil {
            response.TargetNodeFee = targetNodeFee
        }
        return err
    })
    wg.Go(func() error {
        maxNodeFee, err := settings.GetMaximumNodeFee(rp, nil)
        if err == nil {
            response.MaxNodeFee = maxNodeFee
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Suggest minimum node fee
    suggestedMinNodeFee := response.NodeFee + SuggestedNodeFeeDelta
    if suggestedMinNodeFee < response.MinNodeFee {
        suggestedMinNodeFee = response.MinNodeFee
    }
    response.SuggestedNodeFee = suggestedMinNodeFee

    // Return response
    return &response, nil

}
