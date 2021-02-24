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

