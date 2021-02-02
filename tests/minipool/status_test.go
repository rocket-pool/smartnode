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


func TestSubmitMinipoolWithdrawable(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register nodes
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Create & stake minipool
    mp, err := minipoolutils.CreateMinipool(rp, nodeAccount, eth.EthToWei(32))
    if err != nil { t.Fatal(err) }
    if err := minipoolutils.StakeMinipool(rp, mp, nodeAccount); err != nil { t.Fatal(err) }

    // Get & check initial minipool withdrawable status
    if withdrawable, err := minipool.GetMinipoolWithdrawable(rp, mp.Address, nil); err != nil {
        t.Error(err)
    } else if withdrawable {
        t.Error("Incorrect initial minipool withdrawable status")
    }

    // Submit minipool withdrawable status
    if _, err := minipool.SubmitMinipoolWithdrawable(rp, mp.Address, eth.EthToWei(32), eth.EthToWei(32), trustedNodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check updated minipool withdrawable status
    if withdrawable, err := minipool.GetMinipoolWithdrawable(rp, mp.Address, nil); err != nil {
        t.Error(err)
    } else if !withdrawable {
        t.Error("Incorrect updated minipool withdrawable status")
    }

}

