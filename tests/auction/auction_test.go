package auction

import (
    "math/big"
    "testing"

    "github.com/rocket-pool/rocketpool-go/auction"

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
    if err := auctionutils.CreateSlashedRPL(rp, ownerAccount, trustedNodeAccount, userAccount); err != nil {
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
    if _, err := auction.CreateLot(rp, userAccount.GetTransactor()); err != nil {
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

