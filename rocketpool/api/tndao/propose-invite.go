package tndao

import (
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canProposeInvite(c *cli.Context, memberAddress common.Address) (*api.CanProposeTNDAOInviteResponse, error) {
    return nil, nil
}


func proposeInvite(c *cli.Context, memberAddress common.Address, memberId, memberEmail string) (*api.ProposeTNDAOInviteResponse, error) {

    // Get services
    if err := services.RequireNodeTrusted(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.ProposeTNDAOInviteResponse{}

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Submit proposal
    message := fmt.Sprintf("invite %s (%s)", memberId, memberEmail)
    proposalId, txReceipt, err := trustednode.ProposeInviteMember(rp, message, memberAddress, memberId, memberEmail, opts)
    if err != nil {
        return nil, err
    }
    response.ProposalId = proposalId
    response.TxHash = txReceipt.TxHash

    // Return response
    return &response, nil

}

