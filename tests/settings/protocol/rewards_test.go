package protocol

import (
    "testing"

    "github.com/rocket-pool/rocketpool-go/settings/protocol"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
)


func TestRewardsSettings(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Set & get rewards claim interval blocks
    var rewardsClaimIntervalBlocks uint64 = 1
    if _, err := protocol.BootstrapRewardsClaimIntervalBlocks(rp, rewardsClaimIntervalBlocks, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := protocol.GetRewardsClaimIntervalBlocks(rp, nil); err != nil {
        t.Error(err)
    } else if value != rewardsClaimIntervalBlocks {
        t.Error("Incorrect rewards claim interval blocks value")
    }

}

