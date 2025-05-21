package network

import (
	"testing"

	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
	nodeutils "github.com/rocket-pool/rocketpool-go/tests/testutils/node"
)

func TestSubmitPrices(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Register trusted node
	if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil {
		t.Fatal(err)
	}

	// Submit prices
	var pricesBlock uint64 = 100
	var slotTimestamp uint64 = 16000000
	rplPrice := eth.EthToWei(1000)
	effectiveRplStake := eth.EthToWei(24000)
	if _, err := network.SubmitPrices(rp, pricesBlock, slotTimestamp, rplPrice, effectiveRplStake, trustedNodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check network prices block
	if networkPricesBlock, err := network.GetPricesBlock(rp, nil); err != nil {
		t.Error(err)
	} else if networkPricesBlock != pricesBlock {
		t.Errorf("Incorrect network prices block %d", networkPricesBlock)
	}

	// Get & check network RPL price
	if networkRplPrice, err := network.GetRPLPrice(rp, nil); err != nil {
		t.Error(err)
	} else if networkRplPrice.Cmp(rplPrice) != 0 {
		t.Errorf("Incorrect network RPL price %s", networkRplPrice.String())
	}

}
