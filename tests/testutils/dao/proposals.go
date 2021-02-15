package dao

import (
    trustednodedao "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    trustednodesettings "github.com/rocket-pool/rocketpool-go/settings/trustednode"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/accounts"
    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
)


// Pass and execute a proposal
func PassAndExecuteProposal(rp *rocketpool.RocketPool, proposalId uint64, trustedNodeAccount *accounts.Account) error {

    // Get proposal voting delay
    voteDelayBlocks, err := trustednodesettings.GetProposalVoteDelayBlocks(rp, nil)
    if err != nil { return err }

    // Mine blocks until proposal voting delay has passed
    if err := evm.MineBlocks(int(voteDelayBlocks)); err != nil { return err }

    // Vote on & execute proposal
    if _, err := trustednodedao.VoteOnProposal(rp, proposalId, true, trustedNodeAccount.GetTransactor()); err != nil { return err }
    if _, err := trustednodedao.ExecuteProposal(rp, proposalId, trustedNodeAccount.GetTransactor()); err != nil { return err }

    // Return
    return nil

}

