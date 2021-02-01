package node

import (
    "testing"

    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/tests/utils/evm"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


func TestDeposit(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register node
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get initial node minipool count
    minipoolCount1, err := minipool.GetNodeMinipoolCount(rp, nodeAccount.Address, nil)
    if err != nil {
        t.Fatal(err)
    }

    // Deposit
    opts := nodeAccount.GetTransactor()
    opts.Value = eth.EthToWei(16)
    if _, err := node.Deposit(rp, 0, opts); err != nil {
        t.Fatal(err)
    }

    // Get & check updated node minipool count
    minipoolCount2, err := minipool.GetNodeMinipoolCount(rp, nodeAccount.Address, nil)
    if err != nil {
        t.Fatal(err)
    } else if minipoolCount2 != minipoolCount1 + 1 {
        t.Error("Incorrect node minipool count")
    }

}

