package odao

import (
	"github.com/rocket-pool/rocketpool-go/dao"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)


func canVoteOnProposal(c *cli.Context, proposalId uint64) (*api.CanVoteOnTNDAOProposalResponse, error) {

    // Get services
    if err := services.RequireNodeTrusted(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanVoteOnTNDAOProposalResponse{}

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }

    // Data
    var wg errgroup.Group
    var memberJoinedBlock uint64
    var proposalCreatedBlock uint64

    // Check proposal exists
    wg.Go(func() error {
        proposalCount, err := dao.GetProposalCount(rp, nil)
        if err == nil {
            response.DoesNotExist = (proposalId > proposalCount)
        }
        return err
    })

    // Check proposal state
    wg.Go(func() error {
        proposalState, err := dao.GetProposalState(rp, proposalId, nil)
        if err == nil {
            response.InvalidState = (proposalState != rptypes.Active)
        }
        return err
    })

    // Check if member has already voted
    wg.Go(func() error {
        hasVoted, err := dao.GetProposalMemberVoted(rp, proposalId, nodeAccount.Address, nil)
        if err == nil {
            response.AlreadyVoted = hasVoted
        }
        return err
    })

    // Get member joined block
    wg.Go(func() error {
        var err error
        memberJoinedBlock, err = trustednode.GetMemberJoinedBlock(rp, nodeAccount.Address, nil)
        return err
    })

    // Get proposal created block
    wg.Go(func() error {
        var err error
        proposalCreatedBlock, err = dao.GetProposalCreatedBlock(rp, proposalId, nil)
        return err
    })

    // Get gas estimate
    wg.Go(func() error {
        opts, err := w.GetNodeAccountTransactor()
        if err != nil { 
            return err 
        }
        gasInfo, err := trustednode.EstimateVoteOnProposalGas(rp, proposalId, false, opts)
        if err == nil {
            response.GasInfo = gasInfo
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Check data
    response.JoinedAfterCreated = (memberJoinedBlock >= proposalCreatedBlock)

    // Update & return response
    response.CanVote = !(response.DoesNotExist || response.InvalidState || response.JoinedAfterCreated || response.AlreadyVoted)
    return &response, nil

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
    hash, err := trustednode.VoteOnProposal(rp, proposalId, support, opts)
    if err != nil {
        return nil, err
    }
    response.TxHash = hash

    // Return response
    return &response, nil

}

