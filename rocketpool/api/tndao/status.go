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

    // Get membership status
    isMember, err := trustednode.GetMemberExists(rp, nodeAccount.Address, nil)
    if err != nil {
        return nil, err
    }
    response.IsMember = isMember

    // Sync
    var wg errgroup.Group

    // Get pending executed proposal statuses
    if isMember {

        // Check if node can leave
        wg.Go(func() error {
            leaveActionable, err := getProposalIsActionable(rp, nodeAccount.Address, "leave")
            if err == nil {
                response.CanLeave = leaveActionable
            }
            return err
        })

        // Check if node can replace position
        wg.Go(func() error {
            replaceActionable, err := getProposalIsActionable(rp, nodeAccount.Address, "replace")
            if err == nil {
                response.CanReplace = replaceActionable
            }
            return err
        })

    } else {

        // Check if node can join
        wg.Go(func() error {
            joinActionable, err := getProposalIsActionable(rp, nodeAccount.Address, "invited")
            if err == nil {
                response.CanJoin = joinActionable
            }
            return err
        })

    }

    // Get total DAO members
    wg.Go(func() error {
        memberCount, err := trustednode.GetMemberCount(rp, nil)
        if err == nil {
            response.TotalMembers = memberCount
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

