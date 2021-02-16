package trustednode

import (
    "bytes"
    "math/big"
    "testing"

    trustednodedao "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/rocket-pool/rocketpool-go/node"
    trustednodesettings "github.com/rocket-pool/rocketpool-go/settings/trustednode"

    daoutils "github.com/rocket-pool/rocketpool-go/tests/testutils/dao"
    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
    minipoolutils "github.com/rocket-pool/rocketpool-go/tests/testutils/minipool"
    nodeutils "github.com/rocket-pool/rocketpool-go/tests/testutils/node"
)


func TestMemberDetails(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Get & check initial member details
    if members, err := trustednodedao.GetMembers(rp, nil); err != nil {
        t.Error(err)
    } else if len(members) != 0 {
        t.Error("Incorrect initial trusted node DAO member count")
    }

    // Set proposal cooldown
    if _, err := trustednodesettings.BootstrapProposalCooldown(rp, 0, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Register nodes
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Bootstrap trusted node DAO member
    memberId := "coolguy"
    memberEmail := "coolguy@rocketpool.net"
    if _, err := trustednodedao.BootstrapMember(rp, memberId, memberEmail, trustedNodeAccount.Address, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get RPL bond amount
    rplBondAmount, err := trustednodesettings.GetRPLBond(rp, nil)
    if err != nil { t.Fatal(err) }

    // Mint trusted node RPL bond & join trusted node DAO
    if err := nodeutils.MintTrustedNodeBond(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }
    if _, err := trustednodedao.Join(rp, trustedNodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Submit, pass & execute replace member proposal
    proposalId, _, err := trustednodedao.ProposeReplaceMember(rp, "replace me", trustedNodeAccount.Address, nodeAccount.Address, "newguy", "newguy@rocketpool.net", trustedNodeAccount.GetTransactor())
    if err != nil { t.Fatal(err) }
    if err := daoutils.PassAndExecuteProposal(rp, proposalId, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Create an unbonded minipool
    if _, err := minipoolutils.CreateMinipool(rp, ownerAccount, trustedNodeAccount, big.NewInt(0)); err != nil { t.Fatal(err) }

    // Get & check updated member details
    if members, err := trustednodedao.GetMembers(rp, nil); err != nil {
        t.Error(err)
    } else if len(members) != 1 {
        t.Error("Incorrect updated trusted node DAO member count")
    } else {
        member := members[0]
        if !bytes.Equal(member.Address.Bytes(), trustedNodeAccount.Address.Bytes()) {
            t.Errorf("Incorrect member address %s", member.Address.Hex())
        }
        if !member.Exists {
            t.Error("Incorrect member exists status")
        }
        if member.ID != memberId {
            t.Errorf("Incorrect member ID %s", member.ID)
        }
        if member.Email != memberEmail {
            t.Errorf("Incorrect member email %s", member.Email)
        }
        if member.JoinedBlock == 0 {
            t.Errorf("Incorrect member joined block %d", member.JoinedBlock)
        }
        if member.LastProposalBlock == 0 {
            t.Errorf("Incorrect member last proposal block %d", member.LastProposalBlock)
        }
        if member.RPLBondAmount.Cmp(rplBondAmount) != 0 {
            t.Errorf("Incorrect member RPL bond amount %s", member.RPLBondAmount.String())
        }
        if member.UnbondedValidatorCount != 1 {
            t.Errorf("Incorrect member unbonded validator count %d", member.UnbondedValidatorCount)
        }
    }

    // Get & check member invite executed block
    if inviteExecutedBlock, err := trustednodedao.GetMemberInviteProposalExecutedBlock(rp, trustedNodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if inviteExecutedBlock == 0 {
        t.Errorf("Incorrect member invite proposal executed block %d", inviteExecutedBlock)
    }

    // Get & check member replacement address
    if replacementAddress, err := trustednodedao.GetMemberReplacementAddress(rp, trustedNodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if !bytes.Equal(replacementAddress.Bytes(), nodeAccount.Address.Bytes()) {
        t.Errorf("Incorrect member replacement address %s", replacementAddress.Hex())
    }

}

