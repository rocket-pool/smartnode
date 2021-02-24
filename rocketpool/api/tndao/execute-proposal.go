package tndao

import (
    "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canExecuteProposal(c *cli.Context, proposalId uint64) (*api.CanExecuteTNDAOProposalResponse, error) {
    return nil, nil
}


func executeProposal(c *cli.Context, proposalId uint64) (*api.ExecuteTNDAOProposalResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.ExecuteTNDAOProposalResponse{}

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Cancel proposal
    txReceipt, err := trustednode.ExecuteProposal(rp, proposalId, opts)
    if err != nil {
        return nil, err
    }
    response.TxHash = txReceipt.TxHash

    // Return response
    return &response, nil

}

