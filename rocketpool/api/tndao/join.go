package tndao

import (
    tndao "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    tnsettings "github.com/rocket-pool/rocketpool-go/settings/trustednode"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canJoin(c *cli.Context) (*api.CanJoinTNDAOResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanJoinTNDAOResponse{}

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }

    // Data
    var wg errgroup.Group
    var nodeRplBalance *big.Int
    var rplBondAmount *big.Int

    // Check proposal expired status
    wg.Go(func() error {
        proposalExpired, err := getProposalExpired(rp, nodeAccount.Address, "invited")
        if err == nil {
            response.ProposalExpired = proposalExpired
        }
        return err
    })

    // Get node RPL balance
    wg.Go(func() error {
        var err error
        nodeRplBalance, err = tokens.GetRPLBalance(rp, nodeAccount.Address, nil)
        return err
    })

    // Get RPL bond amount
    wg.Go(func() error {
        var err error
        rplBondAmount, err = tnsettings.GetRPLBond(rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Check data
    response.InsufficientRplBalance = (nodeRplBalance.Cmp(rplBondAmount) < 0)

    // Update & return response
    response.CanJoin = !(response.ProposalExpired || proposal.InsufficientRplBalance)
    return &response, nil

}


func join(c *cli.Context) (*api.JoinTNDAOResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.JoinTNDAOResponse{}

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Join
    txReceipt, err := tndao.Join(rp, opts)
    if err != nil {
        return nil, err
    }
    response.TxHash = txReceipt.TxHash

    // Return response
    return &response, nil

}

