package minipool

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/tokens"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
	minipoolutils "github.com/rocket-pool/rocketpool-go/tests/testutils/minipool"
	nodeutils "github.com/rocket-pool/rocketpool-go/tests/testutils/node"
	"github.com/rocket-pool/rocketpool-go/tests/testutils/validator"
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
    mp, err := minipoolutils.CreateMinipool(rp, ownerAccount, nodeAccount, eth.EthToWei(32))
    if err != nil { t.Fatal(err) }

    // Make user deposit
    depositOpts := userAccount.GetTransactor();
    depositOpts.Value = eth.EthToWei(16)
    if _, err := deposit.Deposit(rp, depositOpts); err != nil { t.Fatal(err) }

    // Stake minipool
    if err := minipoolutils.StakeMinipool(rp, mp, nodeAccount); err != nil { t.Fatal(err) }

    // Set minipool withdrawable status
    if _, err := minipool.SubmitMinipoolWithdrawable(rp, mp.Address, trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get & check minipool details
    if status, err := mp.GetStatusDetails(nil); err != nil {
        t.Error(err)
    } else {
        if status.Status != rptypes.Withdrawable {
            t.Errorf("Incorrect minipool status %s", status.Status.String())
        }
        if status.StatusBlock == 0 {
            t.Errorf("Incorrect minipool status block %d", status.StatusBlock)
        }
        if status.StatusTime.Unix() == 0 {
            t.Errorf("Incorrect minipool status time %v", status.StatusTime)
        }
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
        if user.DepositAssignedTime.Unix() == 0 {
            t.Errorf("Incorrect minipool user deposit assigned time %v", user.DepositAssignedTime)
        }
    }
    if withdrawalCredentials, err := mp.GetWithdrawalCredentials(nil); err != nil {
        t.Error(err)
    } else {
        withdrawalPrefix := byte(1)
        padding := make([]byte, 11)
        expectedWithdrawalCredentials := bytes.Join([][]byte{[]byte{withdrawalPrefix}, padding, mp.Address.Bytes()}, []byte{})
        if !bytes.Equal(withdrawalCredentials.Bytes(), expectedWithdrawalCredentials) {
            t.Errorf("Incorrect minipool withdrawal credentials %s", hex.EncodeToString(withdrawalCredentials.Bytes()))
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
    mp, err := minipoolutils.CreateMinipool(rp, ownerAccount, nodeAccount, eth.EthToWei(32))
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
    mp, err := minipoolutils.CreateMinipool(rp, ownerAccount, nodeAccount, eth.EthToWei(32))
    if err != nil { t.Fatal(err) }

    // Get validator & deposit data
    validatorPubkey, err := validator.GetValidatorPubkey()
    if err != nil { t.Fatal(err) }
    withdrawalCredentials, err := mp.GetWithdrawalCredentials(nil)
    if err != nil { t.Fatal(err) }
    validatorSignature, err := validator.GetValidatorSignature()
    if err != nil { t.Fatal(err) }
    depositDataRoot, err := validator.GetDepositDataRoot(validatorPubkey, withdrawalCredentials, validatorSignature)
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



func TestDissolve(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register node
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Create minipool
    mp, err := minipoolutils.CreateMinipool(rp, ownerAccount, nodeAccount, eth.EthToWei(16))
    if err != nil { t.Fatal(err) }

    // Get & check initial minipool status
    if status, err := mp.GetStatus(nil); err != nil {
        t.Error(err)
    } else if status != rptypes.Initialized {
        t.Errorf("Incorrect initial minipool status %s", status.String())
    }

    // Dissolve minipool
    if _, err := mp.Dissolve(nodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check updated minipool status
    if status, err := mp.GetStatus(nil); err != nil {
        t.Error(err)
    } else if status != rptypes.Dissolved {
        t.Errorf("Incorrect updated minipool status %s", status.String())
    }

}


func TestClose(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register node
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Create minipool
    mp, err := minipoolutils.CreateMinipool(rp, ownerAccount, nodeAccount, eth.EthToWei(16))
    if err != nil { t.Fatal(err) }

    // Dissolve minipool
    if _, err := mp.Dissolve(nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get & check initial minipool exists status
    if exists, err := minipool.GetMinipoolExists(rp, mp.Address, nil); err != nil {
        t.Error(err)
    } else if !exists {
        t.Error("Incorrect initial minipool exists status")
    }

    // Close minipool
    if _, err := mp.Close(nodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check updated minipool exists status
    if exists, err := minipool.GetMinipoolExists(rp, mp.Address, nil); err != nil {
        t.Error(err)
    } else if exists {
        t.Error("Incorrect updated minipool exists status")
    }

}


func TestDestroy(t *testing.T) {

    // TODO

}


func TestWithdrawValidatorBalance(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register nodes
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Create minipool
    mp, err := minipoolutils.CreateMinipool(rp, ownerAccount, nodeAccount, eth.EthToWei(16))
    if err != nil { t.Fatal(err) }

    // Make user deposit
    userDepositAmount := eth.EthToWei(16)
    userDepositOpts := userAccount.GetTransactor()
    userDepositOpts.Value = userDepositAmount
    if _, err := deposit.Deposit(rp, userDepositOpts); err != nil { t.Fatal(err) }

    // Stake minipool
    if err := minipoolutils.StakeMinipool(rp, mp, nodeAccount); err != nil { t.Fatal(err) }

    // Set minipool withdrawable status
    if _, err := minipool.SubmitMinipoolWithdrawable(rp, mp.Address, trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get initial token contract ETH balances
    rethContractBalance1, err := tokens.GetRETHContractETHBalance(rp, nil)
    if err != nil {
        t.Fatal(err)
    }

    // Withdraw minipool validator balance
    opts := swcAccount.GetTransactor()
    opts.Value = eth.EthToWei(32)
    if _, err := mp.Contract.Transfer(opts); err != nil {
        t.Fatal(err)
    }

    // Get node balances before withdrawal
    nodeBalance1, err := tokens.GetBalances(rp, nodeAccount.Address, nil)
    if err != nil {
        t.Fatal(err)
    }

    // Call ProcessWithdrawal method
    if _, err := mp.DistributeBalance(nodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check updated node ETH balances
    if nodeBalance2, err := tokens.GetBalances(rp, nodeAccount.Address, nil); err != nil {
        t.Fatal(err)
    } else if nodeBalance2.ETH.Cmp(nodeBalance1.ETH) != 1 {
        t.Error("node ETH balance did not increase after processing withdrawal")
    }

    // Get & check updated token contract ETH balances
    if rethContractBalance2, err := tokens.GetRETHContractETHBalance(rp, nil); err != nil {
        t.Fatal(err)
    } else if rethContractBalance2.Cmp(rethContractBalance1) != 1 {
        t.Error("rETH contract ETH balance did not increase after processing withdrawal")
    }

    // Get & check rETH collateral amount & rate
    if rethTotalCollateral, err := tokens.GetRETHTotalCollateral(rp, nil); err != nil {
        t.Fatal(err)
    } else if rethTotalCollateral.Cmp(userDepositAmount) != 0 {
        t.Errorf("Incorrect rETH total collateral amount %s", rethTotalCollateral.String())
    }
    if rethCollateralRate, err := tokens.GetRETHCollateralRate(rp, nil); err != nil {
        t.Fatal(err)
    } else if rethCollateralRate != 1 {
        t.Errorf("Incorrect rETH collateral rate %f", rethCollateralRate)
    }

    // Confirm the minipool still exists
    if exists, err := minipool.GetMinipoolExists(rp, mp.Address, nil); err != nil {
        t.Error(err)
    } else if !exists {
        t.Error("Minipool no longer exists but it should")
    }

}


func TestWithdrawValidatorBalanceAndDestroy(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register nodes
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Create minipool
    mp, err := minipoolutils.CreateMinipool(rp, ownerAccount, nodeAccount, eth.EthToWei(16))
    if err != nil { t.Fatal(err) }

    // Make user deposit
    userDepositAmount := eth.EthToWei(16)
    userDepositOpts := userAccount.GetTransactor()
    userDepositOpts.Value = userDepositAmount
    if _, err := deposit.Deposit(rp, userDepositOpts); err != nil { t.Fatal(err) }

    // Stake minipool
    if err := minipoolutils.StakeMinipool(rp, mp, nodeAccount); err != nil { t.Fatal(err) }

    // Set minipool withdrawable status
    if _, err := minipool.SubmitMinipoolWithdrawable(rp, mp.Address, trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get initial token contract ETH balances
    rethContractBalance1, err := tokens.GetRETHContractETHBalance(rp, nil)
    if err != nil {
        t.Fatal(err)
    }

    // Withdraw minipool validator balance
    opts := swcAccount.GetTransactor()
    opts.Value = eth.EthToWei(32)
    if _, err := mp.Contract.Transfer(opts); err != nil {
        t.Fatal(err)
    }

    // Get node balances before withdrawal
    nodeBalance1, err := tokens.GetBalances(rp, nodeAccount.Address, nil)
    if err != nil {
        t.Fatal(err)
    }

    // Call DistributeBalanceAndDestroy method
    if _, err := mp.DistributeBalanceAndDestroy(nodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check updated node ETH balances
    if nodeBalance2, err := tokens.GetBalances(rp, nodeAccount.Address, nil); err != nil {
        t.Fatal(err)
    } else if nodeBalance2.ETH.Cmp(nodeBalance1.ETH) != 1 {
        t.Error("node ETH balance did not increase after processing withdrawal")
    }

    // Get & check updated token contract ETH balances
    if rethContractBalance2, err := tokens.GetRETHContractETHBalance(rp, nil); err != nil {
        t.Fatal(err)
    } else if rethContractBalance2.Cmp(rethContractBalance1) != 1 {
        t.Error("rETH contract ETH balance did not increase after processing withdrawal")
    }

    // Get & check rETH collateral amount & rate
    if rethTotalCollateral, err := tokens.GetRETHTotalCollateral(rp, nil); err != nil {
        t.Fatal(err)
    } else if rethTotalCollateral.Cmp(userDepositAmount) != 0 {
        t.Errorf("Incorrect rETH total collateral amount %s", rethTotalCollateral.String())
    }
    if rethCollateralRate, err := tokens.GetRETHCollateralRate(rp, nil); err != nil {
        t.Fatal(err)
    } else if rethCollateralRate != 1 {
        t.Errorf("Incorrect rETH collateral rate %f", rethCollateralRate)
    }

    // Confirm the minipool still exists
    if exists, err := minipool.GetMinipoolExists(rp, mp.Address, nil); err != nil {
        t.Error(err)
    } else if exists {
        t.Error("Minipool still exists but it shouldn't")
    }

}


func TestDelegateUpgrade(t *testing.T) {

    // TODO

}


func TestDelegateRollback(t *testing.T) {

    // TODO

}


func TestSetUseLatestDelegate(t *testing.T) {

    // TODO

}


func TestGetUseLatestDelegate(t *testing.T) {

    // TODO

}


func TestGetDelegate(t *testing.T) {

    // TODO

}


func TestGetPreviousDelegate(t *testing.T) {

    // TODO

}


func TestGetEffectiveDelegate(t *testing.T) {

    // TODO

}

