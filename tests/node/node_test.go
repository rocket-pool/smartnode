package node

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/storage"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
)

func TestRegisterNode(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Get & check initial node exists status
	if exists, err := node.GetNodeExists(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if exists {
		t.Error("Node already existed before registration")
	}

	// Get & check initial node details
	if details, err := node.GetNodes(rp, nil); err != nil {
		t.Error(err)
	} else if len(details) != 0 {
		t.Error("Incorrect initial node count")
	}

	// Register node
	timezoneLocation := "Australia/Brisbane"
	if _, err := node.RegisterNode(rp, timezoneLocation, nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check updated node details
	if details, err := node.GetNodes(rp, nil); err != nil {
		t.Error(err)
	} else if len(details) != 1 {
		t.Error("Incorrect updated node count")
	} else {
		nodeDetails := details[0]
		if !bytes.Equal(nodeDetails.Address.Bytes(), nodeAccount.Address.Bytes()) {
			t.Errorf("Incorrect node address %s", nodeDetails.Address.Hex())
		}
		if !nodeDetails.Exists {
			t.Error("Incorrect node exists status")
		}
		if !bytes.Equal(nodeDetails.PrimaryWithdrawalAddress.Bytes(), nodeAccount.Address.Bytes()) {
			t.Errorf("Incorrect node withdrawal address '%s'", nodeDetails.PrimaryWithdrawalAddress.Hex())
		}
		if nodeDetails.TimezoneLocation != timezoneLocation {
			t.Errorf("Incorrect node timezone location '%s'", nodeDetails.TimezoneLocation)
		}
	}

}

func TestSetWithdrawalAddress(t *testing.T) {

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

	// Set withdrawal address
	withdrawalAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
	if _, err := storage.SetWithdrawalAddress(rp, nodeAccount.Address, withdrawalAddress, true, nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check node withdrawal address
	if nodeWithdrawalAddress, err := storage.GetNodeWithdrawalAddress(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if !bytes.Equal(nodeWithdrawalAddress.Bytes(), withdrawalAddress.Bytes()) {
		t.Errorf("Incorrect node withdrawal address '%s'", nodeWithdrawalAddress.Hex())
	}

}

func TestSetWithdrawalAddressConfirmation(t *testing.T) {

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

	// Set withdrawal address
	withdrawalAddress := withdrawalAccount.Address
	if _, err := storage.SetWithdrawalAddress(rp, nodeAccount.Address, withdrawalAddress, false, nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Confirm withdrawal address
	if _, err := storage.ConfirmWithdrawalAddress(rp, nodeAccount.Address, withdrawalAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check node withdrawal address
	if nodeWithdrawalAddress, err := storage.GetNodeWithdrawalAddress(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if !bytes.Equal(nodeWithdrawalAddress.Bytes(), withdrawalAddress.Bytes()) {
		t.Errorf("Incorrect node withdrawal address '%s'", nodeWithdrawalAddress.Hex())
	}

}

func TestSetTimezoneLocation(t *testing.T) {

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

	// Set timezone
	timezoneLocation := "Australia/Sydney"
	if _, err := node.SetTimezoneLocation(rp, timezoneLocation, nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check node timezone location
	if nodeTimezoneLocation, err := node.GetNodeTimezoneLocation(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if nodeTimezoneLocation != timezoneLocation {
		t.Errorf("Incorrect node timezone location '%s'", nodeTimezoneLocation)
	}

}
