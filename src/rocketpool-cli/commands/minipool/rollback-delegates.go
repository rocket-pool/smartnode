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

func rollbackDelegates(c *cli.Context) error {
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

	// Get rollback-capable minipools
	eligibleMinipools := []api.MinipoolDetails{}
	for _, mp := range status.Data.Minipools {
		if mp.Version < 3 {
			// Rollback is disabled for minipools introduced with Atlas (e.g. LEB8s or downconversions)
			eligibleMinipools = append(eligibleMinipools, mp)
		}
	}

	// Check for rollback-capable minipools
	if len(eligibleMinipools) == 0 {
		fmt.Println("No minipools are eligible for delegate rollbacks.")
		return nil
	}

	// Get selected minipools
	options := make([]utils.SelectionOption[api.MinipoolDetails], len(eligibleMinipools))
	for i, mp := range eligibleMinipools {
		option := &options[i]
		option.Element = &eligibleMinipools[i]
		option.ID = fmt.Sprint(mp.Address)
		option.Display = fmt.Sprintf("%s (using delegate %s, will roll back to %s)", mp.Address.Hex(), mp.Delegate.Hex(), mp.PreviousDelegate.Hex())
	}
	selectedMinipools, err := utils.GetMultiselectIndices(c, minipoolsFlag, options, "Please select a minipool to rollback the delegate for:")
	if err != nil {
		return fmt.Errorf("error determining minipool selection: %w", err)
	}

	// Build the TXs
	addresses := make([]common.Address, len(selectedMinipools))
	for i, mp := range selectedMinipools {
		addresses[i] = mp.Address
	}
	response, err := rp.Api.Minipool.RollbackDelegates(addresses)
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
		fmt.Sprintf("Are you sure you want to rollback %d minipools?", len(selectedMinipools)),
		func(i int) string {
			return fmt.Sprintf("rollback of minipool %s", selectedMinipools[i].Address.Hex())
		},
		"Rolling back minipool delegates...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully rolled back all selected minipools.")
	return nil
}
