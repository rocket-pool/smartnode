package tokens

import (
    "testing"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/rocketpool-go/network"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/rocket-pool/rocketpool-go/utils/eth"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
    nodeutils "github.com/rocket-pool/rocketpool-go/tests/testutils/node"
    tokenutils "github.com/rocket-pool/rocketpool-go/tests/testutils/tokens"
    "github.com/rocket-pool/rocketpool-go/tests/testutils/validator"
)


// GetNETHContractETHBalance test under network.TestProcessWithdrawal


func TestNETHBalances(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Mint nETH
    nethAmount := eth.EthToWei(100)
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }
    if err := tokenutils.MintNETH(rp, ownerAccount, trustedNodeAccount, userAccount, nethAmount); err != nil { t.Fatal(err) }

    // Get & check nETH total supply
    if nethTotalSupply, err := tokens.GetNETHTotalSupply(rp, nil); err != nil {
        t.Error(err)
    } else if nethTotalSupply.Cmp(nethAmount) != 0 {
        t.Errorf("Incorrect nETH total supply %s", nethTotalSupply.String())
    }

    // Get & check nETH account balance
    if nethBalance, err := tokens.GetNETHBalance(rp, userAccount.Address, nil); err != nil {
        t.Error(err)
    } else if nethBalance.Cmp(nethAmount) != 0 {
        t.Errorf("Incorrect nETH account balance %s", nethBalance.String())
    }

}


func TestTransferNETH(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Mint nETH
    nethAmount := eth.EthToWei(100)
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }
    if err := tokenutils.MintNETH(rp, ownerAccount, trustedNodeAccount, userAccount, nethAmount); err != nil { t.Fatal(err) }

    // Transfer nETH
    toAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
    sendAmount := eth.EthToWei(50)
    if _, err := tokens.TransferNETH(rp, toAddress, sendAmount, userAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check nETH account balance
    if nethBalance, err := tokens.GetNETHBalance(rp, toAddress, nil); err != nil {
        t.Error(err)
    } else if nethBalance.Cmp(sendAmount) != 0 {
        t.Errorf("Incorrect nETH account balance %s", nethBalance.String())
    }

}


func TestBurnNETH(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Mint nETH
    nethAmount := eth.EthToWei(100)
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }
    if err := tokenutils.MintNETH(rp, ownerAccount, trustedNodeAccount, userAccount, nethAmount); err != nil { t.Fatal(err) }

    // Transfer validator balance
    opts := userAccount.GetTransactor()
    opts.Value = nethAmount
    if _, err := network.TransferWithdrawal(rp, opts); err != nil { t.Fatal(err) }

    // Process validator withdrawal
    validatorPubkey, err := validator.GetValidatorPubkey()
    if err != nil { t.Fatal(err) }
    if _, err := network.ProcessWithdrawal(rp, validatorPubkey, trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get initial balances
    balances1, err := tokens.GetBalances(rp, userAccount.Address, nil)
    if err != nil {
        t.Fatal(err)
    }

    // Burn nETH
    burnAmount := eth.EthToWei(50)
    if _, err := tokens.BurnNETH(rp, burnAmount, userAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check updated balances
    balances2, err := tokens.GetBalances(rp, userAccount.Address, nil)
    if err != nil {
        t.Fatal(err)
    } else {
        if balances2.NETH.Cmp(balances1.NETH) != -1 {
            t.Error("nETH balance did not decrease after burning nETH")
        }
        if balances2.ETH.Cmp(balances1.ETH) != 1 {
            t.Error("ETH balance did not increase after burning nETH")
        }
    }

}

