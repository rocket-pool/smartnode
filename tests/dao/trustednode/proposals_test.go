package trustednode

import (
    "testing"

    trustednodedao "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/rocket-pool/rocketpool-go/node"
    trustednodesettings "github.com/rocket-pool/rocketpool-go/settings/trustednode"

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
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Submit, pass & execute invite member proposal
    proposalId, _, err := trustednodedao.ProposeInviteMember(rp, "invite coolguy", nodeAccount.Address, "coolguy", "coolguy@rocketpool.net", trustedNodeAccount.GetTransactor())
    if err != nil { t.Fatal(err) }
    if err := daoutils.PassAndExecuteProposal(rp, proposalId, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Get initial member exists status
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

    // Get updated member exists status
    if exists, err := trustednodedao.GetMemberExists(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if !exists {
        t.Error("Incorrect updated member exists status")
    }

}

