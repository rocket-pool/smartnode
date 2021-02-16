package dao

import (
    "bytes"
    "fmt"
    "testing"

    "github.com/rocket-pool/rocketpool-go/dao"
    trustednodedao "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/rocket-pool/rocketpool-go/node"
    trustednodesettings "github.com/rocket-pool/rocketpool-go/settings/trustednode"
    rptypes "github.com/rocket-pool/rocketpool-go/types"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
    nodeutils "github.com/rocket-pool/rocketpool-go/tests/testutils/node"
)


func TestProposalDetails(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // The DAO to check for proposals under
    proposalDaoName := "rocketDAONodeTrustedProposals"

    // Set proposal cooldown
    if _, err := trustednodesettings.BootstrapProposalCooldown(rp, 0, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Register nodes
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Get & check initial proposal details
    if proposals, err := dao.GetProposalsWithMember(rp, trustedNodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if len(proposals) != 0 {
        t.Error("Incorrect initial proposal count")
    }
    if daoProposals, err := dao.GetDAOProposalsWithMember(rp, proposalDaoName, trustedNodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if len(daoProposals) != 0 {
        t.Error("Incorrect initial DAO proposal count")
    }

    // Submit invite member proposal
    proposalMessage := "invite coolguy"
    proposalMemberAddress := nodeAccount.Address
    proposalMemberId := "coolguy"
    proposalMemberEmail := "coolguy@rocketpool.net"
    proposalId, _, err := trustednodedao.ProposeInviteMember(rp, proposalMessage, proposalMemberAddress, proposalMemberId, proposalMemberEmail, trustedNodeAccount.GetTransactor())
    if err != nil { t.Fatal(err) }

    // Mine blocks until proposal voting delay has passed
    voteDelayBlocks, err := trustednodesettings.GetProposalVoteDelayBlocks(rp, nil)
    if err != nil { t.Fatal(err) }
    if err := evm.MineBlocks(int(voteDelayBlocks)); err != nil { t.Fatal(err) }

    // Vote on & execute proposal
    if _, err := trustednodedao.VoteOnProposal(rp, proposalId, true, trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if _, err := trustednodedao.ExecuteProposal(rp, proposalId, trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Submit invite member proposal & cancel it
    cancelledProposalId, _, err := trustednodedao.ProposeInviteMember(rp, "cancel this", nodeAccount.Address, "cancel", "cancel@rocketpool.net", trustedNodeAccount.GetTransactor())
    if err != nil { t.Fatal(err) }
    if _, err := trustednodedao.CancelProposal(rp, cancelledProposalId, trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get & check updated proposal details
    if proposals, err := dao.GetProposalsWithMember(rp, trustedNodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if len(proposals) != 2 {
        t.Error("Incorrect updated proposal count")
    } else {

        // Passed proposal
        proposal := proposals[0]
        if proposal.ID != proposalId {
            t.Errorf("Incorrect proposal ID %d", proposal.ID)
        }
        if proposal.DAO != proposalDaoName {
            t.Errorf("Incorrect proposal DAO %s", proposal.DAO)
        }
        if !bytes.Equal(proposal.ProposerAddress.Bytes(), trustedNodeAccount.Address.Bytes()) {
            t.Errorf("Incorrect proposal proposer address %s", proposal.ProposerAddress.Hex())
        }
        if proposal.CreatedBlock == 0 {
            t.Errorf("Incorrect proposal created block %d", proposal.CreatedBlock)
        }
        if proposal.StartBlock <= proposal.CreatedBlock {
            t.Errorf("Incorrect proposal start block %d", proposal.StartBlock)
        }
        if proposal.EndBlock <= proposal.StartBlock {
            t.Errorf("Incorrect proposal end block %d", proposal.EndBlock)
        }
        if proposal.ExpiryBlock <= proposal.EndBlock {
            t.Errorf("Incorrect proposal expiry block %d", proposal.ExpiryBlock)
        }
        if proposal.VotesRequired == 0.0 {
            t.Errorf("Incorrect proposal required votes %f", proposal.VotesRequired)
        }
        if proposal.VotesFor != 1.0 {
            t.Errorf("Incorrect proposal votes for %f", proposal.VotesFor)
        }
        if proposal.VotesAgainst != 0.0 {
            t.Errorf("Incorrect proposal votes against %f", proposal.VotesAgainst)
        }
        if !proposal.MemberVoted {
            t.Error("Incorrect proposal member voted status")
        }
        if !proposal.MemberSupported {
            t.Error("Incorrect proposal member supported status")
        }
        if proposal.IsCancelled {
            t.Error("Incorrect proposal cancelled status")
        }
        if !proposal.IsExecuted {
            t.Error("Incorrect proposal executed status")
        }
        if proposal.PayloadStr != fmt.Sprintf("proposalInvite(%s,%s,%s)", proposalMemberId, proposalMemberEmail, proposalMemberAddress.Hex()) {
            t.Errorf("Incorrect proposal payload string %s", proposal.PayloadStr)
        }
        if proposal.State != rptypes.Executed {
            t.Errorf("Incorrect proposal state %s", proposal.State.String())
        }

        // Cancelled proposal
        cancelledProposal := proposals[1]
        if cancelledProposal.ID != cancelledProposalId {
            t.Errorf("Incorrect cancelled proposal ID %d", cancelledProposal.ID)
        }
        if !cancelledProposal.IsCancelled {
            t.Error("Incorrect cancelled proposal cancelled status")
        }

    }
    if daoProposals, err := dao.GetDAOProposalsWithMember(rp, proposalDaoName, trustedNodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if len(daoProposals) != 2 {
        t.Error("Incorrect updated DAO proposal count")
    }

}

