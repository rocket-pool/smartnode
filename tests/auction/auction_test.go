package auction

import (
    "math/big"
    "testing"

    "github.com/rocket-pool/rocketpool-go/auction"
    "github.com/rocket-pool/rocketpool-go/utils/eth"

    auctionutils "github.com/rocket-pool/rocketpool-go/tests/testutils/auction"
    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
    nodeutils "github.com/rocket-pool/rocketpool-go/tests/testutils/node"
)


func TestAuctionDetails(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register node
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Get & check initial RPL balances
    totalBalance1, err := auction.GetTotalRPLBalance(rp, nil)
    if err != nil { t.Fatal(err) }
    allottedBalance1, err := auction.GetAllottedRPLBalance(rp, nil)
    if err != nil { t.Fatal(err) }
    remainingBalance1, err := auction.GetRemainingRPLBalance(rp, nil)
    if err != nil { t.Fatal(err) }
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
    if err := auctionutils.CreateSlashedRPL(rp, ownerAccount, trustedNodeAccount, userAccount1); err != nil {
        t.Fatal(err)
    }

    // Get & check updated RPL balances
    totalBalance2, err := auction.GetTotalRPLBalance(rp, nil)
    if err != nil { t.Fatal(err) }
    allottedBalance2, err := auction.GetAllottedRPLBalance(rp, nil)
    if err != nil { t.Fatal(err) }
    remainingBalance2, err := auction.GetRemainingRPLBalance(rp, nil)
    if err != nil { t.Fatal(err) }
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
    if err != nil { t.Fatal(err) }
    allottedBalance3, err := auction.GetAllottedRPLBalance(rp, nil)
    if err != nil { t.Fatal(err) }
    remainingBalance3, err := auction.GetRemainingRPLBalance(rp, nil)
    if err != nil { t.Fatal(err) }
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
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register node
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Mint slashed RPL to auction contract
    if err := auctionutils.CreateSlashedRPL(rp, ownerAccount, trustedNodeAccount, userAccount1); err != nil { t.Fatal(err) }

    // Get initial lot details
    if lots, err := auction.GetLotsWithBids(rp, userAccount1.Address, nil); err != nil {
        t.Error(err)
    } else if len(lots) != 0 {
        t.Error("Incorrect initial lot count")
    }

    // Create lot
    lotIndex, _, err := auction.CreateLot(rp, userAccount1.GetTransactor())
    if err != nil { t.Fatal(err) }

    // Place bid on lot
    bidAmount := eth.EthToWei(1)
    bid1Opts := userAccount1.GetTransactor()
    bid1Opts.Value = bidAmount
    if _, err := auction.PlaceBid(rp, lotIndex, bid1Opts); err != nil { t.Fatal(err) }

    // Place another lot on bid to close it
    bid2Opts := userAccount2.GetTransactor()
    bid2Opts.Value = eth.EthToWei(1000)
    if _, err := auction.PlaceBid(rp, lotIndex, bid2Opts); err != nil { t.Fatal(err) }

    // Get updated lot details
    if lots, err := auction.GetLotsWithBids(rp, userAccount1.Address, nil); err != nil {
        t.Error(err)
    } else if len(lots) != 1 {
        t.Error("Incorrect updated lot count")
    } else {
        lot := lots[0]
        if lot.Index != lotIndex {
            t.Errorf("Incorrect lot index %d", lot.Index)
        }
        if !lot.Exists {
            t.Error("Incorrect lot exists status")
        }
        if lot.StartBlock == 0 {
            t.Errorf("Incorrect lot start block %d", lot.StartBlock)
        }
        if lot.EndBlock <= lot.StartBlock {
            t.Errorf("Incorrect lot end block %d", lot.EndBlock)
        }
        //if lot.StartPrice
        //if lot.ReservePrice
        if lot.PriceAtCurrentBlock.Cmp(lot.StartPrice) == 1 || lot.PriceAtCurrentBlock.Cmp(lot.ReservePrice) == -1 {
            t.Errorf("Incorrect lot price at current block %s", lot.PriceAtCurrentBlock.String())
        }
        if lot.PriceByTotalBids.Cmp(lot.StartPrice) == 1 || lot.PriceByTotalBids.Cmp(lot.ReservePrice) == -1 {
            t.Errorf("Incorrect lot price at current block %s", lot.PriceByTotalBids.String())
        }
        if lot.CurrentPrice.Cmp(lot.StartPrice) == 1 || lot.CurrentPrice.Cmp(lot.ReservePrice) == -1 {
            t.Errorf("Incorrect lot price at current block %s", lot.CurrentPrice.String())
        }
        //if lot.TotalRPLAmount
        //if lot.ClaimedRPLAmount
        //if lot.RemainingRPLAmount
        if lot.TotalBidAmount.Cmp(bidAmount) != 1 {
            t.Errorf("Incorrect lot total bid amount %s", lot.TotalBidAmount.String())
        }
        if lot.AddressBidAmount.Cmp(bidAmount) != 0 {
            t.Errorf("Incorrect lot address bid amount %s", lot.AddressBidAmount.String())
        }
        if !lot.Cleared {
            t.Error("Incorrect lot cleared status")
        }
        if lot.RPLRecovered {
            t.Error("Incorrect lot RPL recovered status")
        }
    }

}

