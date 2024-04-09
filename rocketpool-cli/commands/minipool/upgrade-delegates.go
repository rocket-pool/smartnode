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

func upgradeDelegates(c *cli.Context) error {
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

	// Get upgradeable minipools
	upgradeableMinipools := []api.MinipoolDetails{}
	for _, mp := range status.Data.Minipools {
		if mp.Delegate != status.Data.LatestDelegate && !mp.UseLatestDelegate {
			upgradeableMinipools = append(upgradeableMinipools, mp)
		}
	}

	// Check for upgradeable minipools
	if len(upgradeableMinipools) == 0 {
		fmt.Println("No minipools are eligible for delegate upgrades.")
		return nil
	}

	// Get selected minipools
	options := make([]utils.SelectionOption[api.MinipoolDetails], len(upgradeableMinipools))
	for i, mp := range upgradeableMinipools {
		option := &options[i]
		option.Element = &upgradeableMinipools[i]
		option.ID = fmt.Sprint(mp.Address)
		option.Display = fmt.Sprintf("%s (using delegate %s)", mp.Address.Hex(), mp.Delegate.Hex())
	}
	selectedMinipools, err := utils.GetMultiselectIndices(c, minipoolsFlag, options, "Please select a minipool to upgrade:")
	if err != nil {
		return fmt.Errorf("error determining minipool selection: %w", err)
	}

	// Build the TXs
	addresses := make([]common.Address, len(selectedMinipools))
	for i, mp := range selectedMinipools {
		addresses[i] = mp.Address
	}
	response, err := rp.Api.Minipool.UpgradeDelegates(addresses)
	if err != nil {
		return fmt.Errorf("error during TX generation: %w", err)
	}

	// Validation
	txs := make([]*eth.TransactionInfo, len(selectedMinipools))
	for i := range selectedMinipools {
		txInfo := response.Data.TxInfos[i]
		txs[i] = txInfo
	}

	fmt.Printf("Minipools will upgrade to delegate contract %s.\n", status.Data.LatestDelegate.Hex())

	// Run the TXs
	validated, err := tx.HandleTxBatch(c, rp, txs,
		fmt.Sprintf("Are you sure you want to upgrade %d minipools?", len(selectedMinipools)),
		func(i int) string {
			return fmt.Sprintf("upgrade of minipool %s", selectedMinipools[i].Address.Hex())
		},
		"Upgrading minipools...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully upgraded all selected minipools.")
	return nil
}
