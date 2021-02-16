package tokens

import (
    "testing"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/rocket-pool/rocketpool-go/utils/eth"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
    rplutils "github.com/rocket-pool/rocketpool-go/tests/testutils/tokens/rpl"
)


func TestRPLBalances(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Mint RPL
    rplAmount := eth.EthToWei(100)
    if err := rplutils.MintRPL(rp, ownerAccount, userAccount1, rplAmount); err != nil { t.Fatal(err) }

    // Get & check RPL account balance
    if rplBalance, err := tokens.GetRPLBalance(rp, userAccount1.Address, nil); err != nil {
        t.Error(err)
    } else if rplBalance.Cmp(rplAmount) != 0 {
        t.Errorf("Incorrect RPL account balance %s", rplBalance.String())
    }

    // Get & check RPL total supply
    initialTotalSupply := eth.EthToWei(18000000)
    if rplTotalSupply, err := tokens.GetRPLTotalSupply(rp, nil); err != nil {
        t.Error(err)
    } else if rplTotalSupply.Cmp(initialTotalSupply) != 0 {
        t.Errorf("Incorrect RPL total supply %s", rplTotalSupply.String())
    }

}


func TestTransferRPL(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Mint RPL
    rplAmount := eth.EthToWei(100)
    if err := rplutils.MintRPL(rp, ownerAccount, userAccount1, rplAmount); err != nil { t.Fatal(err) }

    // Transfer RPL
    toAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
    sendAmount := eth.EthToWei(50)
    if _, err := tokens.TransferRPL(rp, toAddress, sendAmount, userAccount1.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check RPL account balance
    if rplBalance, err := tokens.GetRPLBalance(rp, toAddress, nil); err != nil {
        t.Error(err)
    } else if rplBalance.Cmp(sendAmount) != 0 {
        t.Errorf("Incorrect RPL account balance %s", rplBalance.String())
    }

}


func TestTransferFromRPL(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Mint RPL
    rplAmount := eth.EthToWei(100)
    if err := rplutils.MintRPL(rp, ownerAccount, userAccount1, rplAmount); err != nil { t.Fatal(err) }

    // Approve RPL spender
    sendAmount := eth.EthToWei(50)
    if _, err := tokens.ApproveRPL(rp, userAccount2.Address, sendAmount, userAccount1.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check spender allowance
    if allowance, err := tokens.GetRPLAllowance(rp, userAccount1.Address, userAccount2.Address, nil); err != nil {
        t.Error(err)
    } else if allowance.Cmp(sendAmount) != 0 {
        t.Errorf("Incorrect RPL spender allowance %s", allowance.String())
    }

    // Transfer RPL from account
    toAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
    if _, err := tokens.TransferFromRPL(rp, userAccount1.Address, toAddress, sendAmount, userAccount2.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check RPL account balance
    if rplBalance, err := tokens.GetRPLBalance(rp, toAddress, nil); err != nil {
        t.Error(err)
    } else if rplBalance.Cmp(sendAmount) != 0 {
        t.Errorf("Incorrect RPL account balance %s", rplBalance.String())
    }

}

