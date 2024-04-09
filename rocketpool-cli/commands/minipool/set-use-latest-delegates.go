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

func setUseLatestDelegates(c *cli.Context, setting bool) error {
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

	// Get eligible settableMinipools
	settableMinipools := []api.MinipoolDetails{}
	for _, mp := range status.Data.Minipools {
		if mp.UseLatestDelegate != setting && !mp.Finalised {
			settableMinipools = append(settableMinipools, mp)
		}
	}

	// Check for initialized minipools
	if len(settableMinipools) == 0 {
		fmt.Printf("No minipools can have their use-latest-delegate flag set to %t.\n", setting)
		return nil
	}

	// Get selected minipools
	options := make([]utils.SelectionOption[api.MinipoolDetails], len(settableMinipools))
	for i, mp := range settableMinipools {
		option := &options[i]
		option.Element = &settableMinipools[i]
		option.ID = fmt.Sprint(mp.Address)
		option.Display = fmt.Sprintf("%s (using delegate %s)", mp.Address.Hex(), mp.Delegate.Hex())
	}
	var action string
	if setting {
		action = "enabled"
	} else {
		action = "disable"
	}
	selectedMinipools, err := utils.GetMultiselectIndices(c, minipoolsFlag, options, fmt.Sprintf("Please select a minipool to %s the flag for:", action))
	if err != nil {
		return fmt.Errorf("error determining minipool selection: %w", err)
	}

	// Build the TXs
	addresses := make([]common.Address, len(selectedMinipools))
	for i, mp := range selectedMinipools {
		addresses[i] = mp.Address
	}
	response, err := rp.Api.Minipool.SetUseLatestDelegates(addresses, setting)
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
		fmt.Sprintf("Are you sure you want to change the auto-upgrade setting for %d minipools to %t?", len(selectedMinipools), setting),
		func(i int) string {
			return fmt.Sprintf("toggling auto-upgrades for minipool %s", selectedMinipools[i].Address.Hex())
		},
		"Toggling auto-upgrade for minipools...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully updated the setting for all selected minipools.")
	return nil
}
