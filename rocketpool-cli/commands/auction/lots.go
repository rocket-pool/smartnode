package auction

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

func getLots(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get lot details
	lots, err := rp.Api.Auction.Lots()
	if err != nil {
		return err
	}

	// Get lots by status
	openLots := []api.AuctionLotDetails{}
	clearedLots := []api.AuctionLotDetails{}
	claimableLots := []api.AuctionLotDetails{}
	biddableLots := []api.AuctionLotDetails{}
	recoverableLots := []api.AuctionLotDetails{}
	for _, lot := range lots.Data.Lots {
		if lot.IsCleared {
			clearedLots = append(clearedLots, lot)
		} else {
			openLots = append(openLots, lot)
		}
		if lot.ClaimAvailable {
			claimableLots = append(claimableLots, lot)
		}
		if lot.BiddingAvailable {
			biddableLots = append(biddableLots, lot)
		}
		if lot.RplRecoveryAvailable {
			recoverableLots = append(recoverableLots, lot)
		}
	}

	// Print lot details by status
	if len(lots.Data.Lots) == 0 {
		fmt.Println("There are no lots for auction yet.")
	}
	for status := 0; status < 2; status++ {

		// Get status title format & lot list
		var statusFormat string
		var statusLots []api.AuctionLotDetails
		if status == 0 {
			statusFormat = "%d lot(s) open for bidding:\n"
			statusLots = openLots
		} else {
			statusFormat = "%d cleared lot(s):\n"
			statusLots = clearedLots
		}
		if len(statusLots) == 0 {
			continue
		}

		// Print
		fmt.Printf(statusFormat, len(statusLots))
		for _, lot := range statusLots {
			fmt.Printf("--------------------\n")
			fmt.Printf("\n")
			fmt.Printf("Lot ID:               %d\n", lot.Index)
			fmt.Printf("Start block:          %d\n", lot.StartBlock)
			fmt.Printf("End block:            %d\n", lot.EndBlock)
			fmt.Printf("RPL starting price:   %.6f\n", math.RoundDown(eth.WeiToEth(lot.StartPrice), 6))
			fmt.Printf("RPL reserve price:    %.6f\n", math.RoundDown(eth.WeiToEth(lot.ReservePrice), 6))
			fmt.Printf("RPL current price:    %.6f\n", math.RoundDown(eth.WeiToEth(lot.CurrentPrice), 6))
			fmt.Printf("Total RPL amount:     %.6f\n", math.RoundDown(eth.WeiToEth(lot.TotalRplAmount), 6))
			fmt.Printf("Claimed RPL amount:   %.6f\n", math.RoundDown(eth.WeiToEth(lot.ClaimedRplAmount), 6))
			fmt.Printf("Remaining RPL amount: %.6f\n", math.RoundDown(eth.WeiToEth(lot.RemainingRplAmount), 6))
			fmt.Printf("Total ETH bid:        %.6f\n", math.RoundDown(eth.WeiToEth(lot.TotalBidAmount), 6))
			fmt.Printf("ETH bid by node:      %.6f\n", math.RoundDown(eth.WeiToEth(lot.NodeBidAmount), 6))
			if lot.IsCleared {
				fmt.Printf("Cleared:              yes\n")
				if lot.RemainingRplAmount.Cmp(big.NewInt(0)) == 0 {
					fmt.Printf("Unclaimed RPL:        no\n")
				} else if lot.RplRecovered {
					fmt.Printf("Unclaimed RPL:        recovered\n")
				} else {
					fmt.Printf("Unclaimed RPL:        yes\n")
				}
			} else {
				fmt.Printf("Cleared:              no\n")
			}
			fmt.Printf("\n")
		}
		fmt.Println("")

	}

	// Print actionable lot details
	if len(claimableLots) > 0 {
		fmt.Printf("%d lot(s) you have bid on have RPL available to claim:\n", len(claimableLots))
		for _, lot := range claimableLots {
			fmt.Printf("- lot %d (%.6f ETH bid @ %.6f ETH per RPL)\n", lot.Index, math.RoundDown(eth.WeiToEth(lot.NodeBidAmount), 6), math.RoundDown(eth.WeiToEth(lot.CurrentPrice), 6))
		}
		fmt.Println("")
	}
	if len(biddableLots) > 0 {
		fmt.Printf("%d lot(s) are open for bidding:\n", len(biddableLots))
		for _, lot := range biddableLots {
			fmt.Printf("- lot %d (%.6f RPL available @ %.6f ETH per RPL)\n", lot.Index, math.RoundDown(eth.WeiToEth(lot.RemainingRplAmount), 6), math.RoundDown(eth.WeiToEth(lot.CurrentPrice), 6))
		}
		fmt.Println("")
	}
	if len(recoverableLots) > 0 {
		fmt.Printf("%d lot(s) have unclaimed RPL ready to recover:\n", len(recoverableLots))
		for _, lot := range recoverableLots {
			fmt.Printf("- lot %d (%.6f RPL unclaimed)\n", lot.Index, math.RoundDown(eth.WeiToEth(lot.RemainingRplAmount), 6))
		}
		fmt.Println("")
	}

	// Return
	return nil
}
