package trustednode

import (
    "testing"

    "github.com/rocket-pool/rocketpool-go/settings/trustednode"
    "github.com/rocket-pool/rocketpool-go/utils/eth"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
    daoutils "github.com/rocket-pool/rocketpool-go/tests/testutils/dao"
    nodeutils "github.com/rocket-pool/rocketpool-go/tests/testutils/node"
)


func TestBootstrapMembersSettings(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Set & get quorum
    quorum := 0.1
    if _, err := trustednode.BootstrapQuorum(rp, quorum, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetQuorum(rp, nil); err != nil {
        t.Error(err)
    } else if value != quorum {
        t.Error("Incorrect quorum value")
    }

    // Set & get rpl bond
    rplBond := eth.EthToWei(1)
    if _, err := trustednode.BootstrapRPLBond(rp, rplBond, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetRPLBond(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(rplBond) != 0 {
        t.Error("Incorrect rpl bond value")
    }

    // Set & get maximum unbonded minipools
    var minipoolUnbondedMax uint64 = 1
    if _, err := trustednode.BootstrapMinipoolUnbondedMax(rp, minipoolUnbondedMax, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetMinipoolUnbondedMax(rp, nil); err != nil {
        t.Error(err)
    } else if value != minipoolUnbondedMax {
        t.Error("Incorrect maximum unbonded minipools value")
    }

}


func TestProposeMembersSettings(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Set proposal cooldown
    if _, err := trustednode.BootstrapProposalCooldown(rp, 0, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Register trusted node
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Set & get quorum
    quorum := 0.1
    if proposalId, _, err := trustednode.ProposeQuorum(rp, quorum, trustedNodeAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if err := daoutils.PassAndExecuteProposal(rp, proposalId, trustedNodeAccount); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetQuorum(rp, nil); err != nil {
        t.Error(err)
    } else if value != quorum {
        t.Error("Incorrect quorum value")
    }

    // Set & get rpl bond
    rplBond := eth.EthToWei(1)
    if proposalId, _, err := trustednode.ProposeRPLBond(rp, rplBond, trustedNodeAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if err := daoutils.PassAndExecuteProposal(rp, proposalId, trustedNodeAccount); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetRPLBond(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(rplBond) != 0 {
        t.Error("Incorrect rpl bond value")
    }

    // Set & get maximum unbonded minipools
    var minipoolUnbondedMax uint64 = 1
    if proposalId, _, err := trustednode.ProposeMinipoolUnbondedMax(rp, minipoolUnbondedMax, trustedNodeAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if err := daoutils.PassAndExecuteProposal(rp, proposalId, trustedNodeAccount); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetMinipoolUnbondedMax(rp, nil); err != nil {
        t.Error(err)
    } else if value != minipoolUnbondedMax {
        t.Error("Incorrect maximum unbonded minipools value")
    }

}

