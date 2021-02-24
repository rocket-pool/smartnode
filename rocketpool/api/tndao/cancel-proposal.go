package tndao

import (
    "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canCancelProposal(c *cli.Context, proposalId uint64) (*api.CanCancelTNDAOProposalResponse, error) {
    return nil, nil
}


func cancelProposal(c *cli.Context, proposalId uint64) (*api.CancelTNDAOProposalResponse, error) {

    // Get services
    if err := services.RequireNodeTrusted(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CancelTNDAOProposalResponse{}

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Cancel proposal
    txReceipt, err := trustednode.CancelProposal(rp, proposalId, opts)
    if err != nil {
        return nil, err
    }
    response.TxHash = txReceipt.TxHash

    // Return response
    return &response, nil

}

