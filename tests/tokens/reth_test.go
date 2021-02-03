package tokens

import (
    "testing"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/rocketpool-go/network"
    "github.com/rocket-pool/rocketpool-go/tests/utils/evm"
    nodeutils "github.com/rocket-pool/rocketpool-go/tests/utils/node"
    tokenutils "github.com/rocket-pool/rocketpool-go/tests/utils/tokens"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


func TestRETHBalances(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Mint rETH
    rethAmount := eth.EthToWei(100)
    if err := tokenutils.MintRETH(rp, userAccount, rethAmount); err != nil { t.Fatal(err) }

    // Get & check rETH total supply
    if rethTotalSupply, err := tokens.GetRETHTotalSupply(rp, nil); err != nil {
        t.Error(err)
    } else if rethTotalSupply.Cmp(rethAmount) != 0 {
        t.Errorf("Incorrect rETH total supply %s", rethTotalSupply.String())
    }

    // Get & check rETH account balance
    if rethBalance, err := tokens.GetRETHBalance(rp, userAccount.Address, nil); err != nil {
        t.Error(err)
    } else if rethBalance.Cmp(rethAmount) != 0 {
        t.Errorf("Incorrect rETH account balance %s", rethBalance.String())
    }

}


func TestTransferRETH(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Mint rETH
    rethAmount := eth.EthToWei(100)
    if err := tokenutils.MintRETH(rp, userAccount, rethAmount); err != nil { t.Fatal(err) }

    // Transfer rETH
    toAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
    sendAmount := eth.EthToWei(50)
    if _, err := tokens.TransferRETH(rp, toAddress, sendAmount, userAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check rETH account balance
    if rethBalance, err := tokens.GetRETHBalance(rp, toAddress, nil); err != nil {
        t.Error(err)
    } else if rethBalance.Cmp(sendAmount) != 0 {
        t.Errorf("Incorrect rETH account balance %s", rethBalance.String())
    }

}


func TestBurnRETH(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Mint rETH
    rethAmount := eth.EthToWei(100)
    if err := tokenutils.MintRETH(rp, userAccount, rethAmount); err != nil { t.Fatal(err) }

    // Get initial balances
    balances1, err := tokens.GetBalances(rp, userAccount.Address, nil)
    if err != nil {
        t.Fatal(err)
    }

    // Burn rETH
    burnAmount := eth.EthToWei(50)
    if _, err := tokens.BurnRETH(rp, burnAmount, userAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check updated balances
    balances2, err := tokens.GetBalances(rp, userAccount.Address, nil)
    if err != nil {
        t.Fatal(err)
    } else {
        if balances2.RETH.Cmp(balances1.RETH) != -1 {
            t.Error("rETH balance did not decrease after burning rETH")
        }
        if balances2.ETH.Cmp(balances1.ETH) != 1 {
            t.Error("ETH balance did not increase after burning rETH")
        }
    }

}


func TestRETHExchangeRate(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register trusted node
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Submit network balances
    if _, err := network.SubmitBalances(rp, 1, eth.EthToWei(100), eth.EthToWei(100), eth.EthToWei(50), trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get & check ETH value of rETH amount
    rethAmount := eth.EthToWei(1)
    if ethValue, err := tokens.GetETHValueOfRETH(rp, rethAmount, nil); err != nil {
        t.Error(err)
    } else if ethValue.Cmp(eth.EthToWei(2)) != 0 {
        t.Errorf("Incorrect ETH value %s of rETH amount %s", ethValue.String(), rethAmount.String())
    }

    // Get & check rETH value of ETH amount
    ethAmount := eth.EthToWei(2)
    if rethValue, err := tokens.GetRETHValueOfETH(rp, ethAmount, nil); err != nil {
        t.Error(err)
    } else if rethValue.Cmp(eth.EthToWei(1)) != 0 {
        t.Errorf("Incorrect rETH value %s of ETH amount %s", rethValue.String(), ethAmount.String())
    }

    // Get & check ETH : rETH exchange rate
    if exchangeRate, err := tokens.GetRETHExchangeRate(rp, nil); err != nil {
        t.Error(err)
    } else if exchangeRate != 2 {
        t.Errorf("Incorrect ETH : rETH exchange rate %f : 1", exchangeRate)
    }

}

