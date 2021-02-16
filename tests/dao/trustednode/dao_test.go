package trustednode

import (
    "bytes"
    "testing"

    trustednodedao "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/rocket-pool/rocketpool-go/node"
    trustednodesettings "github.com/rocket-pool/rocketpool-go/settings/trustednode"
    "github.com/rocket-pool/rocketpool-go/tokens"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
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

    // Register node
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

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
    if _, err := trustednodedao.Join(rp, trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

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
        if member.LastProposalBlock != 0 {
            t.Errorf("Incorrect member last proposal block %d", member.LastProposalBlock)
        }
        if member.RPLBondAmount.Cmp(rplBondAmount) != 0 {
            t.Errorf("Incorrect member RPL bond amount %s", member.RPLBondAmount.String())
        }
        if member.UnbondedValidatorCount != 0 {
            t.Errorf("Incorrect member unbonded validator count %d", member.UnbondedValidatorCount)
        }
    }

}

