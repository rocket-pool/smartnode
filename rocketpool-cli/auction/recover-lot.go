package auction

import (
	"fmt"
	"strconv"

	rocketpoolapi "github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func recoverRplFromLot(c *cli.Context) error {

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

	// Get recoverable lots
	recoverableLots := []api.LotDetails{}
	for _, lot := range lots.Lots {
		if lot.RPLRecoveryAvailable {
			recoverableLots = append(recoverableLots, lot)
		}
	}

	// Check for recoverable lots
	if len(recoverableLots) == 0 {
		fmt.Println("No lots are available for RPL recovery.")
		return nil
	}

	// Get selected lots
	var selectedLots []api.LotDetails
	if c.String("lot") == "all" {

		// Select all recoverable lots
		selectedLots = recoverableLots

	} else if c.String("lot") != "" {

		// Get selected lot index
		selectedIndex, err := strconv.ParseUint(c.String("lot"), 10, 64)
		if err != nil {
			return fmt.Errorf("Invalid lot ID '%s': %w", c.String("lot"), err)
		}

		// Get matching lot
		found := false
		for _, lot := range recoverableLots {
			if lot.Details.Index == selectedIndex {
				selectedLots = []api.LotDetails{lot}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Lot %d is not available for RPL recovery.", selectedIndex)
		}

	} else {

		// Prompt for lot selection
		options := make([]string, len(recoverableLots)+1)
		options[0] = "All available lots"
		for li, lot := range recoverableLots {
			options[li+1] = fmt.Sprintf("lot %d (%.6f RPL unclaimed)", lot.Details.Index, math.RoundDown(eth.WeiToEth(lot.Details.RemainingRPLAmount), 6))
		}
		selected, _ := cliutils.Select("Please select a lot to recover unclaimed RPL from:", options)

		// Get lots
		if selected == 0 {
			selectedLots = recoverableLots
		} else {
			selectedLots = []api.LotDetails{recoverableLots[selected-1]}
		}

	}

	// Get the total gas limit estimate
	var totalGas uint64 = 0
	var totalSafeGas uint64 = 0
	var gasInfo rocketpoolapi.GasInfo
	for _, lot := range selectedLots {
		canResponse, err := rp.CanRecoverUnclaimedRPLFromLot(lot.Details.Index)
		if err != nil {
			return fmt.Errorf("Error checking if recovering lot %d is possible: %w", lot.Details.Index, err)
		}
		gasInfo = canResponse.GasInfo
		totalGas += canResponse.GasInfo.EstGasLimit
		totalSafeGas += canResponse.GasInfo.SafeGasLimit
	}
	gasInfo.EstGasLimit = totalGas
	gasInfo.SafeGasLimit = totalSafeGas

	// Get max fees
	g, err := gas.GetMaxFeeAndLimit(gasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to recover %d lots?", len(selectedLots)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Claim RPL from lots
	for _, lot := range selectedLots {
		g.Assign(rp)
		response, err := rp.RecoverUnclaimedRPLFromLot(lot.Details.Index)
		if err != nil {
			fmt.Printf("Could not recover unclaimed RPL from lot %d: %s.\n", lot.Details.Index, err)
			continue
		}

		fmt.Printf("Recovering lot %d...\n", lot.Details.Index)
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			fmt.Printf("Could not recover unclaimed RPL from lot %d: %s.\n", lot.Details.Index, err)
		} else {
			fmt.Printf("Successfully recovered unclaimed RPL from lot %d.\n", lot.Details.Index)
		}
	}

	// Return
	return nil

}
