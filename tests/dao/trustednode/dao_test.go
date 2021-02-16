package trustednode

import (
    "bytes"
    "math/big"
    "testing"

    trustednodedao "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/rocket-pool/rocketpool-go/node"
    trustednodesettings "github.com/rocket-pool/rocketpool-go/settings/trustednode"
    "github.com/rocket-pool/rocketpool-go/tokens"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
    minipoolutils "github.com/rocket-pool/rocketpool-go/tests/testutils/minipool"
    rplutils "github.com/rocket-pool/rocketpool-go/tests/testutils/tokens/rpl"
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

    // Mint RPL bond to node & allow trusted node DAO contract to spend it
    rplBondAmount, err := trustednodesettings.GetRPLBond(rp, nil)
    if err != nil { t.Fatal(err) }
    rocketDAONodeTrustedActionsAddress, err := rp.GetAddress("rocketDAONodeTrustedActions")
    if err != nil { t.Fatal(err) }
    if err := rplutils.MintRPL(rp, ownerAccount, trustedNodeAccount, rplBondAmount); err != nil { t.Fatal(err) }
    if _, err := tokens.ApproveRPL(rp, *rocketDAONodeTrustedActionsAddress, rplBondAmount, trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Join trusted node DAO
    if _, err := trustednodedao.Join(rp, trustedNodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Submit replace member proposal
    proposalId, _, err := trustednodedao.ProposeReplaceMember(rp, "replace me", trustedNodeAccount.Address, nodeAccount.Address, "newguy", "newguy@rocketpool.net", trustedNodeAccount.GetTransactor())
    if err != nil { t.Fatal(err) }

    // Mine blocks until proposal voting delay has passed
    voteDelayBlocks, err := trustednodesettings.GetProposalVoteDelayBlocks(rp, nil)
    if err != nil { t.Fatal(err) }
    if err := evm.MineBlocks(int(voteDelayBlocks)); err != nil { t.Fatal(err) }

    // Pass & execute replace member proposal
    if _, err := trustednodedao.VoteOnProposal(rp, proposalId, true, trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if _, err := trustednodedao.ExecuteProposal(rp, proposalId, trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

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

