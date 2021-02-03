package settings

import (
    "testing"

    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


func TestMinipoolSettings(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Get & check launch balance and deposit amounts
    fullMinipoolBalance := eth.EthToWei(32)
    halfMinipoolBalance := eth.EthToWei(16)
    emptyMinipoolBalance := eth.EthToWei(0)
    if value, err := settings.GetMinipoolLaunchBalance(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(fullMinipoolBalance) != 0 {
        t.Error("Incorrect minipool launch balance")
    }
    if value, err := settings.GetMinipoolFullDepositNodeAmount(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(fullMinipoolBalance) != 0 {
        t.Error("Incorrect minipool full deposit node amount")
    }
    if value, err := settings.GetMinipoolHalfDepositNodeAmount(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(halfMinipoolBalance) != 0 {
        t.Error("Incorrect minipool half deposit node amount")
    }
    if value, err := settings.GetMinipoolEmptyDepositNodeAmount(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(emptyMinipoolBalance) != 0 {
        t.Error("Incorrect minipool empty deposit node amount")
    }
    if value, err := settings.GetMinipoolFullDepositUserAmount(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(halfMinipoolBalance) != 0 {
        t.Error("Incorrect minipool full deposit user amount")
    }
    if value, err := settings.GetMinipoolHalfDepositUserAmount(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(halfMinipoolBalance) != 0 {
        t.Error("Incorrect minipool half deposit user amount")
    }
    if value, err := settings.GetMinipoolEmptyDepositUserAmount(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(fullMinipoolBalance) != 0 {
        t.Error("Incorrect minipool empty deposit user amount")
    }

    // Set & get submit withdrawable enabled
    submitWithdrawableEnabled := false
    if _, err := settings.SetMinipoolSubmitWithdrawableEnabled(rp, submitWithdrawableEnabled, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetMinipoolSubmitWithdrawableEnabled(rp, nil); err != nil {
        t.Error(err)
    } else if value != submitWithdrawableEnabled {
        t.Error("Incorrect minipool withdrawable submissions enabled value")
    }

    // Set & get minipool launch timeout
    var minipoolLaunchTimeout uint64 = 5
    if _, err := settings.SetMinipoolLaunchTimeout(rp, minipoolLaunchTimeout, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetMinipoolLaunchTimeout(rp, nil); err != nil {
        t.Error(err)
    } else if value != minipoolLaunchTimeout {
        t.Error("Incorrect minipool launch timeout value")
    }

    // Set & get minipool withdrawal delay
    var minipoolWithdrawalDelay uint64 = 5
    if _, err := settings.SetMinipoolWithdrawalDelay(rp, minipoolWithdrawalDelay, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetMinipoolWithdrawalDelay(rp, nil); err != nil {
        t.Error(err)
    } else if value != minipoolWithdrawalDelay {
        t.Error("Incorrect minipool withdrawal delay value")
    }

}

