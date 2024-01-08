package auction

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

const (
	claimLotsFlag string = "lots"
)

func claimFromLot(c *cli.Context) error {
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

	// Get claimable lots
	claimableLots := []api.AuctionLotDetails{}
	for _, lot := range lots.Data.Lots {
		if lot.ClaimAvailable {
			claimableLots = append(claimableLots, lot)
		}
	}

	// Check for claimable lots
	if len(claimableLots) == 0 {
		fmt.Println("No lots are available for RPL claims.")
		return nil
	}

	// Get selected lots
	options := make([]utils.SelectionOption[api.AuctionLotDetails], len(claimableLots))
	for i, lot := range claimableLots {
		option := &options[i]
		option.Element = &lot
		option.ID = fmt.Sprint(lot.Index)
		option.Display = fmt.Sprintf("lot %d (%.6f ETH bid @ %.6f ETH per RPL)", lot.Index, math.RoundDown(eth.WeiToEth(lot.NodeBidAmount), 6), math.RoundDown(eth.WeiToEth(lot.CurrentPrice), 6))
	}
	selectedLots, err := utils.GetMultiselectIndices[api.AuctionLotDetails](c, claimLotsFlag, options, "Please select a lot to claim RPL from:")
	if err != nil {
		return fmt.Errorf("error determining lot selection: %w", err)
	}

	// Validation
	txs := make([]*core.TransactionInfo, len(selectedLots))
	for i, lot := range selectedLots {
		response, err := rp.Api.Auction.ClaimFromLot(lot.Index)
		if err != nil {
			return fmt.Errorf("error checking if claiming lot %d is possible: %w", lot.Index, err)
		}
		if !response.Data.CanClaim {
			fmt.Printf("Cannot claim lot %d:\n", lot.Index)
			if response.Data.DoesNotExist {
				fmt.Println("The lot does not exist.")
			}
			if response.Data.NoBidFromAddress {
				fmt.Println("The lot currently doesn't have a bid from your node.")
			}
			if response.Data.NotCleared {
				fmt.Println("The lot is not cleared yet.")
			}
			return nil
		}
		if response.Data.TxInfo.SimError != "" {
			return fmt.Errorf("error simulating claim of lot %d: %s", lot.Index, response.Data.TxInfo.SimError)
		}
		txs[i] = response.Data.TxInfo
	}

	// Claim RPL from lots
	err = tx.HandleTxBatch(c, rp, txs,
		fmt.Sprintf("Are you sure you want to claim %d lots?", len(selectedLots)),
		"Claiming lots...",
	)
	if err != nil {
		return err
	}

	// Log & return
	fmt.Println("Successfully claimed from all selected lots.")
	return nil

}
