package minipool

import (
    "testing"

    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/tests/utils/evm"
    minipoolutils "github.com/rocket-pool/rocketpool-go/tests/utils/minipool"
    nodeutils "github.com/rocket-pool/rocketpool-go/tests/utils/node"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


func TestQueueLengths(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register nodes
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Get & check queue lengths
    if queueLengths, err := minipool.GetQueueLengths(rp, nil); err != nil {
        t.Error(err)
    } else {
        if queueLengths.Total != 0 {
            t.Errorf("Incorrect total queue length 1 %d", queueLengths.Total)
        }
        if queueLengths.FullDeposit != 0 {
            t.Errorf("Incorrect full deposit queue length 1 %d", queueLengths.FullDeposit)
        }
        if queueLengths.HalfDeposit != 0 {
            t.Errorf("Incorrect half deposit queue length 1 %d", queueLengths.HalfDeposit)
        }
        if queueLengths.EmptyDeposit != 0 {
            t.Errorf("Incorrect empty deposit queue length 1 %d", queueLengths.EmptyDeposit)
        }
    }

    // Create minipool
    if _, err := minipoolutils.CreateMinipool(rp, nodeAccount, eth.EthToWei(32)); err != nil { t.Fatal(err) }

    // Get & check queue lengths
    if queueLengths, err := minipool.GetQueueLengths(rp, nil); err != nil {
        t.Error(err)
    } else {
        if queueLengths.Total != 1 {
            t.Errorf("Incorrect total queue length 2 %d", queueLengths.Total)
        }
        if queueLengths.FullDeposit != 1 {
            t.Errorf("Incorrect full deposit queue length 2 %d", queueLengths.FullDeposit)
        }
        if queueLengths.HalfDeposit != 0 {
            t.Errorf("Incorrect half deposit queue length 2 %d", queueLengths.HalfDeposit)
        }
        if queueLengths.EmptyDeposit != 0 {
            t.Errorf("Incorrect empty deposit queue length 2 %d", queueLengths.EmptyDeposit)
        }
    }

    // Create minipool
    if _, err := minipoolutils.CreateMinipool(rp, nodeAccount, eth.EthToWei(16)); err != nil { t.Fatal(err) }

    // Get & check queue lengths
    if queueLengths, err := minipool.GetQueueLengths(rp, nil); err != nil {
        t.Error(err)
    } else {
        if queueLengths.Total != 2 {
            t.Errorf("Incorrect total queue length 3 %d", queueLengths.Total)
        }
        if queueLengths.FullDeposit != 1 {
            t.Errorf("Incorrect full deposit queue length 3 %d", queueLengths.FullDeposit)
        }
        if queueLengths.HalfDeposit != 1 {
            t.Errorf("Incorrect half deposit queue length 3 %d", queueLengths.HalfDeposit)
        }
        if queueLengths.EmptyDeposit != 0 {
            t.Errorf("Incorrect empty deposit queue length 3 %d", queueLengths.EmptyDeposit)
        }
    }

    // Create minipool
    if _, err := minipoolutils.CreateMinipool(rp, trustedNodeAccount, eth.EthToWei(0)); err != nil { t.Fatal(err) }

    // Get & check queue lengths
    if queueLengths, err := minipool.GetQueueLengths(rp, nil); err != nil {
        t.Error(err)
    } else {
        if queueLengths.Total != 3 {
            t.Errorf("Incorrect total queue length 4 %d", queueLengths.Total)
        }
        if queueLengths.FullDeposit != 1 {
            t.Errorf("Incorrect full deposit queue length 4 %d", queueLengths.FullDeposit)
        }
        if queueLengths.HalfDeposit != 1 {
            t.Errorf("Incorrect half deposit queue length 4 %d", queueLengths.HalfDeposit)
        }
        if queueLengths.EmptyDeposit != 1 {
            t.Errorf("Incorrect empty deposit queue length 4 %d", queueLengths.EmptyDeposit)
        }
    }

}

