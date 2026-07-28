package minipool

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"

	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func refundMinipools(minipool string, yes bool) error {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get minipool statuses
	status, err := rp.MinipoolStatus()
	if err != nil {
		return err
	}

	// Get refundable minipools
	refundableMinipools := []api.MinipoolDetails{}
	for _, minipool := range status.Minipools {
		if minipool.RefundAvailable {
			refundableMinipools = append(refundableMinipools, minipool)
		}
	}

	// Check for refundable minipools
	if len(refundableMinipools) == 0 {
		fmt.Println("No minipools have refunds available.")
		return nil
	}

	// Get selected minipools
	var selectedMinipools []api.MinipoolDetails
	if minipool == "" {

		// Prompt for minipool selection
		options := make([]string, len(refundableMinipools)+1)
		options[0] = "All available minipools"
		for mi, minipool := range refundableMinipools {
			options[mi+1] = fmt.Sprintf("%s (%.6f ETH to claim)", minipool.Address.Hex(), math.RoundDown(math.WeiToEth(minipool.Node.RefundBalance), 6))
		}
		selected, _ := prompt.Select("Please select a minipool to refund ETH from:", options)

		// Get minipools
		if selected == 0 {
			selectedMinipools = refundableMinipools
		} else {
			selectedMinipools = []api.MinipoolDetails{refundableMinipools[selected-1]}
		}

	} else {

		// Get matching minipools
		if minipool == "all" {
			selectedMinipools = refundableMinipools
		} else {
			selectedAddress := common.HexToAddress(minipool)
			for _, minipool := range refundableMinipools {
				if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
					selectedMinipools = []api.MinipoolDetails{minipool}
					break
				}
			}
			if selectedMinipools == nil {
				return fmt.Errorf("The minipool %s is not available for refund.", selectedAddress.Hex())
			}
		}

	}

	// Get the total gas limit estimate
	var gasLimits gaslimit.Limits
	for _, minipool := range selectedMinipools {
		canResponse, err := rp.CanRefundMinipool(minipool.Address)
		if err != nil {
			fmt.Printf("WARNING: Couldn't get gas price for refund transaction (%s)", err.Error())
			break
		}
		gasLimits = gasLimits.Add(canResponse.GasLimits)
	}

	// Get max fees
	g, err := gas.GetMaxFeeAndLimit(gasLimits, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if prompt.Declined(yes, "Are you sure you want to refund %d minipools?", len(selectedMinipools)) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Refund minipools
	for _, minipool := range selectedMinipools {
		g.Assign(rp)
		response, err := rp.RefundMinipool(minipool.Address)
		if err != nil {
			fmt.Printf("Could not refund ETH from minipool %s: %s.\n", minipool.Address.Hex(), err.Error())
			continue
		}

		fmt.Printf("Refunding minipool %s...\n", minipool.Address.Hex())
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			fmt.Printf("Could not refund ETH from minipool %s: %s.\n", minipool.Address.Hex(), err.Error())
		} else {
			fmt.Printf("Successfully refunded ETH from minipool %s.\n", minipool.Address.Hex())
		}
	}

	// Return
	return nil

}
