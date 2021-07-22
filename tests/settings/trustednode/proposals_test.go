package trustednode

import (
	"testing"

	"github.com/rocket-pool/rocketpool-go/settings/trustednode"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/accounts"
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
    if _, err := trustednode.BootstrapProposalCooldownTime(rp, cooldown, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalCooldownTime(rp, nil); err != nil {
        t.Error(err)
    } else if value != cooldown {
        t.Error("Incorrect cooldown value")
    }

    // Set & get vote time
    var voteTime uint64 = 10
    if _, err := trustednode.BootstrapProposalVoteTime(rp, voteTime, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalVoteTime(rp, nil); err != nil {
        t.Error(err)
    } else if value != voteTime {
        t.Error("Incorrect vote time value")
    }

    // Set & get execute time
    var executeTime uint64 = 10
    if _, err := trustednode.BootstrapProposalExecuteTime(rp, executeTime, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalExecuteTime(rp, nil); err != nil {
        t.Error(err)
    } else if value != executeTime {
        t.Error("Incorrect execute time value")
    }

    // Set & get action time
    var actionTime uint64 = 10
    if _, err := trustednode.BootstrapProposalActionTime(rp, actionTime, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalActionTime(rp, nil); err != nil {
        t.Error(err)
    } else if value != actionTime {
        t.Error("Incorrect action time value")
    }

    // Set & get vote delay time
    var voteDelayTime uint64 = 1000
    if _, err := trustednode.BootstrapProposalVoteDelayTime(rp, voteDelayTime, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalVoteDelayTime(rp, nil); err != nil {
        t.Error(err)
    } else if value != voteDelayTime {
        t.Error("Incorrect vote delay time value")
    }

}


func TestProposeProposalsSettings(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Set proposal cooldown
    if _, err := trustednode.BootstrapProposalCooldownTime(rp, 0, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if _, err := trustednode.BootstrapProposalVoteDelayTime(rp, 5, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Register trusted node
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount1); err != nil { t.Fatal(err) }
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount2); err != nil { t.Fatal(err) }
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount3); err != nil { t.Fatal(err) }

    // Set & get cooldown
    var cooldown uint64 = 1
    if proposalId, _, err := trustednode.ProposeProposalCooldownTime(rp, cooldown, trustedNodeAccount1.GetTransactor()); err != nil {
        t.Error(err)
    } else if err := daoutils.PassAndExecuteProposal(rp, proposalId, []*accounts.Account{trustedNodeAccount1, trustedNodeAccount2}); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalCooldownTime(rp, nil); err != nil {
        t.Error(err)
    } else if value != cooldown {
        t.Error("Incorrect cooldown value")
    }

    // Set & get vote time
    var voteTime uint64 = 10
    if proposalId, _, err := trustednode.ProposeProposalVoteTime(rp, voteTime, trustedNodeAccount1.GetTransactor()); err != nil {
        t.Error(err)
    } else if err := daoutils.PassAndExecuteProposal(rp, proposalId, []*accounts.Account{trustedNodeAccount1, trustedNodeAccount2}); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalVoteTime(rp, nil); err != nil {
        t.Error(err)
    } else if value != voteTime {
        t.Error("Incorrect vote time value")
    }

    // Set & get execute time
    var executeTime uint64 = 10
    if proposalId, _, err := trustednode.ProposeProposalExecuteTime(rp, executeTime, trustedNodeAccount1.GetTransactor()); err != nil {
        t.Error(err)
    } else if err := daoutils.PassAndExecuteProposal(rp, proposalId, []*accounts.Account{trustedNodeAccount1, trustedNodeAccount2}); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalExecuteTime(rp, nil); err != nil {
        t.Error(err)
    } else if value != executeTime {
        t.Error("Incorrect execute time value")
    }

    // Set & get action time
    var actionTime uint64 = 10
    if proposalId, _, err := trustednode.ProposeProposalActionTime(rp, actionTime, trustedNodeAccount1.GetTransactor()); err != nil {
        t.Error(err)
    } else if err := daoutils.PassAndExecuteProposal(rp, proposalId, []*accounts.Account{trustedNodeAccount1, trustedNodeAccount2}); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalActionTime(rp, nil); err != nil {
        t.Error(err)
    } else if value != actionTime {
        t.Error("Incorrect action time value")
    }

    // Set & get vote delay time
    var voteDelayTime uint64 = 1000
    if proposalId, _, err := trustednode.ProposeProposalVoteDelayTime(rp, voteDelayTime, trustedNodeAccount1.GetTransactor()); err != nil {
        t.Error(err)
    } else if err := daoutils.PassAndExecuteProposal(rp, proposalId, []*accounts.Account{trustedNodeAccount1, trustedNodeAccount2}); err != nil {
        t.Error(err)
    } else if value, err := trustednode.GetProposalVoteDelayTime(rp, nil); err != nil {
        t.Error(err)
    } else if value != voteDelayTime {
        t.Error("Incorrect vote delay time value")
    }

}

