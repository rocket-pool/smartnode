package auction

import (
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func getStatus(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get auction status
	status, err := rp.AuctionStatus()
	if err != nil {
		return err
	}

	// Print & return
	fmt.Printf(
		"A total of %.6f RPL is up for auction, with %.6f RPL currently allotted and %.6f RPL remaining.\n",
		math.RoundDown(eth.WeiToEth(status.TotalRPLBalance), 6),
		math.RoundDown(eth.WeiToEth(status.AllottedRPLBalance), 6),
		math.RoundDown(eth.WeiToEth(status.RemainingRPLBalance), 6))
	if status.LotCounts.ClaimAvailable > 0 {
		fmt.Printf("%d lot(s) you have bid on have RPL available to claim!\n", status.LotCounts.ClaimAvailable)
	}
	if status.LotCounts.BiddingAvailable > 0 {
		fmt.Printf("%d lot(s) are open for bidding!\n", status.LotCounts.BiddingAvailable)
	}
	if status.LotCounts.RPLRecoveryAvailable > 0 {
		fmt.Printf("%d cleared lot(s) have unclaimed RPL ready to recover!\n", status.LotCounts.RPLRecoveryAvailable)
	}
	if status.CanCreateLot {
		fmt.Println("A new lot can be created with remaining RPL in the auction contract.")
	}
	return nil

}
