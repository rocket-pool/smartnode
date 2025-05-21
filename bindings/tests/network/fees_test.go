package network

import (
	"math/big"
	"testing"

	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
)

func TestNodeFee(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Get settings
	targetNodeFee, err := protocol.GetTargetNodeFee(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	minNodeFee, err := protocol.GetMinimumNodeFee(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	maxNodeFee, err := protocol.GetMaximumNodeFee(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	demandRange, err := protocol.GetNodeFeeDemandRange(rp, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Get & check initial node demand
	if nodeDemand, err := network.GetNodeDemand(rp, nil); err != nil {
		t.Error(err)
	} else if nodeDemand.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Incorrect initial node demand value %s", nodeDemand.String())
	}

	// Get & check initial node fee
	if nodeFee, err := network.GetNodeFee(rp, nil); err != nil {
		t.Error(err)
	} else if nodeFee != targetNodeFee {
		t.Errorf("Incorrect initial node fee %f", nodeFee)
	}

	// Make user deposit
	opts := userAccount.GetTransactor()
	opts.Value = demandRange
	if _, err := deposit.Deposit(rp, opts); err != nil {
		t.Fatal(err)
	}

	// Get & check updated node demand
	if nodeDemand, err := network.GetNodeDemand(rp, nil); err != nil {
		t.Error(err)
	} else if nodeDemand.Cmp(opts.Value) != 0 {
		t.Errorf("Incorrect updated node demand value %s", nodeDemand.String())
	}

	// Get & check updated node fee
	if nodeFee, err := network.GetNodeFee(rp, nil); err != nil {
		t.Error(err)
	} else if nodeFee != maxNodeFee {
		t.Errorf("Incorrect updated node fee %f", nodeFee)
	}

	// Get & check node fees by demand values
	negDemandRange := new(big.Int)
	negDemandRange.Neg(demandRange)
	if nodeFee, err := network.GetNodeFeeByDemand(rp, big.NewInt(0), nil); err != nil {
		t.Error(err)
	} else if nodeFee != targetNodeFee {
		t.Errorf("Incorrect node fee for zero demand %f", nodeFee)
	}
	if nodeFee, err := network.GetNodeFeeByDemand(rp, negDemandRange, nil); err != nil {
		t.Error(err)
	} else if nodeFee != minNodeFee {
		t.Errorf("Incorrect node fee for negative demand %f", nodeFee)
	}
	if nodeFee, err := network.GetNodeFeeByDemand(rp, demandRange, nil); err != nil {
		t.Error(err)
	} else if nodeFee != maxNodeFee {
		t.Errorf("Incorrect node fee for positive demand %f", nodeFee)
	}

}
