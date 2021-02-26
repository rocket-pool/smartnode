package tndao

import (
    "context"

    "github.com/ethereum/go-ethereum/common"
    tndao "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    tnsettings "github.com/rocket-pool/rocketpool-go/settings/trustednode"
    "golang.org/x/sync/errgroup"
)


// Check if the proposal cooldown for a trusted node is active
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
    var currentBlock uint64
    var proposalExecutedBlock uint64
    var actionBlocks uint64

    // Get current block
    wg.Go(func() error {
        header, err := rp.Client.HeaderByNumber(context.Background(), nil)
        if err == nil {
            currentBlock = header.Number.Uint64()
        }
        return err
    })

    // Get proposal executed block
    wg.Go(func() error {
        var err error
        proposalExecutedBlock, err = tndao.GetMemberProposalExecutedBlock(rp, proposalType, nodeAddress, nil)
        return err
    })

    // Get action window in blocks
    wg.Go(func() error {
        var err error
        actionBlocks, err = tnsettings.GetProposalActionBlocks(rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return false, err
    }

    // Return
    return (currentBlock < (proposalExecutedBlock + actionBlocks)), nil

}


// Check if members can leave the trusted node DAO
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

