package tndao

import (
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/math"
)


func canProposeKick(c *cli.Context, memberAddress common.Address, fineAmountWei *big.Int) (*api.CanProposeTNDAOKickResponse, error) {
    return nil, nil
}


func proposeKick(c *cli.Context, memberAddress common.Address, fineAmountWei *big.Int) (*api.ProposeTNDAOKickResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.ProposeTNDAOKickResponse{}

    // Data
    var wg errgroup.Group
    var memberId string
    var memberEmail string

    // Get member details
    wg.Go(func() error {
        var err error
        memberId, err = trustednode.GetMemberID(rp, memberAddress, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        memberEmail, err = trustednode.GetMemberEmail(rp, memberAddress, nil)
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
    message := fmt.Sprintf("kick %s (%s) with %.6f RPL fine", memberId, memberEmail, math.RoundDown(eth.WeiToEth(fineAmountWei), 6))
    proposalId, txReceipt, err := trustednode.ProposeKickMember(rp, message, memberAddress, fineAmountWei, opts)
    if err != nil {
        return nil, err
    }
    response.ProposalId = proposalId
    response.TxHash = txReceipt.TxHash

    // Return response
    return &response, nil

}

