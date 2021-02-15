package dao

import (
    "github.com/rocket-pool/rocketpool-go/dao/trustednode"
    "github.com/rocket-pool/rocketpool-go/rocketpool"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/accounts"
    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
)


// Pass and execute a proposal
func PassAndExecuteProposal(rp *rocketpool.RocketPool, proposalId uint64, trustedNodeAccount *accounts.Account) error {

    // Mine blocks until proposal voting delay has passed
    if err := evm.MineBlocks(1); err != nil { return err }

    // Vote on & execute proposal
    if _, err := trustednode.VoteOnProposal(rp, proposalId, true, trustedNodeAccount.GetTransactor()); err != nil { return err }
    if _, err := trustednode.ExecuteProposal(rp, proposalId, trustedNodeAccount.GetTransactor()); err != nil { return err }

    // Return
    return nil

}

