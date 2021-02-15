package protocol

import (
    "testing"

    protocoldao "github.com/rocket-pool/rocketpool-go/dao/protocol"
    "github.com/rocket-pool/rocketpool-go/settings/protocol"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
)


func TestRewardsSettings(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Bootstrap a claimer & get claimer settings
    claimerPerc := 0.1
    if _, err := protocoldao.BootstrapClaimer(rp, "rocketClaimNode", claimerPerc, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else {
        if value, err := protocol.GetRewardsClaimerPerc(rp, "rocketClaimNode", nil); err != nil {
            t.Error(err)
        } else if value != claimerPerc {
            t.Errorf("Incorrect rewards claimer percent %f", value)
        }
        if value, err := protocol.GetRewardsClaimerPercBlockUpdated(rp, "rocketClaimNode", nil); err != nil {
            t.Error(err)
        } else if value == 0 {
            t.Errorf("Incorrect rewards claimer percent block updated %d", value)
        }
        if value, err := protocol.GetRewardsClaimersPercTotal(rp, nil); err != nil {
            t.Error(err)
        } else if value == 0 {
            t.Errorf("Incorrect rewards claimers total percent %f", value)
        }
    }

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

