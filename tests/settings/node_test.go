package settings

import (
    "testing"

    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/rocket-pool/rocketpool-go/tests/utils/evm"
)


func TestNodeSettings(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Set & get node registrations enabled
    nodeRegistrationsEnabled := false
    if _, err := settings.SetNodeRegistrationEnabled(rp, nodeRegistrationsEnabled, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetNodeRegistrationEnabled(rp, nil); err != nil {
        t.Error(err)
    } else if value != nodeRegistrationsEnabled {
        t.Error("Incorrect node registrations enabled value")
    }

    // Set & get node deposits enabled
    nodeDepositsEnabled := false
    if _, err := settings.SetNodeDepositEnabled(rp, nodeDepositsEnabled, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetNodeDepositEnabled(rp, nil); err != nil {
        t.Error(err)
    } else if value != nodeDepositsEnabled {
        t.Error("Incorrect node deposits enabled value")
    }

}

