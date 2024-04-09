package auction

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
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
		option.Element = &claimableLots[i]
		option.ID = fmt.Sprint(lot.Index)
		option.Display = fmt.Sprintf("lot %d (%.6f ETH bid @ %.6f ETH per RPL)", lot.Index, math.RoundDown(eth.WeiToEth(lot.NodeBidAmount), 6), math.RoundDown(eth.WeiToEth(lot.CurrentPrice), 6))
	}
	selectedLots, err := utils.GetMultiselectIndices[api.AuctionLotDetails](c, claimLotsFlag, options, "Please select a lot to claim RPL from:")
	if err != nil {
		return fmt.Errorf("error determining lot selection: %w", err)
	}

	// Build the TXs
	indices := make([]uint64, len(selectedLots))
	for i, lot := range selectedLots {
		indices[i] = lot.Index
	}
	response, err := rp.Api.Auction.ClaimFromLots(indices)
	if err != nil {
		return fmt.Errorf("error during TX generation: %w", err)
	}

	// Validation
	txs := make([]*eth.TransactionInfo, len(selectedLots))
	for i, lot := range selectedLots {
		data := response.Data.Batch[i]
		if !data.CanClaim {
			fmt.Printf("Cannot claim lot %d:\n", lot.Index)
			if data.DoesNotExist {
				fmt.Println("The lot does not exist.")
			}
			if data.NoBidFromAddress {
				fmt.Println("The lot currently doesn't have a bid from your node.")
			}
			if data.NotCleared {
				fmt.Println("The lot is not cleared yet.")
			}
			return nil
		}
		txs[i] = data.TxInfo
	}

	// Claim RPL from lots
	validated, err := tx.HandleTxBatch(c, rp, txs,
		fmt.Sprintf("Are you sure you want to claim %d lots?", len(selectedLots)),
		func(i int) string {
			return fmt.Sprintf("claim of lot %d", selectedLots[i].Index)
		},
		"Claiming lots...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully claimed from all selected lots.")
	return nil
}
