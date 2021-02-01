package network

import (
    "bytes"
    "testing"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/rocketpool-go/deposit"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/network"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/tests/utils/evm"
    minipoolutils "github.com/rocket-pool/rocketpool-go/tests/utils/minipool"
    nodeutils "github.com/rocket-pool/rocketpool-go/tests/utils/node"
    "github.com/rocket-pool/rocketpool-go/tests/utils/validator"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


func TestSetWithdrawalCredentials(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Set withdrawal credentials
    withdrawalCredentials := common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111")
    if _, err := network.SetWithdrawalCredentials(rp, withdrawalCredentials, ownerAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check withdrawal credentials
    if networkWithdrawalCredentials, err := network.GetWithdrawalCredentials(rp, nil); err != nil {
        t.Error(err)
    } else if !bytes.Equal(networkWithdrawalCredentials.Bytes(), withdrawalCredentials.Bytes()) {
        t.Errorf("Incorrect network withdrawal credentials %s", networkWithdrawalCredentials.Hex())
    }

}


func TestTransferWithdrawal(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Transfer validator balance
    opts := userAccount.GetTransactor()
    opts.Value = eth.EthToWei(50)
    if _, err := network.TransferWithdrawal(rp, opts); err != nil {
        t.Fatal(err)
    }

    // Get & check withdrawal contract balance
    if withdrawalBalance, err := network.GetWithdrawalBalance(rp, nil); err != nil {
        t.Error(err)
    } else if withdrawalBalance.Cmp(opts.Value) != 0 {
        t.Errorf("Incorrect withdrawal contract balance %s", withdrawalBalance.String())
    }

}


func TestProcessWithdrawal(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register nodes
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Create minipool
    mp, err := minipoolutils.CreateMinipool(rp, nodeAccount, eth.EthToWei(16))
    if err != nil { t.Fatal(err) }

    // Make user deposit
    userDepositOpts := userAccount.GetTransactor()
    userDepositOpts.Value = eth.EthToWei(16)
    if _, err := deposit.Deposit(rp, userDepositOpts); err != nil { t.Fatal(err) }

    // Stake minipool
    if err := minipoolutils.StakeMinipool(rp, mp, nodeAccount); err != nil { t.Fatal(err) }

    // Mark minipool as withdrawable
    if _, err := minipool.SubmitMinipoolWithdrawable(rp, mp.Address, eth.EthToWei(32), eth.EthToWei(32), trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Transfer validator balance
    transferWithdrawalOpts := userAccount.GetTransactor()
    transferWithdrawalOpts.Value = eth.EthToWei(32)
    if _, err := network.TransferWithdrawal(rp, transferWithdrawalOpts); err != nil { t.Fatal(err) }

    // Get initial token contract ETH balances
    nethContractBalance1, err := tokens.GetNETHContractETHBalance(rp, nil)
    if err != nil {
        t.Fatal(err)
    }
    rethContractBalance1, err := tokens.GetRETHContractETHBalance(rp, nil)
    if err != nil {
        t.Fatal(err)
    }

    // Process withdrawal
    validatorPubkey, err := validator.GetValidatorPubkey()
    if err != nil { t.Fatal(err) }
    if _, err := network.ProcessWithdrawal(rp, validatorPubkey, trustedNodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check updated token contract ETH balances
    if nethContractBalance2, err := tokens.GetNETHContractETHBalance(rp, nil); err != nil {
        t.Fatal(err)
    } else if nethContractBalance2.Cmp(nethContractBalance1) != 1 {
        t.Error("nETH contract ETH balance did not increase after processing withdrawal")
    }
    if rethContractBalance2, err := tokens.GetRETHContractETHBalance(rp, nil); err != nil {
        t.Fatal(err)
    } else if rethContractBalance2.Cmp(rethContractBalance1) != 1 {
        t.Error("rETH contract ETH balance did not increase after processing withdrawal")
    }

}

