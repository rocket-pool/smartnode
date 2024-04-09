package minipool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

const (
	closeConfirmFlag string = "confirm-slashing"
)

func closeMinipools(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get minipool statuses
	details, err := rp.Api.Minipool.GetCloseDetails()
	if err != nil {
		return err
	}

	// Exit if the fee distributor hasn't been initialized yet
	if !details.Data.IsFeeDistributorInitialized {
		fmt.Println("Minipools cannot be closed until your fee distributor has been initialized.\nPlease run `rocketpool node initialize-fee-distributor` first, then return here to close your minipools.")
		return nil
	}

	closableMinipools := []api.MinipoolCloseDetails{}
	versionTooLowMinipools := []api.MinipoolCloseDetails{}
	balanceLessThanRefundMinipools := []api.MinipoolCloseDetails{}
	unwithdrawnMinipools := []api.MinipoolCloseDetails{}

	for _, mp := range details.Data.Details {
		if mp.IsFinalized {
			// Ignore minipools that are already closed
			continue
		}
		if mp.CanClose {
			closableMinipools = append(closableMinipools, mp)
		} else {
			if mp.Version < 3 {
				versionTooLowMinipools = append(versionTooLowMinipools, mp)
			}
			if mp.Balance.Cmp(mp.Refund) == -1 {
				balanceLessThanRefundMinipools = append(balanceLessThanRefundMinipools, mp)
			}
			if mp.Status != types.MinipoolStatus_Dissolved &&
				mp.BeaconState != beacon.ValidatorState_WithdrawalDone {
				unwithdrawnMinipools = append(unwithdrawnMinipools, mp)
			}
		}
	}

	// Print ineligible ones
	if len(unwithdrawnMinipools) > 0 {
		fmt.Printf("%sNOTE: The following minipools have not had their full balances withdrawn from the Beacon Chain yet:\n", terminal.ColorBlue)
		for _, mp := range unwithdrawnMinipools {
			fmt.Printf("\t%s\n", mp.Address)
		}
		fmt.Printf("\nTo close them, first run `rocketpool minipool exit` on them and wait until their balances have been withdrawn.%s\n\n", terminal.ColorReset)
	}
	if len(versionTooLowMinipools) > 0 {
		fmt.Printf("%sWARNING: The following minipools are using an old delegate and cannot be safely closed:\n", terminal.ColorYellow)
		for _, mp := range versionTooLowMinipools {
			fmt.Printf("\t%s\n", mp.Address)
		}
		fmt.Printf("\nPlease upgrade the delegate for these minipools using `rocketpool minipool delegate-upgrade` in order to close them.%s\n\n", terminal.ColorReset)
	}
	if len(balanceLessThanRefundMinipools) > 0 {
		fmt.Printf("%sWARNING: The following minipools have refunds larger than their current balances and cannot be closed at this time:\n", terminal.ColorYellow)
		for _, mp := range balanceLessThanRefundMinipools {
			fmt.Printf("\t%s\n", mp.Address)
		}
		fmt.Printf("\nIf you have recently exited their validators from the Beacon Chain, please wait until their balances have been sent to the minipools before closing them.%s\n\n", terminal.ColorReset)
	}

	// Check for closable minipools
	if len(closableMinipools) == 0 {
		fmt.Println("No minipools can be closed.")
		return nil
	}

	// Get selected minipools
	options := make([]utils.SelectionOption[api.MinipoolCloseDetails], len(closableMinipools))
	for i, mp := range closableMinipools {
		option := &options[i]
		option.Element = &closableMinipools[i]
		option.ID = fmt.Sprint(mp.Address)
		if mp.Status == types.MinipoolStatus_Dissolved {
			option.Display = fmt.Sprintf("%s (%.6f ETH will be returned)", mp.Address.Hex(), math.RoundDown(eth.WeiToEth(mp.Balance), 6))
		} else {
			option.Display = fmt.Sprintf("%s (%.6f ETH available, %.6f ETH is yours plus a refund of %.6f ETH)", mp.Address.Hex(), math.RoundDown(eth.WeiToEth(mp.Balance), 6), math.RoundDown(eth.WeiToEth(mp.NodeShareOfEffectiveBalance), 6), math.RoundDown(eth.WeiToEth(mp.Refund), 6))
		}
	}
	selectedMinipools, err := utils.GetMultiselectIndices(c, minipoolsFlag, options, "Please select a minipool to close:")
	if err != nil {
		return fmt.Errorf("error determining minipool selection: %w", err)
	}

	// Force confirmation of slashable minipools
	eight := eth.EthToWei(8)
	yellowThreshold := eth.EthToWei(31.5)
	thirtyTwo := eth.EthToWei(32)
	for _, minipool := range selectedMinipools {
		distributableBalance := big.NewInt(0).Sub(minipool.Balance, minipool.Refund)
		if distributableBalance.Cmp(eight) >= 0 {
			if distributableBalance.Cmp(minipool.UserDepositBalance) < 0 {
				// Less than the user deposit balance, ETH + RPL will be slashed
				fmt.Printf("%sWARNING: Minipool %s has a distributable balance of %.6f ETH which is lower than the amount borrowed from the staking pool (%.6f ETH).\nPlease visit the Rocket Pool Discord's #support channel (https://discord.gg/rocketpool) if you are not expecting this.%s\n", terminal.ColorRed, minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(distributableBalance), 6), math.RoundDown(eth.WeiToEth(minipool.UserDepositBalance), 6), terminal.ColorReset)
				if !c.Bool("confirm-slashing") {
					fmt.Printf("\n%sIf you are *sure* you want to close the minipool anyway, rerun this command with the `--confirm-slashing` flag. Doing so WILL RESULT in both your ETH bond and your RPL collateral being slashed.%s\n", terminal.ColorRed, terminal.ColorReset)
					return nil
				} else {
					if !utils.ConfirmWithIAgree(fmt.Sprintf("\n%sYou have the `--confirm-slashing` flag enabled. Closing this minipool WILL RESULT in the complete loss of your initial ETH bond and enough of your RPL stake to cover the losses to the staking pool. Please confirm you understand this and want to continue closing the minipool.%s", terminal.ColorRed, terminal.ColorReset)) {
						fmt.Println("Cancelled.")
						return nil
					}
				}
			} else if distributableBalance.Cmp(yellowThreshold) < 0 {
				// More than the user deposit balance but less than 31.5, ETH will be slashed with a red warning
				if !utils.ConfirmWithIAgree(fmt.Sprintf("%sWARNING: Minipool %s has a distributable balance of %.6f ETH. Closing it in this state WILL RESULT in a loss of ETH. You will only receive %.6f ETH back. Please confirm you understand this and want to continue closing the minipool.%s", terminal.ColorRed, minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(distributableBalance), 6), math.RoundDown(eth.WeiToEth(minipool.NodeShareOfEffectiveBalance), 6), terminal.ColorReset)) {
					fmt.Println("Cancelled.")
					return nil
				}
			} else if distributableBalance.Cmp(thirtyTwo) < 0 {
				// More than 31.5 but less than 32, ETH will be slashed with a yellow warning
				if !utils.Confirm(fmt.Sprintf("%sWARNING: Minipool %s has a distributable balance of %.6f ETH. Closing it in this state WILL RESULT in a loss of ETH. You will only receive %.6f ETH back. Please confirm you understand this and want to continue closing the minipool.%s", terminal.ColorYellow, minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(distributableBalance), 6), math.RoundDown(eth.WeiToEth(minipool.NodeShareOfEffectiveBalance), 6), terminal.ColorReset)) {
					fmt.Println("Cancelled.")
					return nil
				}
			}
		} else {
			fmt.Printf("Cannot close minipool %s: it has an effective balance of %.6f ETH which is too low to close the minipool. Please run `rocketpool minipool distribute-balance` on it instead.\n", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(distributableBalance), 6))
			return nil
		}
	}

	// Build the TXs
	addresses := make([]common.Address, len(selectedMinipools))
	for i, mp := range selectedMinipools {
		addresses[i] = mp.Address
	}
	response, err := rp.Api.Minipool.Close(addresses)
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
		fmt.Sprintf("Are you sure you want to close %d minipools?", len(selectedMinipools)),
		func(i int) string {
			return fmt.Sprintf("closing minipool %s", selectedMinipools[i].Address.Hex())
		},
		"Closing minipools...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully closed all selected minipools.")
	return nil
}
