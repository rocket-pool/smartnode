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
	recoverLotsFlag string = "lots"
)

func recoverRplFromLot(c *cli.Context) error {
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

	// Get recoverable lots
	recoverableLots := []api.AuctionLotDetails{}
	for _, lot := range lots.Data.Lots {
		if lot.RplRecoveryAvailable {
			recoverableLots = append(recoverableLots, lot)
		}
	}

	// Check for recoverable lots
	if len(recoverableLots) == 0 {
		fmt.Println("No lots are available for RPL recovery.")
		return nil
	}

	// Get selected lots
	options := make([]utils.SelectionOption[api.AuctionLotDetails], len(recoverableLots))
	for i, lot := range recoverableLots {
		option := &options[i]
		option.Element = &recoverableLots[i]
		option.ID = fmt.Sprint(lot.Index)
		option.Display = fmt.Sprintf("lot %d (%.6f RPL unclaimed)", lot.Index, math.RoundDown(eth.WeiToEth(lot.RemainingRplAmount), 6))
	}
	selectedLots, err := utils.GetMultiselectIndices[api.AuctionLotDetails](c, recoverLotsFlag, options, "Please select a lot to recover unclaimed from:")
	if err != nil {
		return fmt.Errorf("error determining lot selection: %w", err)
	}

	// Build the TXs
	indices := make([]uint64, len(selectedLots))
	for i, lot := range selectedLots {
		indices[i] = lot.Index
	}
	response, err := rp.Api.Auction.RecoverUnclaimedRplFromLots(indices)
	if err != nil {
		return fmt.Errorf("error during TX generation: %w", err)
	}

	// Validation
	txs := make([]*eth.TransactionInfo, len(selectedLots))
	for i, lot := range selectedLots {
		data := response.Data.Batch[i]
		if !data.CanRecover {
			fmt.Printf("Cannot recover lot %d:\n", lot.Index)
			if data.DoesNotExist {
				fmt.Println("The lot does not exist.")
			}
			if data.BiddingNotEnded {
				fmt.Println("Bidding on the lot has not ended yet.")
			}
			if data.NoUnclaimedRpl {
				fmt.Println("The lot does not have any unclaimed RPL.")
			}
			if data.RplAlreadyRecovered {
				fmt.Println("The lot's RPL has already been recovered.")
			}
			return nil
		}
		txs[i] = data.TxInfo
	}

	// Claim RPL from lots
	validated, err := tx.HandleTxBatch(c, rp, txs,
		fmt.Sprintf("Are you sure you want to recover %d lots?", len(selectedLots)),
		func(i int) string {
			return fmt.Sprintf("recovery of lot %d", selectedLots[i].Index)
		},
		"Recovering lots...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully recovered unclaimed RPL from all selected lots.")
	return nil
}
