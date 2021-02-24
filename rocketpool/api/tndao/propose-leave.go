package tndao

import (
    "fmt"

    "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canProposeLeave(c *cli.Context) (*api.CanProposeTNDAOLeaveResponse, error) {

    // Get services
    if err := services.RequireNodeTrusted(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanProposeTNDAOLeaveResponse{}

    // Data
    var wg errgroup.Group
    var memberCount uint64
    var minMemberCount uint64

    // Check if proposal cooldown is active
    wg.Go(func() error {
        nodeAccount, err := w.GetNodeAccount()
        if err != nil {
            return err
        }
        proposalCooldownActive, err := getProposalCooldownActive(rp, nodeAccount.Address)
        if err == nil {
            response.ProposalCooldownActive = proposalCooldownActive
        }
        return err
    })

    // Get member count
    wg.Go(func() error {
        var err error
        memberCount, err = trustednode.GetMemberCount(rp, nil)
        return err
    })

    // Get min member count
    wg.Go(func() error {
        var err error
        minMemberCount, err = trustednode.GetMinimumMemberCount(rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Check data
    response.InsufficientMembers = (memberCount <= minMemberCount)

    // Update & return response
    response.CanPropose = !(response.ProposalCooldownActive || response.InsufficientMembers)
    return &response, nil

}


func proposeLeave(c *cli.Context) (*api.ProposeTNDAOLeaveResponse, error) {

    // Get services
    if err := services.RequireNodeTrusted(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.ProposeTNDAOLeaveResponse{}

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }

    // Data
    var wg errgroup.Group
    var nodeMemberId string
    var nodeMemberEmail string

    // Get node member details
    wg.Go(func() error {
        var err error
        nodeMemberId, err = trustednode.GetMemberID(rp, nodeAccount.Address, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        nodeMemberEmail, err = trustednode.GetMemberEmail(rp, nodeAccount.Address, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Submit proposal
    message := fmt.Sprintf("%s (%s) leaves", nodeMemberId, nodeMemberEmail)
    proposalId, txReceipt, err := trustednode.ProposeMemberLeave(rp, message, nodeAccount.Address, opts)
    if err != nil {
        return nil, err
    }
    response.ProposalId = proposalId
    response.TxHash = txReceipt.TxHash

    // Return response
    return &response, nil

}

