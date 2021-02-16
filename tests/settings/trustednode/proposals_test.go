package trustednode

import (
    "testing"

    "github.com/rocket-pool/rocketpool-go/settings/trustednode"

    daoutils "github.com/rocket-pool/rocketpool-go/tests/testutils/dao"
    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
    nodeutils "github.com/rocket-pool/rocketpool-go/tests/testutils/node"
)


func TestBootstrapProposalsSettings(t *testing.T) {

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
    var voteBlocks uint64 = 10
    if _, err := trustednode.BootstrapProposalVoteBlocks(rp, voteBlocks, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalVoteBlocks(rp, nil); err != nil {
        t.Error(err)
    } else if value != voteBlocks {
        t.Error("Incorrect vote blocks value")
    }

    // Set & get execute blocks
    var executeBlocks uint64 = 10
    if _, err := trustednode.BootstrapProposalExecuteBlocks(rp, executeBlocks, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalExecuteBlocks(rp, nil); err != nil {
        t.Error(err)
    } else if value != executeBlocks {
        t.Error("Incorrect execute blocks value")
    }

    // Set & get action blocks
    var actionBlocks uint64 = 10
    if _, err := trustednode.BootstrapProposalActionBlocks(rp, actionBlocks, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalActionBlocks(rp, nil); err != nil {
        t.Error(err)
    } else if value != actionBlocks {
        t.Error("Incorrect action blocks value")
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

}


func TestProposeProposalsSettings(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Set proposal cooldown
    if _, err := trustednode.BootstrapProposalCooldown(rp, 0, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Register trusted node
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Set & get cooldown
    var cooldown uint64 = 1
    if proposalId, _, err := trustednode.ProposeProposalCooldown(rp, cooldown, trustedNodeAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if err := daoutils.PassAndExecuteProposal(rp, proposalId, trustedNodeAccount); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalCooldown(rp, nil); err != nil {
        t.Error(err)
    } else if value != cooldown {
        t.Error("Incorrect cooldown value")
    }

    // Set & get vote blocks
    var voteBlocks uint64 = 10
    if proposalId, _, err := trustednode.ProposeProposalVoteBlocks(rp, voteBlocks, trustedNodeAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if err := daoutils.PassAndExecuteProposal(rp, proposalId, trustedNodeAccount); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalVoteBlocks(rp, nil); err != nil {
        t.Error(err)
    } else if value != voteBlocks {
        t.Error("Incorrect vote blocks value")
    }

    // Set & get execute blocks
    var executeBlocks uint64 = 10
    if proposalId, _, err := trustednode.ProposeProposalExecuteBlocks(rp, executeBlocks, trustedNodeAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if err := daoutils.PassAndExecuteProposal(rp, proposalId, trustedNodeAccount); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalExecuteBlocks(rp, nil); err != nil {
        t.Error(err)
    } else if value != executeBlocks {
        t.Error("Incorrect execute blocks value")
    }

    // Set & get action blocks
    var actionBlocks uint64 = 10
    if proposalId, _, err := trustednode.ProposeProposalActionBlocks(rp, actionBlocks, trustedNodeAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if err := daoutils.PassAndExecuteProposal(rp, proposalId, trustedNodeAccount); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalActionBlocks(rp, nil); err != nil {
        t.Error(err)
    } else if value != actionBlocks {
        t.Error("Incorrect action blocks value")
    }

    // Set & get vote delay blocks
    var voteDelayBlocks uint64 = 1000
    if proposalId, _, err := trustednode.ProposeProposalVoteDelayBlocks(rp, voteDelayBlocks, trustedNodeAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if err := daoutils.PassAndExecuteProposal(rp, proposalId, trustedNodeAccount); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalVoteDelayBlocks(rp, nil); err != nil {
        t.Error(err)
    } else if value != voteDelayBlocks {
        t.Error("Incorrect vote delay blocks value")
    }

}

