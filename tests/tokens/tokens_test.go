package tokens

import (
    "log"
    "os"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/rocketpool-go/network"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/tests"
    "github.com/rocket-pool/rocketpool-go/tests/utils/accounts"
    "github.com/rocket-pool/rocketpool-go/tests/utils/evm"
    nodeutils "github.com/rocket-pool/rocketpool-go/tests/utils/node"
    tokenutils "github.com/rocket-pool/rocketpool-go/tests/utils/tokens"
    "github.com/rocket-pool/rocketpool-go/tests/utils/validator"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


var (
    client *ethclient.Client
    rp *rocketpool.RocketPool

    ownerAccount *accounts.Account
    trustedNodeAccount *accounts.Account
    userAccount *accounts.Account
)


func TestMain(m *testing.M) {
    var err error

    // Initialize eth client
    client, err = ethclient.Dial(tests.Eth1ProviderAddress)
    if err != nil { log.Fatal(err) }

    // Initialize contract manager
    rp, err = rocketpool.NewRocketPool(client, common.HexToAddress(tests.RocketStorageAddress))
    if err != nil { log.Fatal(err) }

    // Initialize accounts
    ownerAccount, err = accounts.GetAccount(0)
    if err != nil { log.Fatal(err) }
    trustedNodeAccount, err = accounts.GetAccount(1)
    if err != nil { log.Fatal(err) }
    userAccount, err = accounts.GetAccount(9)
    if err != nil { log.Fatal(err) }

    // Run tests
    os.Exit(m.Run())

}


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

