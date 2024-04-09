package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

func refundMinipools(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get minipool statuses
	status, err := rp.Api.Minipool.Status()
	if err != nil {
		return err
	}

	// Get refundable minipools
	refundableMinipools := []api.MinipoolDetails{}
	for _, minipool := range status.Data.Minipools {
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
	options := make([]utils.SelectionOption[api.MinipoolDetails], len(refundableMinipools))
	for i, mp := range refundableMinipools {
		option := &options[i]
		option.Element = &refundableMinipools[i]
		option.ID = fmt.Sprint(mp.Address)
		option.Display = fmt.Sprintf("%s (%.6f ETH to claim)", mp.Address.Hex(), math.RoundDown(eth.WeiToEth(mp.Node.RefundBalance), 6))
	}
	selectedMinipools, err := utils.GetMultiselectIndices(c, minipoolsFlag, options, "Please select a minipool to refund ETH from:")
	if err != nil {
		return fmt.Errorf("error determining minipool selection: %w", err)
	}

	// Build the TXs
	addresses := make([]common.Address, len(selectedMinipools))
	for i, mp := range selectedMinipools {
		addresses[i] = mp.Address
	}
	response, err := rp.Api.Minipool.Refund(addresses)
	if err != nil {
		return fmt.Errorf("error during TX generation: %w", err)
	}

	// Validation
	txs := make([]*eth.TransactionInfo, len(selectedMinipools))
	for i, minipool := range selectedMinipools {
		txInfo := response.Data.TxInfos[i]
		if txInfo.SimulationResult.SimulationError != "" {
			return fmt.Errorf("error simulating refund for minipool %s: %s", minipool.Address.Hex(), txInfo.SimulationResult.SimulationError)
		}
		txs[i] = txInfo
	}

	// Run the TXs
	validated, err := tx.HandleTxBatch(c, rp, txs,
		fmt.Sprintf("Are you sure you want to refund %d minipools?", len(selectedMinipools)),
		func(i int) string {
			return fmt.Sprintf("refund on minipool %s", selectedMinipools[i].Address.Hex())
		},
		"Refunding minipools...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully refunded ETH from all selected minipools.")
	return nil
}
