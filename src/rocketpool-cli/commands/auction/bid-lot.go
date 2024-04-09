package auction

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

const (
	bidLotFlag    string = "lot"
	bidAmountFlag string = "amount"
)

func bidOnLot(c *cli.Context) error {
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

	// Get open lots
	openLots := []api.AuctionLotDetails{}
	for _, lot := range lots.Data.Lots {
		if lot.BiddingAvailable {
			openLots = append(openLots, lot)
		}
	}

	// Check for open lots
	if len(openLots) == 0 {
		fmt.Println("No lots can be bid on.")
		return nil
	}

	// Get selected lot
	var selectedLot api.AuctionLotDetails
	if c.String(bidLotFlag) != "" {
		// Get selected lot index
		selectedIndex, err := strconv.ParseUint(c.String(bidLotFlag), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid lot ID '%s': %w", c.String(bidLotFlag), err)
		}

		// Get matching lot
		found := false
		for _, lot := range openLots {
			if lot.Index == selectedIndex {
				selectedLot = lot
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("lot %d is not available for bidding", selectedIndex)
		}
	} else {
		// Prompt for lot selection
		options := make([]string, len(openLots))
		for li, lot := range openLots {
			options[li] = fmt.Sprintf("lot %d (%.6f RPL available @ %.6f ETH per RPL)", lot.Index, math.RoundDown(eth.WeiToEth(lot.RemainingRplAmount), 6), math.RoundDown(eth.WeiToEth(lot.CurrentPrice), 6))
		}
		selected, _ := utils.Select("Please select a lot to bid on:", options)
		selectedLot = openLots[selected]
	}

	// Get bid amount
	var amountWei *big.Int
	if c.String(bidAmountFlag) == "max" {
		// Set bid amount to maximum
		var tmp big.Int
		var maxAmount big.Int
		tmp.Mul(selectedLot.RemainingRplAmount, selectedLot.CurrentPrice)
		maxAmount.Quo(&tmp, eth.EthToWei(1))
		amountWei = &maxAmount
	} else if c.String(bidAmountFlag) != "" {
		// Parse amount
		bidAmount, err := strconv.ParseFloat(c.String(bidAmountFlag), 64)
		if err != nil {
			return fmt.Errorf("invalid bid amount '%s': %w", c.String(bidAmountFlag), err)
		}
		amountWei = eth.EthToWei(bidAmount)
	} else {
		// Calculate maximum bid amount
		var tmp big.Int
		var maxAmount big.Int
		tmp.Mul(selectedLot.RemainingRplAmount, selectedLot.CurrentPrice)
		maxAmount.Quo(&tmp, eth.EthToWei(1))

		// Prompt for maximum amount
		if utils.Confirm(fmt.Sprintf("Would you like to bid the maximum amount of ETH (%.6f ETH)?", math.RoundDown(eth.WeiToEth(&maxAmount), 6))) {
			amountWei = &maxAmount
		} else {

			// Prompt for custom amount
			inputAmount := utils.Prompt("Please enter an amount of ETH to bid:", "^\\d+(\\.\\d+)?$", "Invalid amount")
			bidAmount, err := strconv.ParseFloat(inputAmount, 64)
			if err != nil {
				return fmt.Errorf("invalid bid amount '%s': %w", inputAmount, err)
			}
			amountWei = eth.EthToWei(bidAmount)
		}
	}

	// Build the TX
	response, err := rp.Api.Auction.BidOnLot(selectedLot.Index, amountWei)
	if err != nil {
		return fmt.Errorf("error checking if bidding on lot %d is possible: %w", selectedLot.Index, err)
	}

	// Verify
	if !response.Data.CanBid {
		fmt.Println("Cannot bid on lot:")
		if response.Data.BidOnLotDisabled {
			fmt.Println("Bidding on lots is currently disabled.")
		}
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf("Are you sure you want to bid %.6f ETH on lot %d? Bids are final and non-refundable.", math.RoundDown(eth.WeiToEth(amountWei), 6), selectedLot.Index),
		fmt.Sprintf("bid on lot %d", selectedLot.Index),
		fmt.Sprintf("Bidding on lot %d...", selectedLot.Index),
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("Successfully bid %.6f ETH on lot %d.\n", math.RoundDown(eth.WeiToEth(amountWei), 6), selectedLot.Index)
	return nil
}
