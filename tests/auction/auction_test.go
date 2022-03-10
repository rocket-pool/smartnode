package auction

import (
	"math/big"
	"testing"

	"github.com/rocket-pool/rocketpool-go/settings/trustednode"

	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	auctionutils "github.com/rocket-pool/rocketpool-go/tests/testutils/auction"
	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
	nodeutils "github.com/rocket-pool/rocketpool-go/tests/testutils/node"
)

func TestAuctionDetails(t *testing.T) {

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
	if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount1); err != nil {
		t.Fatal(err)
	}
	if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount2); err != nil {
		t.Fatal(err)
	}
	if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount3); err != nil {
		t.Fatal(err)
	}

	// Disable min commission rate for unbonded pools
	if _, err := trustednode.BootstrapMinipoolUnbondedMinFee(rp, uint64(0), ownerAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check initial RPL balances
	totalBalance1, err := auction.GetTotalRPLBalance(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	allottedBalance1, err := auction.GetAllottedRPLBalance(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	remainingBalance1, err := auction.GetRemainingRPLBalance(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	if totalBalance1.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Incorrect initial auction contract total RPL balance %s", totalBalance1.String())
	}
	if allottedBalance1.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Incorrect initial auction contract allotted RPL balance %s", allottedBalance1.String())
	}
	if remainingBalance1.Cmp(totalBalance1) != 0 {
		t.Errorf("Incorrect initial auction contract remaining RPL balance %s", remainingBalance1.String())
	}

	// Mint slashed RPL to auction contract
	if err := auctionutils.CreateSlashedRPL(t, rp, ownerAccount, trustedNodeAccount1, trustedNodeAccount2, userAccount1); err != nil {
		t.Fatal(err)
	}

	// Get & check updated RPL balances
	totalBalance2, err := auction.GetTotalRPLBalance(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	allottedBalance2, err := auction.GetAllottedRPLBalance(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	remainingBalance2, err := auction.GetRemainingRPLBalance(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	if totalBalance2.Cmp(big.NewInt(0)) != 1 {
		t.Errorf("Incorrect updated auction contract total RPL balance 1 %s", totalBalance2.String())
	}
	if allottedBalance2.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Incorrect updated auction contract allotted RPL balance 1 %s", allottedBalance2.String())
	}
	if remainingBalance2.Cmp(totalBalance2) != 0 {
		t.Errorf("Incorrect updated auction contract remaining RPL balance 1 %s", remainingBalance2.String())
	}

	// Create a new lot
	if _, _, err := auction.CreateLot(rp, userAccount1.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check updated RPL balances
	totalBalance3, err := auction.GetTotalRPLBalance(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	allottedBalance3, err := auction.GetAllottedRPLBalance(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	remainingBalance3, err := auction.GetRemainingRPLBalance(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	var expectedRemainingBalance big.Int
	expectedRemainingBalance.Sub(totalBalance3, allottedBalance3)
	if allottedBalance3.Cmp(big.NewInt(0)) != 1 {
		t.Errorf("Incorrect updated auction contract allotted RPL balance 2 %s", allottedBalance3.String())
	}
	if remainingBalance3.Cmp(&expectedRemainingBalance) != 0 {
		t.Errorf("Incorrect updated auction contract remaining RPL balance 2 %s", remainingBalance3.String())
	}

}

func TestLotDetails(t *testing.T) {

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
	if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount1); err != nil {
		t.Fatal(err)
	}
	if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount2); err != nil {
		t.Fatal(err)
	}
	if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount3); err != nil {
		t.Fatal(err)
	}

	// Disable min commission rate for unbonded pools
	if _, err := trustednode.BootstrapMinipoolUnbondedMinFee(rp, uint64(0), ownerAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Set network parameters
	if _, err := network.SubmitPrices(rp, 1, eth.EthToWei(1), eth.EthToWei(24), trustedNodeAccount1.GetTransactor()); err != nil {
		t.Fatal(err)
	}
	if _, err := network.SubmitPrices(rp, 1, eth.EthToWei(1), eth.EthToWei(24), trustedNodeAccount2.GetTransactor()); err != nil {
		t.Fatal(err)
	}
	if _, err := protocol.BootstrapLotStartingPriceRatio(rp, 1.0, ownerAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}
	if _, err := protocol.BootstrapLotReservePriceRatio(rp, 0.5, ownerAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}
	if _, err := protocol.BootstrapLotMaximumEthValue(rp, eth.EthToWei(10), ownerAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}
	if _, err := protocol.BootstrapLotDuration(rp, 5, ownerAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Mint slashed RPL to auction contract
	if err := auctionutils.CreateSlashedRPL(t, rp, ownerAccount, trustedNodeAccount1, trustedNodeAccount2, userAccount1); err != nil {
		t.Fatal(err)
	}

	// Get & check initial lot details
	if lots, err := auction.GetLots(rp, nil); err != nil {
		t.Error(err)
	} else if len(lots) != 0 {
		t.Error("Incorrect initial lot count")
	}
	if lots, err := auction.GetLotsWithBids(rp, userAccount1.Address, nil); err != nil {
		t.Error(err)
	} else if len(lots) != 0 {
		t.Error("Incorrect initial lot count")
	}

	// Create lots
	lot1Index, _, err := auction.CreateLot(rp, userAccount1.GetTransactor())
	if err != nil {
		t.Fatal(err)
	}
	lot2Index, _, err := auction.CreateLot(rp, userAccount1.GetTransactor())
	if err != nil {
		t.Fatal(err)
	}

	// Place bid on lot 1
	bidAmount := eth.EthToWei(1)
	bid1Opts := userAccount1.GetTransactor()
	bid1Opts.Value = bidAmount
	if _, err := auction.PlaceBid(rp, lot1Index, bid1Opts); err != nil {
		t.Fatal(err)
	}

	// Place another bid on lot 1 to clear it
	bid2Opts := userAccount2.GetTransactor()
	bid2Opts.Value = eth.EthToWei(1000)
	if _, err := auction.PlaceBid(rp, lot1Index, bid2Opts); err != nil {
		t.Fatal(err)
	}

	// Mine blocks until lot 2 hits reserve price & recover unclaimed RPL from it
	if err := evm.MineBlocks(5); err != nil {
		t.Fatal(err)
	}
	if _, err := auction.RecoverUnclaimedRPL(rp, lot2Index, userAccount1.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check updated lot details
	if lots, err := auction.GetLots(rp, nil); err != nil {
		t.Error(err)
	} else if len(lots) != 2 {
		t.Error("Incorrect updated lot count")
	} else if lots[0].Index != lot1Index || lots[1].Index != lot2Index {
		t.Error("Incorrect lot indexes")
	}
	if lots, err := auction.GetLotsWithBids(rp, userAccount1.Address, nil); err != nil {
		t.Error(err)
	} else if len(lots) != 2 {
		t.Error("Incorrect updated lot count")
	} else {
		lot1 := lots[0]
		lot2 := lots[1]

		// Lot 1
		if lot1.Index != lot1Index {
			t.Errorf("Incorrect lot index %d", lot1.Index)
		}
		if !lot1.Exists {
			t.Error("Incorrect lot exists status")
		}
		if lot1.StartBlock == 0 {
			t.Errorf("Incorrect lot start block %d", lot1.StartBlock)
		}
		if lot1.EndBlock <= lot1.StartBlock {
			t.Errorf("Incorrect lot end block %d", lot1.EndBlock)
		}
		if lot1.StartPrice.Cmp(eth.EthToWei(1)) != 0 {
			t.Errorf("Incorrect lot start price %s", lot1.StartPrice.String())
		}
		if lot1.ReservePrice.Cmp(eth.EthToWei(0.5)) != 0 {
			t.Errorf("Incorrect lot reserve price %s", lot1.ReservePrice.String())
		}
		if lot1.PriceAtCurrentBlock.Cmp(lot1.StartPrice) == 1 || lot1.PriceAtCurrentBlock.Cmp(lot1.ReservePrice) == -1 {
			t.Errorf("Incorrect lot price at current block %s", lot1.PriceAtCurrentBlock.String())
		}
		if lot1.PriceByTotalBids.Cmp(lot1.StartPrice) == 1 || lot1.PriceByTotalBids.Cmp(lot1.ReservePrice) == -1 {
			t.Errorf("Incorrect lot price at current block %s", lot1.PriceByTotalBids.String())
		}
		if lot1.CurrentPrice.Cmp(lot1.StartPrice) == 1 || lot1.CurrentPrice.Cmp(lot1.ReservePrice) == -1 {
			t.Errorf("Incorrect lot price at current block %s", lot1.CurrentPrice.String())
		}
		if lot1.TotalRPLAmount.Cmp(eth.EthToWei(10)) != 0 {
			t.Errorf("Incorrect lot total RPL amount %s", lot1.TotalRPLAmount.String())
		}
		if lot1.ClaimedRPLAmount.Cmp(eth.EthToWei(10)) != 0 {
			t.Errorf("Incorrect lot claimed RPL amount %s", lot1.ClaimedRPLAmount.String())
		}
		if lot1.RemainingRPLAmount.Cmp(big.NewInt(0)) != 0 {
			t.Errorf("Incorrect lot remaining RPL amount %s", lot1.RemainingRPLAmount.String())
		}
		if lot1.TotalBidAmount.Cmp(bidAmount) != 1 {
			t.Errorf("Incorrect lot total bid amount %s", lot1.TotalBidAmount.String())
		}
		if lot1.AddressBidAmount.Cmp(bidAmount) != 0 {
			t.Errorf("Incorrect lot address bid amount %s", lot1.AddressBidAmount.String())
		}
		if !lot1.Cleared {
			t.Error("Incorrect lot cleared status")
		}
		if lot1.RPLRecovered {
			t.Error("Incorrect lot RPL recovered status")
		}

		// Lot 1 prices at blocks
		if priceAtBlock, err := auction.GetLotPriceAtBlock(rp, lot1Index, 0, nil); err != nil {
			t.Error(err)
		} else if priceAtBlock.Cmp(lot1.StartPrice) != 0 {
			t.Errorf("Incorrect lot price at block 1 %s", priceAtBlock.String())
		}
		if priceAtBlock, err := auction.GetLotPriceAtBlock(rp, lot1Index, 1000000, nil); err != nil {
			t.Error(err)
		} else if priceAtBlock.Cmp(lot1.ReservePrice) != 0 {
			t.Errorf("Incorrect lot price at block 2 %s", priceAtBlock.String())
		}

		// Lot 2
		if lot2.Index != lot2Index {
			t.Errorf("Incorrect lot index %d", lot2.Index)
		}
		if !lot2.RPLRecovered {
			t.Error("Incorrect lot RPL recovered status")
		}

	}

	// Get & check initial bidder RPL balance
	if rplBalance, err := tokens.GetRPLBalance(rp, userAccount1.Address, nil); err != nil {
		t.Error(err)
	} else if rplBalance.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Incorrect initial bidder RPL balance %s", rplBalance.String())
	}

	// Claim bid on lot 1
	if _, err := auction.ClaimBid(rp, lot1Index, userAccount1.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check updated bidder RPL balance
	if rplBalance, err := tokens.GetRPLBalance(rp, userAccount1.Address, nil); err != nil {
		t.Error(err)
	} else if rplBalance.Cmp(big.NewInt(0)) != 1 {
		t.Errorf("Incorrect updated bidder RPL balance %s", rplBalance.String())
	}

}
