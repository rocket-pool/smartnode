package protocol

import (
	"testing"

	"github.com/rocket-pool/rocketpool-go/settings/protocol"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
)

func TestNodeSettings(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Set & get node registrations enabled
	nodeRegistrationsEnabled := false
	if _, err := protocol.BootstrapNodeRegistrationEnabled(rp, nodeRegistrationsEnabled, ownerAccount.GetTransactor()); err != nil {
		t.Error(err)
	} else if value, err := protocol.GetNodeRegistrationEnabled(rp, nil); err != nil {
		t.Error(err)
	} else if value != nodeRegistrationsEnabled {
		t.Error("Incorrect node registrations enabled value")
	}

	// Set & get node deposits enabled
	nodeDepositsEnabled := false
	if _, err := protocol.BootstrapNodeDepositEnabled(rp, nodeDepositsEnabled, ownerAccount.GetTransactor()); err != nil {
		t.Error(err)
	} else if value, err := protocol.GetNodeDepositEnabled(rp, nil); err != nil {
		t.Error(err)
	} else if value != nodeDepositsEnabled {
		t.Error("Incorrect node deposits enabled value")
	}

	// Set & get minimum per minipool RPL stake
	minimumPerMinipoolStake := 1.0
	if _, err := protocol.BootstrapMinimumPerMinipoolStake(rp, minimumPerMinipoolStake, ownerAccount.GetTransactor()); err != nil {
		t.Error(err)
	} else if value, err := protocol.GetMinimumPerMinipoolStake(rp, nil); err != nil {
		t.Error(err)
	} else if value != minimumPerMinipoolStake {
		t.Error("Incorrect minimum per minipool stake value")
	}

	// Set & get maximum per minipool RPL stake
	maximumPerMinipoolStake := 10.0
	if _, err := protocol.BootstrapMaximumPerMinipoolStake(rp, maximumPerMinipoolStake, ownerAccount.GetTransactor()); err != nil {
		t.Error(err)
	} else if value, err := protocol.GetMaximumPerMinipoolStake(rp, nil); err != nil {
		t.Error(err)
	} else if value != maximumPerMinipoolStake {
		t.Error("Incorrect maximum per minipool stake value")
	}

}
