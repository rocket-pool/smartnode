package minipool

import (
    "bytes"
    "testing"

    "github.com/rocket-pool/rocketpool-go/deposit"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/network"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/rocket-pool/rocketpool-go/tests/utils/evm"
    minipoolutils "github.com/rocket-pool/rocketpool-go/tests/utils/minipool"
    nodeutils "github.com/rocket-pool/rocketpool-go/tests/utils/node"
    "github.com/rocket-pool/rocketpool-go/tests/utils/validator"
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


func TestRefund(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register node
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Create minipool
    mp, err := minipoolutils.CreateMinipool(rp, nodeAccount, eth.EthToWei(32))
    if err != nil { t.Fatal(err) }

    // Make user deposit
    depositOpts := userAccount.GetTransactor();
    depositOpts.Value = eth.EthToWei(16)
    if _, err := deposit.Deposit(rp, depositOpts); err != nil { t.Fatal(err) }

    // Get initial node refund balance
    nodeRefundBalance1, err := mp.GetNodeRefundBalance(nil)
    if err != nil {
        t.Fatal(err)
    }

    // Refund
    if _, err := mp.Refund(nodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check updated node refund balance
    nodeRefundBalance2, err := mp.GetNodeRefundBalance(nil)
    if err != nil {
        t.Fatal(err)
    } else if nodeRefundBalance2.Cmp(nodeRefundBalance1) != -1 {
        t.Error("Node refund balance did not decrease after refunding from minipool")
    }

}


func TestStake(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register node
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Create minipool
    mp, err := minipoolutils.CreateMinipool(rp, nodeAccount, eth.EthToWei(32))
    if err != nil { t.Fatal(err) }

    // Get validator & deposit data
    validatorPubkey, err := validator.GetValidatorPubkey()
    if err != nil { t.Fatal(err) }
    validatorSignature, err := validator.GetValidatorSignature()
    if err != nil { t.Fatal(err) }
    depositDataRoot, err := validator.GetDepositDataRoot(validatorPubkey, validator.GetWithdrawalCredentials(), validatorSignature)
    if err != nil { t.Fatal(err) }

    // Get & check initial minipool status
    if status, err := mp.GetStatus(nil); err != nil {
        t.Error(err)
    } else if status != rptypes.Prelaunch {
        t.Errorf("Incorrect initial minipool status %s", status.String())
    }

    // Stake minipool
    if _, err := mp.Stake(validatorPubkey, validatorSignature, depositDataRoot, nodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check updated minipool status
    if status, err := mp.GetStatus(nil); err != nil {
        t.Error(err)
    } else if status != rptypes.Staking {
        t.Errorf("Incorrect updated minipool status %s", status.String())
    }

}


func TestWithdraw(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register nodes
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Create minipool
    mp, err := minipoolutils.CreateMinipool(rp, nodeAccount, eth.EthToWei(32))
    if err != nil { t.Fatal(err) }

    // Stake minipool
    if err := minipoolutils.StakeMinipool(rp, mp, nodeAccount); err != nil { t.Fatal(err) }

    // Set minipool withdrawable status
    if _, err := minipool.SubmitMinipoolWithdrawable(rp, mp.Address, eth.EthToWei(32), eth.EthToWei(32), trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get & check initial minipool exists status
    if exists, err := minipool.GetMinipoolExists(rp, mp.Address, nil); err != nil {
        t.Error(err)
    } else if !exists {
        t.Error("Incorrect initial minipool exists status")
    }

    // Disable minipool withdrawal delay
    withdrawalDelay, err := settings.GetMinipoolWithdrawalDelay(rp, nil)
    if err != nil { t.Fatal(err) }
    if _, err := settings.SetMinipoolWithdrawalDelay(rp, 0, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Withdraw minipool
    if _, err := mp.Withdraw(nodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Re-enable minipool withdrawal delay
    if _, err := settings.SetMinipoolWithdrawalDelay(rp, withdrawalDelay, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get & check updated minipool exists status
    if exists, err := minipool.GetMinipoolExists(rp, mp.Address, nil); err != nil {
        t.Error(err)
    } else if exists {
        t.Error("Incorrect updated minipool exists status")
    }

}

