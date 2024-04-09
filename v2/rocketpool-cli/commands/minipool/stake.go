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

func stakeMinipools(c *cli.Context) error {
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

	// Get stakeable minipools
	stakeableMinipools := []api.MinipoolDetails{}
	for _, minipool := range status.Data.Minipools {
		if minipool.CanStake {
			stakeableMinipools = append(stakeableMinipools, minipool)
		}
	}

	// Check for stakeable minipools
	if len(stakeableMinipools) == 0 {
		fmt.Println("No minipools can be staked.")
		return nil
	}

	// Get selected minipools
	options := make([]utils.SelectionOption[api.MinipoolDetails], len(stakeableMinipools))
	for i, mp := range stakeableMinipools {
		option := &options[i]
		option.Element = &stakeableMinipools[i]
		option.ID = fmt.Sprint(mp.Address)
		option.Display = fmt.Sprintf("%s (%s until dissolved)", mp.Address.Hex(), mp.TimeUntilDissolve)
	}
	selectedMinipools, err := utils.GetMultiselectIndices(c, minipoolsFlag, options, "Please select a minipool to stake:")
	if err != nil {
		return fmt.Errorf("error determining minipool selection: %w", err)
	}

	// Build the TXs
	addresses := make([]common.Address, len(selectedMinipools))
	for i, mp := range selectedMinipools {
		addresses[i] = mp.Address
	}
	response, err := rp.Api.Minipool.Stake(addresses)
	if err != nil {
		return fmt.Errorf("error during TX generation: %w", err)
	}

	// Validation
	txs := make([]*eth.TransactionInfo, len(selectedMinipools))
	for i := range selectedMinipools {
		txInfo := response.Data.TxInfos[i]
		txs[i] = txInfo
	}

	fmt.Println()
	fmt.Println("NOTE: Your Validator Client must be restarted after this process so it loads the new validator key.")
	fmt.Println("Since you are manually staking the minipool, this must be done manually.")
	fmt.Println("When you have finished staking all your minipools, please restart your validator.")
	fmt.Println()

	// Run the TXs
	validated, err := tx.HandleTxBatch(c, rp, txs,
		fmt.Sprintf("Are you sure you want to stake %d minipools?", len(selectedMinipools)),
		func(i int) string {
			return fmt.Sprintf("stake of minipool %s", selectedMinipools[i].Address.Hex())
		},
		"Staking minipools...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully staked all selected minipools.")
	return nil
}
