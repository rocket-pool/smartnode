package protocol

import (
	"testing"

	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
)

func TestAuctionSettings(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Set & get creat lots enabled
	createLotEnabled := false
	if _, err := protocol.BootstrapCreateLotEnabled(rp, createLotEnabled, ownerAccount.GetTransactor()); err != nil {
		t.Error(err)
	} else if value, err := protocol.GetCreateLotEnabled(rp, nil); err != nil {
		t.Error(err)
	} else if value != createLotEnabled {
		t.Error("Incorrect creat lots enabled value")
	}

	// Set & get bid on lot enabled
	bidOnLotEnabled := false
	if _, err := protocol.BootstrapBidOnLotEnabled(rp, bidOnLotEnabled, ownerAccount.GetTransactor()); err != nil {
		t.Error(err)
	} else if value, err := protocol.GetBidOnLotEnabled(rp, nil); err != nil {
		t.Error(err)
	} else if value != bidOnLotEnabled {
		t.Error("Incorrect bid on lot enabled value")
	}

	// Set & get lot minimum ETH value
	lotMinimumEthValue := eth.EthToWei(1000)
	if _, err := protocol.BootstrapLotMinimumEthValue(rp, lotMinimumEthValue, ownerAccount.GetTransactor()); err != nil {
		t.Error(err)
	} else if value, err := protocol.GetLotMinimumEthValue(rp, nil); err != nil {
		t.Error(err)
	} else if value.Cmp(lotMinimumEthValue) != 0 {
		t.Error("Incorrect lot minimum ETH value value")
	}

	// Set & get lot maximum ETH value
	lotMaximumEthValue := eth.EthToWei(0.01)
	if _, err := protocol.BootstrapLotMaximumEthValue(rp, lotMaximumEthValue, ownerAccount.GetTransactor()); err != nil {
		t.Error(err)
	} else if value, err := protocol.GetLotMaximumEthValue(rp, nil); err != nil {
		t.Error(err)
	} else if value.Cmp(lotMaximumEthValue) != 0 {
		t.Error("Incorrect lot maximum ETH value value")
	}

	// Set & get lot duration
	var lotDuration uint64 = 1
	if _, err := protocol.BootstrapLotDuration(rp, lotDuration, ownerAccount.GetTransactor()); err != nil {
		t.Error(err)
	} else if value, err := protocol.GetLotDuration(rp, nil); err != nil {
		t.Error(err)
	} else if value != lotDuration {
		t.Error("Incorrect lot duration value")
	}

	// Set & get lot starting price ratio
	lotStartingPriceRatio := 2.0
	if _, err := protocol.BootstrapLotStartingPriceRatio(rp, lotStartingPriceRatio, ownerAccount.GetTransactor()); err != nil {
		t.Error(err)
	} else if value, err := protocol.GetLotStartingPriceRatio(rp, nil); err != nil {
		t.Error(err)
	} else if value != lotStartingPriceRatio {
		t.Error("Incorrect lot starting price ratio value")
	}

	// Set & get lot reserve price ratio
	lotReservePriceRatio := 1.9
	if _, err := protocol.BootstrapLotReservePriceRatio(rp, lotReservePriceRatio, ownerAccount.GetTransactor()); err != nil {
		t.Error(err)
	} else if value, err := protocol.GetLotReservePriceRatio(rp, nil); err != nil {
		t.Error(err)
	} else if value != lotReservePriceRatio {
		t.Error("Incorrect lot reserve price ratio value")
	}

}
