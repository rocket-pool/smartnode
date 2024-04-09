package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

func promoteMinipools(c *cli.Context) error {
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

	// Get promotable minipools
	promotableMinipools := []api.MinipoolDetails{}
	for _, minipool := range status.Data.Minipools {
		if minipool.CanPromote {
			promotableMinipools = append(promotableMinipools, minipool)
		}
	}

	// Check for promotable minipools
	if len(promotableMinipools) == 0 {
		fmt.Println("No minipools can be promoted.")
		return nil
	}

	// Get selected minipools
	options := make([]utils.SelectionOption[api.MinipoolDetails], len(promotableMinipools))
	for i, mp := range promotableMinipools {
		option := &options[i]
		option.Element = &promotableMinipools[i]
		option.ID = fmt.Sprint(mp.Address)
		option.Display = fmt.Sprintf("%s (%s until dissolved)", mp.Address.Hex(), mp.TimeUntilDissolve)
	}
	selectedMinipools, err := utils.GetMultiselectIndices(c, minipoolsFlag, options, "Please select a minipool to promote:")
	if err != nil {
		return fmt.Errorf("error determining minipool selection: %w", err)
	}

	// Build the TXs
	addresses := make([]common.Address, len(selectedMinipools))
	for i, mp := range selectedMinipools {
		addresses[i] = mp.Address
	}
	response, err := rp.Api.Minipool.Promote(addresses)
	if err != nil {
		return fmt.Errorf("error during TX generation: %w", err)
	}

	// Validation
	txs := make([]*eth.TransactionInfo, len(selectedMinipools))
	for i := range selectedMinipools {
		txInfo := response.Data.TxInfos[i]
		txs[i] = txInfo
	}

	// Run the TXs
	validated, err := tx.HandleTxBatch(c, rp, txs,
		fmt.Sprintf("Are you sure you want to promote %d minipools?", len(selectedMinipools)),
		func(i int) string {
			return fmt.Sprintf("promoting minipool %s", selectedMinipools[i].Address.Hex())
		},
		"Promoting minipools...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully promoted all selected minipools.")
	return nil
}
