package tndao

import (
    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canLeave(c *cli.Context) (*api.CanLeaveTNDAOResponse, error) {

    // Get services
    if err := services.RequireNodeTrusted(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanLeaveTNDAOResponse{}

    // Sync
    var wg errgroup.Group

    // Check proposal expired status
    wg.Go(func() error {
        nodeAccount, err := w.GetNodeAccount()
        if err != nil {
            return err
        }
        proposalExpired, err := getProposalExpired(rp, nodeAccount.Address, "leave")
        if err == nil {
            response.ProposalExpired = proposalExpired
        }
        return err
    })

    // Check if members can leave the trusted node DAO
    wg.Go(func() error {
        membersCanLeave, err := getMembersCanLeave(rp)
        if err == nil {
            response.InsufficientMembers = !membersCanLeave
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Update & return response
    response.CanLeave = !(response.ProposalExpired || response.InsufficientMembers)
    return &response, nil

}


func leave(c *cli.Context, bondRefundAddress common.Address) (*api.LeaveTNDAOResponse, error) {

    // Get services
    if err := services.RequireNodeTrusted(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.LeaveTNDAOResponse{}

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Leave
    txReceipt, err := trustednode.Leave(rp, bondRefundAddress, opts)
    if err != nil {
        return nil, err
    }
    response.TxHash = txReceipt.TxHash

    // Return response
    return &response, nil

}

