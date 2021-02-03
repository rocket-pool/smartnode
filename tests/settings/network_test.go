package settings

import (
    "testing"

    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/rocket-pool/rocketpool-go/utils/eth"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
)


func TestNetworkSettings(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Set & get node consensus threshold
    nodeConsensusThreshold := 0.1
    if _, err := settings.SetNodeConsensusThreshold(rp, nodeConsensusThreshold, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetNodeConsensusThreshold(rp, nil); err != nil {
        t.Error(err)
    } else if value != nodeConsensusThreshold {
        t.Error("Incorrect node consensus threshold value")
    }

    // Set & get network balance submissions enabled
    submitBalancesEnabled := false
    if _, err := settings.SetSubmitBalancesEnabled(rp, submitBalancesEnabled, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetSubmitBalancesEnabled(rp, nil); err != nil {
        t.Error(err)
    } else if value != submitBalancesEnabled {
        t.Error("Incorrect network balance submissions enabled value")
    }

    // Set & get network balance submission frequency
    var submitBalancesFrequency uint64 = 10
    if _, err := settings.SetSubmitBalancesFrequency(rp, submitBalancesFrequency, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetSubmitBalancesFrequency(rp, nil); err != nil {
        t.Error(err)
    } else if value != submitBalancesFrequency {
        t.Error("Incorrect network balance submission frequency value")
    }

    // Set & get process withdrawals enabled
    processWithdrawalsEnabled := false
    if _, err := settings.SetProcessWithdrawalsEnabled(rp, processWithdrawalsEnabled, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetProcessWithdrawalsEnabled(rp, nil); err != nil {
        t.Error(err)
    } else if value != processWithdrawalsEnabled {
        t.Error("Incorrect process withdrawals enabled value")
    }

    // Set & get minimum node fee
    minimumNodeFee := 0.80
    if _, err := settings.SetMinimumNodeFee(rp, minimumNodeFee, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetMinimumNodeFee(rp, nil); err != nil {
        t.Error(err)
    } else if value != minimumNodeFee {
        t.Error("Incorrect minimum node fee value")
    }

    // Set & get target node fee
    targetNodeFee := 0.85
    if _, err := settings.SetTargetNodeFee(rp, targetNodeFee, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetTargetNodeFee(rp, nil); err != nil {
        t.Error(err)
    } else if value != targetNodeFee {
        t.Error("Incorrect target node fee value")
    }

    // Set & get maximum node fee
    maximumNodeFee := 0.90
    if _, err := settings.SetMaximumNodeFee(rp, maximumNodeFee, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetMaximumNodeFee(rp, nil); err != nil {
        t.Error(err)
    } else if value != maximumNodeFee {
        t.Error("Incorrect maximum node fee value")
    }

    // Set & get node fee demand range
    nodeFeeDemandRange := eth.EthToWei(10)
    if _, err := settings.SetNodeFeeDemandRange(rp, nodeFeeDemandRange, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetNodeFeeDemandRange(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(nodeFeeDemandRange) != 0 {
        t.Error("Incorrect node fee demand range value")
    }

    // Set & get target rETH collateral rate
    targetRethCollateralRate := 0.95
    if _, err := settings.SetTargetRethCollateralRate(rp, targetRethCollateralRate, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetTargetRethCollateralRate(rp, nil); err != nil {
        t.Error(err)
    } else if value != targetRethCollateralRate {
        t.Error("Incorrect target rETH collateral rate value")
    }

}

