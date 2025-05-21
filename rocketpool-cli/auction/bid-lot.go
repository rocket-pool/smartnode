package auction

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func bidOnLot(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get lot details
	lots, err := rp.AuctionLots()
	if err != nil {
		return err
	}

	// Get open lots
	openLots := []api.LotDetails{}
	for _, lot := range lots.Lots {
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
	var selectedLot api.LotDetails
	if c.String("lot") != "" {

		// Get selected lot index
		selectedIndex, err := strconv.ParseUint(c.String("lot"), 10, 64)
		if err != nil {
			return fmt.Errorf("Invalid lot ID '%s': %w", c.String("lot"), err)
		}

		// Get matching lot
		found := false
		for _, lot := range openLots {
			if lot.Details.Index == selectedIndex {
				selectedLot = lot
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Lot %d is not available for bidding.", selectedIndex)
		}

	} else {

		// Prompt for lot selection
		options := make([]string, len(openLots))
		for li, lot := range openLots {
			options[li] = fmt.Sprintf("lot %d (%.6f RPL available @ %.6f ETH per RPL)", lot.Details.Index, math.RoundDown(eth.WeiToEth(lot.Details.RemainingRPLAmount), 6), math.RoundDown(eth.WeiToEth(lot.Details.CurrentPrice), 6))
		}
		selected, _ := prompt.Select("Please select a lot to bid on:", options)
		selectedLot = openLots[selected]

	}

	// Get bid amount
	var amountWei *big.Int
	if c.String("amount") == "max" {

		// Set bid amount to maximum
		var tmp big.Int
		var maxAmount big.Int
		tmp.Mul(selectedLot.Details.RemainingRPLAmount, selectedLot.Details.CurrentPrice)
		maxAmount.Quo(&tmp, eth.EthToWei(1))
		amountWei = &maxAmount

	} else if c.String("amount") != "" {

		// Parse amount
		bidAmount, err := strconv.ParseFloat(c.String("amount"), 64)
		if err != nil {
			return fmt.Errorf("Invalid bid amount '%s': %w", c.String("amount"), err)
		}
		amountWei = eth.EthToWei(bidAmount)

	} else {

		// Calculate maximum bid amount
		var tmp big.Int
		var maxAmount big.Int
		tmp.Mul(selectedLot.Details.RemainingRPLAmount, selectedLot.Details.CurrentPrice)
		maxAmount.Quo(&tmp, eth.EthToWei(1))

		// Prompt for maximum amount
		if prompt.Confirm(fmt.Sprintf("Would you like to bid the maximum amount of ETH (%.6f ETH)?", math.RoundDown(eth.WeiToEth(&maxAmount), 6))) {
			amountWei = &maxAmount
		} else {

			// Prompt for custom amount
			inputAmount := prompt.Prompt("Please enter an amount of ETH to bid:", "^\\d+(\\.\\d+)?$", "Invalid amount")
			bidAmount, err := strconv.ParseFloat(inputAmount, 64)
			if err != nil {
				return fmt.Errorf("Invalid bid amount '%s': %w", inputAmount, err)
			}
			amountWei = eth.EthToWei(bidAmount)

		}

	}

	// Check lot can be bid on
	canBid, err := rp.CanBidOnLot(selectedLot.Details.Index, amountWei)
	if err != nil {
		return fmt.Errorf("Error checking if bidding on lot %d is possible: %w", selectedLot.Details.Index, err)
	}
	if !canBid.CanBid {
		fmt.Println("Cannot bid on lot:")
		if canBid.BidOnLotDisabled {
			fmt.Println("Bidding on lots is currently disabled.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canBid.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to bid %.6f ETH on lot %d? Bids are final and non-refundable.", math.RoundDown(eth.WeiToEth(amountWei), 6), selectedLot.Details.Index))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Bid on lot
	response, err := rp.BidOnLot(selectedLot.Details.Index, amountWei)
	if err != nil {
		return err
	}

	fmt.Printf("Bidding on lot...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully bid %.6f ETH on lot %d.\n", math.RoundDown(eth.WeiToEth(amountWei), 6), selectedLot.Details.Index)
	return nil

}
