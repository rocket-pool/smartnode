package trustednode

import (
    "testing"

    "github.com/rocket-pool/rocketpool-go/settings/trustednode"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
)


func TestProposalsSettings(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Set & get cooldown
    var cooldown uint64 = 1
    if _, err := trustednode.BootstrapProposalCooldown(rp, cooldown, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalCooldown(rp, nil); err != nil {
        t.Error(err)
    } else if value != cooldown {
        t.Error("Incorrect cooldown value")
    }

    // Set & get vote blocks
    var voteBlocks uint64 = 1
    if _, err := trustednode.BootstrapProposalVoteBlocks(rp, voteBlocks, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalVoteBlocks(rp, nil); err != nil {
        t.Error(err)
    } else if value != voteBlocks {
        t.Error("Incorrect vote blocks value")
    }

    // Set & get vote delay blocks
    var voteDelayBlocks uint64 = 1000
    if _, err := trustednode.BootstrapProposalVoteDelayBlocks(rp, voteDelayBlocks, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalVoteDelayBlocks(rp, nil); err != nil {
        t.Error(err)
    } else if value != voteDelayBlocks {
        t.Error("Incorrect vote delay blocks value")
    }

    // Set & get execute blocks
    var executeBlocks uint64 = 1
    if _, err := trustednode.BootstrapProposalExecuteBlocks(rp, executeBlocks, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalExecuteBlocks(rp, nil); err != nil {
        t.Error(err)
    } else if value != executeBlocks {
        t.Error("Incorrect execute blocks value")
    }

    // Set & get action blocks
    var actionBlocks uint64 = 1
    if _, err := trustednode.BootstrapProposalActionBlocks(rp, actionBlocks, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalActionBlocks(rp, nil); err != nil {
        t.Error(err)
    } else if value != actionBlocks {
        t.Error("Incorrect action blocks value")
    }

}

