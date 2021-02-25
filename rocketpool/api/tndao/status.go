package tndao

import (
    "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func getStatus(c *cli.Context) (*api.TNDAOStatusResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.TNDAOStatusResponse{}

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }

    // Sync
    var wg errgroup.Group

    // Get member status
    wg.Go(func() error {
        isMember, err := trustednode.GetMemberExists(rp, nodeAccount.Address, nil)
        if err == nil {
            response.IsMember = isMember
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Return response
    return &response, nil

}

