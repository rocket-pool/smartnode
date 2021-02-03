package minipool

import (
    "bytes"
    "testing"

    "github.com/rocket-pool/rocketpool-go/deposit"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/network"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/tests/utils/evm"
    minipoolutils "github.com/rocket-pool/rocketpool-go/tests/utils/minipool"
    nodeutils "github.com/rocket-pool/rocketpool-go/tests/utils/node"
    rptypes "github.com/rocket-pool/rocketpool-go/types"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


func TestDetails(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register nodes
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Get current network node fee
    networkNodeFee, err := network.GetNodeFee(rp, nil)
    if err != nil { t.Fatal(err) }

    // Create minipool
    mp, err := minipoolutils.CreateMinipool(rp, nodeAccount, eth.EthToWei(32))
    if err != nil { t.Fatal(err) }

    // Make user deposit
    depositOpts := userAccount.GetTransactor();
    depositOpts.Value = eth.EthToWei(16)
    if _, err := deposit.Deposit(rp, depositOpts); err != nil { t.Fatal(err) }

    // Stake minipool
    if err := minipoolutils.StakeMinipool(rp, mp, nodeAccount); err != nil { t.Fatal(err) }

    // Set minipool withdrawable status
    if _, err := minipool.SubmitMinipoolWithdrawable(rp, mp.Address, eth.EthToWei(34), eth.EthToWei(36), trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get & check minipool details
    if status, err := mp.GetStatusDetails(nil); err != nil {
        t.Error(err)
    } else if status.Status != rptypes.Withdrawable {
        t.Errorf("Incorrect minipool status %s", status.Status.String())
    }
    if depositType, err := mp.GetDepositType(nil); err != nil {
        t.Error(err)
    } else if depositType != rptypes.Full {
        t.Errorf("Incorrect minipool deposit type %s", depositType.String())
    }
    if node, err := mp.GetNodeDetails(nil); err != nil {
        t.Error(err)
    } else {
        if !bytes.Equal(node.Address.Bytes(), nodeAccount.Address.Bytes()) {
            t.Errorf("Incorrect minipool node address %s", node.Address.Hex())
        }
        if node.Fee != networkNodeFee {
            t.Errorf("Incorrect minipool node fee %f", node.Fee)
        }
        if node.DepositBalance.Cmp(eth.EthToWei(16)) != 0 {
            t.Errorf("Incorrect minipool node deposit balance %s", node.DepositBalance.String())
        }
        if node.RefundBalance.Cmp(eth.EthToWei(16)) != 0 {
            t.Errorf("Incorrect minipool node refund balance %s", node.RefundBalance.String())
        }
        if !node.DepositAssigned {
            t.Error("Incorrect minipool node deposit assigned status")
        }
    }
    if user, err := mp.GetUserDetails(nil); err != nil {
        t.Error(err)
    } else {
        if user.DepositBalance.Cmp(eth.EthToWei(16)) != 0 {
            t.Errorf("Incorrect minipool user deposit balance %s", user.DepositBalance.String())
        }
        if !user.DepositAssigned {
            t.Error("Incorrect minipool user deposit assigned status")
        }
    }
    if staking, err := mp.GetStakingDetails(nil); err != nil {
        t.Error(err)
    } else {
        if staking.StartBalance.Cmp(eth.EthToWei(34)) != 0 {
            t.Errorf("Incorrect minipool staking start balance %s", staking.StartBalance.String())
        }
        if staking.EndBalance.Cmp(eth.EthToWei(36)) != 0 {
            t.Errorf("Incorrect minipool staking end balance %s", staking.EndBalance.String())
        }
    }

}

