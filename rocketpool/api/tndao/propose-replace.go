package tndao

import (
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canProposeReplace(c *cli.Context, newMemberAddress common.Address) (*api.CanProposeTNDAOReplaceResponse, error) {
    return nil, nil
}


func proposeReplace(c *cli.Context, newMemberAddress common.Address, newMemberId, newMemberEmail string) (*api.ProposeTNDAOReplaceResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.ProposeTNDAOReplaceResponse{}

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
    message := fmt.Sprintf("replace %s (%s) with %s (%s)", nodeMemberId, nodeMemberEmail, newMemberId, newMemberEmail)
    proposalId, txReceipt, err := trustednode.ProposeReplaceMember(rp, message, nodeAccount.Address, newMemberAddress, newMemberId, newMemberEmail, opts)
    if err != nil {
        return nil, err
    }
    response.ProposalId = proposalId
    response.TxHash = txReceipt.TxHash

    // Return response
    return &response, nil

}

