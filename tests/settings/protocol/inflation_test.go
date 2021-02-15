package protocol

import (
    "testing"

    "github.com/rocket-pool/rocketpool-go/settings/protocol"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
)


func TestInflationSettings(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Set & get inflation interval rate
    inflationIntervalRate := 0.5
    if _, err := protocol.BootstrapInflationIntervalRate(rp, inflationIntervalRate, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := protocol.GetInflationIntervalRate(rp, nil); err != nil {
        t.Error(err)
    } else if value != inflationIntervalRate {
        t.Error("Incorrect inflation interval rate value")
    }

    // Set & get inflation interval blocks
    var inflationIntervalBlocks uint64 = 1
    if _, err := protocol.BootstrapInflationIntervalBlocks(rp, inflationIntervalBlocks, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := protocol.GetInflationIntervalBlocks(rp, nil); err != nil {
        t.Error(err)
    } else if value != inflationIntervalBlocks {
        t.Error("Incorrect inflation interval blocks value")
    }

    // Set & get inflation start block
    var inflationStartBlock uint64 = 1000000
    if _, err := protocol.BootstrapInflationStartBlock(rp, inflationStartBlock, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := protocol.GetInflationStartBlock(rp, nil); err != nil {
        t.Error(err)
    } else if value != inflationStartBlock {
        t.Error("Incorrect inflation start block value")
    }

}

