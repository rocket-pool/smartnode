package trustednode

import (
    "testing"

    trustednodedao "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/rocket-pool/rocketpool-go/node"
    trustednodesettings "github.com/rocket-pool/rocketpool-go/settings/trustednode"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/accounts"
    daoutils "github.com/rocket-pool/rocketpool-go/tests/testutils/dao"
    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
    nodeutils "github.com/rocket-pool/rocketpool-go/tests/testutils/node"
)


func TestInviteMember(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Set proposal cooldown
    if _, err := trustednodesettings.BootstrapProposalCooldown(rp, 0, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Register nodes
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount1); err != nil { t.Fatal(err) }

    // Submit, pass & execute invite member proposal
    proposalId, _, err := trustednodedao.ProposeInviteMember(rp, "invite coolguy", nodeAccount.Address, "coolguy", "coolguy@rocketpool.net", trustedNodeAccount1.GetTransactor())
    if err != nil { t.Fatal(err) }
    if err := daoutils.PassAndExecuteProposal(rp, proposalId, []*accounts.Account{trustedNodeAccount1}); err != nil { t.Fatal(err) }

    // Get & check initial member exists status
    if exists, err := trustednodedao.GetMemberExists(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if exists {
        t.Error("Incorrect initial member exists status")
    }

    // Mint trusted node RPL bond & join trusted node DAO
    if err := nodeutils.MintTrustedNodeBond(rp, ownerAccount, nodeAccount); err != nil { t.Fatal(err) }
    if _, err := trustednodedao.Join(rp, nodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check updated member exists status
    if exists, err := trustednodedao.GetMemberExists(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if !exists {
        t.Error("Incorrect updated member exists status")
    }

}


func TestMemberLeave(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Set proposal cooldown
    if _, err := trustednodesettings.BootstrapProposalCooldown(rp, 0, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Register nodes
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount1); err != nil { t.Fatal(err) }
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount2); err != nil { t.Fatal(err) }
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount3); err != nil { t.Fatal(err) }
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount4); err != nil { t.Fatal(err) }

    // Submit, pass & execute member leave proposal
    proposalId, _, err := trustednodedao.ProposeMemberLeave(rp, "bye", trustedNodeAccount1.Address, trustedNodeAccount1.GetTransactor())
    if err != nil { t.Fatal(err) }
    if err := daoutils.PassAndExecuteProposal(rp, proposalId, []*accounts.Account{
        trustedNodeAccount1,
        trustedNodeAccount2,
        trustedNodeAccount3,
        trustedNodeAccount4,
    }); err != nil { t.Fatal(err) }

    // Get & check initial member exists status
    if exists, err := trustednodedao.GetMemberExists(rp, trustedNodeAccount1.Address, nil); err != nil {
        t.Error(err)
    } else if !exists {
        t.Error("Incorrect initial member exists status")
    }

    // Leave trusted node DAO
    if _, err := trustednodedao.Leave(rp, trustedNodeAccount1.Address, trustedNodeAccount1.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check updated member exists status
    if exists, err := trustednodedao.GetMemberExists(rp, trustedNodeAccount1.Address, nil); err != nil {
        t.Error(err)
    } else if exists {
        t.Error("Incorrect updated member exists status")
    }

}

