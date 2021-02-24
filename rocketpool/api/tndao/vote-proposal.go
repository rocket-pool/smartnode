package tndao

import (
    "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canVoteOnProposal(c *cli.Context, proposalId uint64) (*api.CanVoteOnTNDAOProposalResponse, error) {
    return nil, nil
}


func voteOnProposal(c *cli.Context, proposalId uint64, support bool) (*api.VoteOnTNDAOProposalResponse, error) {

    // Get services
    if err := services.RequireNodeTrusted(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.VoteOnTNDAOProposalResponse{}

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Vote on proposal
    txReceipt, err := trustednode.VoteOnProposal(rp, proposalId, support, opts)
    if err != nil {
        return nil, err
    }
    response.TxHash = txReceipt.TxHash

    // Return response
    return &response, nil

}

