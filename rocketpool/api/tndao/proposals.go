package tndao

import (
    "github.com/rocket-pool/rocketpool-go/dao"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func getProposals(c *cli.Context) (*api.TNDAOProposalsResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.TNDAOProposalsResponse{}

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }

    // Get proposals
    proposals, err := dao.GetDAOProposalsWithMember(rp, "", nodeAccount.Address, nil)
    if err != nil {
        return nil, err
    }
    response.Proposals = proposals

    // Return response
    return &response, nil

}

