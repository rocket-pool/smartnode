package auction

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
)

func getStatus(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get auction status
	status, err := rp.Api.Auction.Status()
	if err != nil {
		return err
	}

	// Print & return
	fmt.Printf(
		"A total of %.6f RPL is up for auction, with %.6f RPL currently allotted and %.6f RPL remaining.\n",
		math.RoundDown(eth.WeiToEth(status.Data.TotalRplBalance), 6),
		math.RoundDown(eth.WeiToEth(status.Data.AllottedRplBalance), 6),
		math.RoundDown(eth.WeiToEth(status.Data.RemainingRplBalance), 6))
	if status.Data.LotCounts.ClaimAvailable > 0 {
		fmt.Printf("%d lot(s) you have bid on have RPL available to claim!\n", status.Data.LotCounts.ClaimAvailable)
	}
	if status.Data.LotCounts.BiddingAvailable > 0 {
		fmt.Printf("%d lot(s) are open for bidding!\n", status.Data.LotCounts.BiddingAvailable)
	}
	if status.Data.LotCounts.RplRecoveryAvailable > 0 {
		fmt.Printf("%d cleared lot(s) have unclaimed RPL ready to recover!\n", status.Data.LotCounts.RplRecoveryAvailable)
	}
	if status.Data.CanCreateLot {
		fmt.Println("A new lot can be created with remaining RPL in the auction contract.")
	}
	return nil
}
