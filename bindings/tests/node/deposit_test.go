package node

import (
	"testing"

	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
	minipoolutils "github.com/rocket-pool/rocketpool-go/tests/testutils/minipool"
	nodeutils "github.com/rocket-pool/rocketpool-go/tests/testutils/node"
)

func TestDeposit(t *testing.T) {

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

	// Get initial node minipool count
	minipoolCount1, err := minipool.GetNodeMinipoolCount(rp, nodeAccount.Address, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Mint & stake RPL required for mininpool
	rplRequired, err := minipoolutils.GetMinipoolRPLRequired(rp)
	if err != nil {
		t.Fatal(err)
	}
	if err := nodeutils.StakeRPL(rp, ownerAccount, nodeAccount, rplRequired); err != nil {
		t.Fatal(err)
	}

	// Deposit
	if _, _, err := nodeutils.Deposit(t, rp, nodeAccount, eth.EthToWei(16), 1); err != nil {
		t.Fatal(err)
	}

	// Get & check updated node minipool count
	minipoolCount2, err := minipool.GetNodeMinipoolCount(rp, nodeAccount.Address, nil)
	if err != nil {
		t.Fatal(err)
	} else if minipoolCount2 != minipoolCount1+1 {
		t.Error("Incorrect node minipool count")
	}

}
