package minipool

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	trustednodedao "github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/settings/trustednode"
	"github.com/rocket-pool/rocketpool-go/utils"

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
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Register nodes
	if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}
	if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil {
		t.Fatal(err)
	}

	// Get current network node fee
	networkNodeFee, err := network.GetNodeFee(rp, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create minipool
	mp, err := minipoolutils.CreateMinipool(t, rp, ownerAccount, nodeAccount, eth.EthToWei(32), 1)
	if err != nil {
		t.Fatal(err)
	}

	// Make user deposit
	depositOpts := userAccount.GetTransactor()
	depositOpts.Value = eth.EthToWei(16)
	if _, err := deposit.Deposit(rp, depositOpts); err != nil {
		t.Fatal(err)
	}

	// Delay for the time between depositing and staking
	scrubPeriod, err := trustednode.GetScrubPeriod(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = evm.IncreaseTime(int(scrubPeriod + 1))
	if err != nil {
		t.Fatal(fmt.Errorf("error increasing time: %w", err))
	}

	// Stake minipool
	if err := minipoolutils.StakeMinipool(rp, mp, nodeAccount); err != nil {
		t.Fatal(err)
	}

	// Set minipool withdrawable status
	if _, err := minipool.SubmitMinipoolWithdrawable(rp, mp.Address, trustedNodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

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
	if withdrawalCredentials, err := minipool.GetMinipoolWithdrawalCredentials(rp, mp.Address, nil); err != nil {
		t.Error(err)
	} else {
		withdrawalPrefix := byte(1)
		padding := make([]byte, 11)
		expectedWithdrawalCredentials := bytes.Join([][]byte{{withdrawalPrefix}, padding, mp.Address.Bytes()}, []byte{})
		if !bytes.Equal(withdrawalCredentials.Bytes(), expectedWithdrawalCredentials) {
			t.Errorf("Incorrect minipool withdrawal credentials %s", hex.EncodeToString(withdrawalCredentials.Bytes()))
		}
	}

}

func TestRefund(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Register node
	if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Create minipool
	mp, err := minipoolutils.CreateMinipool(t, rp, ownerAccount, nodeAccount, eth.EthToWei(32), 1)
	if err != nil {
		t.Fatal(err)
	}

	// Make user deposit
	depositOpts := userAccount.GetTransactor()
	depositOpts.Value = eth.EthToWei(16)
	if _, err := deposit.Deposit(rp, depositOpts); err != nil {
		t.Fatal(err)
	}

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
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Register node
	if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Create minipool
	mp, err := minipoolutils.CreateMinipool(t, rp, ownerAccount, nodeAccount, eth.EthToWei(32), 1)
	if err != nil {
		t.Fatal(err)
	}

	// Get validator & deposit data
	validatorPubkey, err := validator.GetValidatorPubkey(1)
	if err != nil {
		t.Fatal(err)
	}
	withdrawalCredentials, err := minipool.GetMinipoolWithdrawalCredentials(rp, mp.Address, nil)
	if err != nil {
		t.Fatal(err)
	}
	validatorSignature, err := validator.GetValidatorSignature(1)
	if err != nil {
		t.Fatal(err)
	}
	depositDataRoot, err := validator.GetDepositDataRoot(validatorPubkey, withdrawalCredentials, validatorSignature)
	if err != nil {
		t.Fatal(err)
	}

	// Get & check initial minipool status
	if status, err := mp.GetStatus(nil); err != nil {
		t.Error(err)
	} else if status != rptypes.Prelaunch {
		t.Errorf("Incorrect initial minipool status %s", status.String())
	}

	// Delay for the time between depositing and staking
	scrubPeriod, err := trustednode.GetScrubPeriod(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = evm.IncreaseTime(int(scrubPeriod + 1))
	if err != nil {
		t.Fatal(fmt.Errorf("error increasing time: %w", err))
	}

	// Stake minipool
	if _, err := mp.Stake(validatorSignature, depositDataRoot, nodeAccount.GetTransactor()); err != nil {
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
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Register node
	if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Create minipool
	mp, err := minipoolutils.CreateMinipool(t, rp, ownerAccount, nodeAccount, eth.EthToWei(16), 1)
	if err != nil {
		t.Fatal(err)
	}

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
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Register node
	if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Create minipool
	mp, err := minipoolutils.CreateMinipool(t, rp, ownerAccount, nodeAccount, eth.EthToWei(16), 1)
	if err != nil {
		t.Fatal(err)
	}

	// Dissolve minipool
	if _, err := mp.Dissolve(nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check initial minipool exists status
	if exists, err := minipool.GetMinipoolExists(rp, mp.Address, nil); err != nil {
		t.Error(err)
	} else if !exists {
		t.Error("Incorrect initial minipool exists status")
	}

	// Simulate a post-merge withdrawal by sending 16 ETH to the minipool
	opts := nodeAccount.GetTransactor()
	opts.Value = eth.EthToWei(16)
	hash, err := eth.SendTransaction(rp.Client, mp.Address, big.NewInt(1337), opts) // Ganache's default chain ID is 1337
	if err != nil {
		t.Errorf("Error sending ETH to minipool: %s", err.Error())
	}
	utils.WaitForTransaction(rp.Client, hash)

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

func TestFinalise(t *testing.T) {

	// TODO

}

func TestWithdrawValidatorBalance(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Register nodes
	if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}
	if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil {
		t.Fatal(err)
	}

	// Create minipool
	mp, err := minipoolutils.CreateMinipool(t, rp, ownerAccount, nodeAccount, eth.EthToWei(16), 1)
	if err != nil {
		t.Fatal(err)
	}

	// Make user deposit
	userDepositAmount := eth.EthToWei(16)
	userDepositOpts := userAccount.GetTransactor()
	userDepositOpts.Value = userDepositAmount
	if _, err := deposit.Deposit(rp, userDepositOpts); err != nil {
		t.Fatal(err)
	}

	// Delay for the time between depositing and staking
	scrubPeriod, err := trustednode.GetScrubPeriod(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = evm.IncreaseTime(int(scrubPeriod + 1))
	if err != nil {
		t.Fatal(fmt.Errorf("error increasing time: %w", err))
	}

	// Stake minipool
	if err := minipoolutils.StakeMinipool(rp, mp, nodeAccount); err != nil {
		t.Fatal(err)
	}

	// Set minipool withdrawable status
	if _, err := minipool.SubmitMinipoolWithdrawable(rp, mp.Address, trustedNodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

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

	// Call refund method to withdraw node's balance
	if _, err := mp.Refund(nodeAccount.GetTransactor()); err != nil {
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

func TestWithdrawValidatorBalanceAndFinalise(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Register nodes
	if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}
	if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil {
		t.Fatal(err)
	}

	// Create minipool
	mp, err := minipoolutils.CreateMinipool(t, rp, ownerAccount, nodeAccount, eth.EthToWei(16), 1)
	if err != nil {
		t.Fatal(err)
	}

	// Make user deposit
	userDepositAmount := eth.EthToWei(16)
	userDepositOpts := userAccount.GetTransactor()
	userDepositOpts.Value = userDepositAmount
	if _, err := deposit.Deposit(rp, userDepositOpts); err != nil {
		t.Fatal(err)
	}

	// Delay for the time between depositing and staking
	scrubPeriod, err := trustednode.GetScrubPeriod(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = evm.IncreaseTime(int(scrubPeriod + 1))
	if err != nil {
		t.Fatal(fmt.Errorf("error increasing time: %w", err))
	}

	// Stake minipool
	if err := minipoolutils.StakeMinipool(rp, mp, nodeAccount); err != nil {
		t.Fatal(err)
	}

	// Set minipool withdrawable status
	if _, err := minipool.SubmitMinipoolWithdrawable(rp, mp.Address, trustedNodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

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

	// Call DistributeBalanceAndFinalise method
	if _, err := mp.DistributeBalanceAndFinalise(nodeAccount.GetTransactor()); err != nil {
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
	} else if rethCollateralRate != 0.1 {
		t.Errorf("Incorrect rETH collateral rate %f", rethCollateralRate)
	}

	// Confirm the minipool still exists
	if exists, err := minipool.GetMinipoolExists(rp, mp.Address, nil); err != nil {
		t.Error(err)
	} else if !exists {
		t.Error("Minipool doesn't exist but it should")
	}

}

func TestDelegateUpgradeAndRollback(t *testing.T) {
	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Register nodes
	if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}
	if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil {
		t.Fatal(err)
	}

	// Create minipool
	mp, err := minipoolutils.CreateMinipool(t, rp, ownerAccount, nodeAccount, eth.EthToWei(16), 1)
	if err != nil {
		t.Fatal(err)
	}

	// Get original delegate contract
	originalDelegate, err := mp.GetEffectiveDelegate(nil)
	if err != nil {
		t.Fatal(err)
	}

	newDelegate := common.HexToAddress("0x1111111111111111111111111111111111111111")
	newAbi := "[{\"name\":\"foo\",\"type\":\"function\",\"inputs\":[],\"outputs\":[]}]"

	// Upgrade the network delegate contract
	_, err = trustednodedao.BootstrapUpgrade(rp, "upgradeContract", "rocketMinipoolDelegate", newAbi, newDelegate, ownerAccount.GetTransactor())
	if err != nil {
		t.Fatal(err)
	}

	// Get new effective delegate
	effectiveDelegate, err := mp.GetEffectiveDelegate(nil)
	if err != nil {
		t.Fatal(err)
	}

	// Check
	if effectiveDelegate != originalDelegate {
		t.Errorf("Effective delegate %s did not match original delegate %s", effectiveDelegate.Hex(), originalDelegate.Hex())
	}

	// Call upgrade
	if _, err := mp.DelegateUpgrade(nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Check effective delegate
	if effectiveDelegate, err = mp.GetEffectiveDelegate(nil); err != nil {
		t.Fatal(err)
	} else if effectiveDelegate != newDelegate {
		t.Errorf("Effective delegate %s did not match new delegate %s", effectiveDelegate.Hex(), newDelegate.Hex())
	}

	// Check previous delegate
	if previousDelegate, err := mp.GetPreviousDelegate(nil); err != nil {
		t.Fatal(err)
	} else if previousDelegate != originalDelegate {
		t.Errorf("Previous delegate %s did not match original delegate %s", previousDelegate.Hex(), originalDelegate.Hex())
	}

	// Check current delegate
	if currentDelegate, err := mp.GetDelegate(nil); err != nil {
		t.Fatal(err)
	} else if currentDelegate != newDelegate {
		t.Errorf("Current delegate %s did not match new delegate %s", currentDelegate.Hex(), newDelegate.Hex())
	}

	// Rollback
	if _, err := mp.DelegateRollback(nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get new effective delegate
	if effectiveDelegate, err = mp.GetEffectiveDelegate(nil); err != nil {
		t.Fatal(err)
	} else if effectiveDelegate != originalDelegate {
		t.Errorf("Effective delegate %s did not match original delegate %s", effectiveDelegate.Hex(), newDelegate.Hex())
	}
}

func TestUseLatestDelegate(t *testing.T) {
	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Register nodes
	if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}
	if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil {
		t.Fatal(err)
	}

	// Create minipool
	mp, err := minipoolutils.CreateMinipool(t, rp, ownerAccount, nodeAccount, eth.EthToWei(16), 1)
	if err != nil {
		t.Fatal(err)
	}

	// New delegate params
	newDelegate := common.HexToAddress("0x1111111111111111111111111111111111111111")
	newAbi := "[{\"name\":\"foo\",\"type\":\"function\",\"inputs\":[],\"outputs\":[]}]"

	// Upgrade the network delegate contract
	_, err = trustednodedao.BootstrapUpgrade(rp, "upgradeContract", "rocketMinipoolDelegate", newAbi, newDelegate, ownerAccount.GetTransactor())
	if err != nil {
		t.Fatal(err)
	}

	// Set use latest delegate
	if _, err = mp.SetUseLatestDelegate(true, nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get use latest delegate
	if useLatest, err := mp.GetUseLatestDelegate(nil); err != nil {
		t.Fatal(err)
	} else if !useLatest {
		t.Error("GetUseLatestDelegate returned false after being set")
	}

	// Check effective delegate
	if effectiveDelegate, err := mp.GetEffectiveDelegate(nil); err != nil {
		t.Fatal(err)
	} else if effectiveDelegate != newDelegate {
		t.Errorf("Effective delegate %s did not match new delegate %s", effectiveDelegate.Hex(), newDelegate.Hex())
	}
}
