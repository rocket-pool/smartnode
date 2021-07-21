package odao

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao"
	tndao "github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	tnsettings "github.com/rocket-pool/rocketpool-go/settings/trustednode"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"golang.org/x/sync/errgroup"
)

// Settings
const ProposalStatesBatchSize = 50


// Check if the proposal cooldown for an oracle node is active
func getProposalCooldownActive(rp *rocketpool.RocketPool, nodeAddress common.Address) (bool, error) {

    // Data
    var wg errgroup.Group
    var currentBlock uint64
    var lastProposalBlock uint64
    var proposalCooldown uint64

    // Get current block
    wg.Go(func() error {
        header, err := rp.Client.HeaderByNumber(context.Background(), nil)
        if err == nil {
            currentBlock = header.Number.Uint64()
        }
        return err
    })

    // Get last proposal block
    wg.Go(func() error {
        var err error
        lastProposalBlock, err = tndao.GetMemberLastProposalBlock(rp, nodeAddress, nil)
        return err
    })

    // Get proposal cooldown
    wg.Go(func() error {
        var err error
        proposalCooldown, err = tnsettings.GetProposalCooldown(rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return false, err
    }

    // Return
    return ((currentBlock - lastProposalBlock) < proposalCooldown), nil

}


// Check if a proposal for a node exists & is actionable
func getProposalIsActionable(rp *rocketpool.RocketPool, nodeAddress common.Address, proposalType string) (bool, error) {

    // Data
    var wg errgroup.Group
    var currentTime uint64
    var proposalExecutedTime uint64
    var actionTime uint64

    // Get current block
    wg.Go(func() error {
        header, err := rp.Client.HeaderByNumber(context.Background(), nil)
        if err == nil {
            currentTime = header.Time
        }
        return err
    })

    // Get proposal executed block
    wg.Go(func() error {
        var err error
        proposalExecutedTime, err = tndao.GetMemberProposalExecutedTime(rp, proposalType, nodeAddress, nil)
        return err
    })

    // Get action window in blocks
    wg.Go(func() error {
        var err error
        actionTime, err = tnsettings.GetProposalActionTime(rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return false, err
    }

    // Return
    return (currentTime < (proposalExecutedTime + actionTime)), nil

}


// Check if members can leave the oracle DAO
func getMembersCanLeave(rp *rocketpool.RocketPool) (bool, error) {

    // Data
    var wg errgroup.Group
    var memberCount uint64
    var minMemberCount uint64

    // Get member count
    wg.Go(func() error {
        var err error
        memberCount, err = tndao.GetMemberCount(rp, nil)
        return err
    })

    // Get min member count
    wg.Go(func() error {
        var err error
        minMemberCount, err = tndao.GetMinimumMemberCount(rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return false, err
    }

    // Return
    return (memberCount > minMemberCount), nil

}


// Get all proposal states
func getProposalStates(rp *rocketpool.RocketPool) ([]rptypes.ProposalState, error) {

    // Get proposal IDs
    proposalIds, err := dao.GetDAOProposalIDs(rp, "rocketDAONodeTrustedProposals", nil)
    if err != nil {
        return []rptypes.ProposalState{}, err
    }

    // Load proposal states in batches
    states := make([]rptypes.ProposalState, len(proposalIds))
    for bsi := 0; bsi < len(proposalIds); bsi += ProposalStatesBatchSize {

        // Get batch start & end index
        psi := bsi
        pei := bsi + ProposalStatesBatchSize
        if pei > len(proposalIds) { pei = len(proposalIds) }

        // Load states
        var wg errgroup.Group
        for pi := psi; pi < pei; pi++ {
            pi := pi
            wg.Go(func() error {
                proposalState, err := dao.GetProposalState(rp, proposalIds[pi], nil)
                if err == nil { states[pi] = proposalState }
                return err
            })
        }
        if err := wg.Wait(); err != nil {
            return []rptypes.ProposalState{}, err
        }

    }

    // Return
    return states, nil

}

