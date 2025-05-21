package node

import (
	"testing"

	"github.com/rocket-pool/rocketpool-go/node"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
)

func TestNodeDistributor(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	distributorAddress, err := node.GetDistributorAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		t.Fatal(err)
	}

	if distributorAddress.Hex() == "0x0000000000000000000000000000000000000000" {
		t.Errorf("Invalid distributor address")
	}
}
