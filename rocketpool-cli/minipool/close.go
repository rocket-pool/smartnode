package minipool

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	rocketpoolapi "github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

const (
	colorBlue string = "\033[36m"
)

func closeMinipools(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get minipool statuses
	details, err := rp.GetMinipoolCloseDetailsForNode()
	if err != nil {
		return err
	}

	// Exit if the fee distributor hasn't been initialized yet
	if !details.IsFeeDistributorInitialized {
		fmt.Println("Minipools cannot be closed until your fee distributor has been initialized.\nPlease run `rocketpool node initialize-fee-distributor` first, then return here to close your minipools.")
		return nil
	}

	if !details.IsVotingInitialized {
		fmt.Println("Minipools should not be closed until your node has initialized voting power. \nPlease run `rocketpool pdao initialize-voting` first, then return here to close your minipools.")
	}

	closableMinipools := []api.MinipoolCloseDetails{}
	versionTooLowMinipools := []api.MinipoolCloseDetails{}
	balanceLessThanRefundMinipools := []api.MinipoolCloseDetails{}
	unwithdrawnMinipools := []api.MinipoolCloseDetails{}

	for _, mp := range details.Details {
		if mp.IsFinalized {
			// Ignore minipools that are already closed
			continue
		}
		if mp.MinipoolStatus == types.Prelaunch {
			// Ignore minipools that are currently in prelaunch
			continue
		}
		if mp.CanClose {
			closableMinipools = append(closableMinipools, mp)
		} else {
			if mp.MinipoolVersion < 3 {
				versionTooLowMinipools = append(versionTooLowMinipools, mp)
			}
			if mp.Balance.Cmp(mp.Refund) == -1 {
				balanceLessThanRefundMinipools = append(balanceLessThanRefundMinipools, mp)
			}
			if mp.MinipoolStatus != types.Dissolved &&
				mp.BeaconState != beacon.ValidatorState_WithdrawalDone {
				unwithdrawnMinipools = append(unwithdrawnMinipools, mp)
			}
		}
	}

	// Print ineligible ones
	if len(unwithdrawnMinipools) > 0 {
		fmt.Printf("%sNOTE: The following minipools have not had their full balances withdrawn from the Beacon Chain yet:\n", colorBlue)
		for _, mp := range unwithdrawnMinipools {
			fmt.Printf("\t%s\n", mp.Address)
		}
		fmt.Printf("\nTo close them, first run `rocketpool minipool exit` on them and wait until their balances have been withdrawn.%s\n\n", colorReset)
	}
	if len(versionTooLowMinipools) > 0 {
		fmt.Printf("%sWARNING: The following minipools are using an old delegate and cannot be safely closed:\n", colorYellow)
		for _, mp := range versionTooLowMinipools {
			fmt.Printf("\t%s\n", mp.Address)
		}
		fmt.Printf("\nPlease upgrade the delegate for these minipools using `rocketpool minipool delegate-upgrade` in order to close them.%s\n\n", colorReset)
	}
	if len(balanceLessThanRefundMinipools) > 0 {
		fmt.Printf("%sWARNING: The following minipools have refunds larger than their current balances and cannot be closed at this time:\n", colorYellow)
		for _, mp := range balanceLessThanRefundMinipools {
			fmt.Printf("\t%s\n", mp.Address)
		}
		fmt.Printf("\nIf you have recently exited their validators from the Beacon Chain, please wait until their balances have been sent to the minipools before closing them.%s\n\n", colorReset)
	}

	// Check for closable minipools
	if len(closableMinipools) == 0 {
		fmt.Println("No minipools can be closed.")
		return nil
	}

	// Get selected minipools
	var selectedMinipools []api.MinipoolCloseDetails
	if c.String("minipool") == "" {

		// Prompt for minipool selection
		options := make([]string, len(closableMinipools)+1)
		options[0] = "All available minipools"
		for mi, minipool := range closableMinipools {
			if minipool.MinipoolStatus == types.Dissolved {
				options[mi+1] = fmt.Sprintf("%s (%.6f ETH will be returned)", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(minipool.Balance), 6))
			} else {
				options[mi+1] = fmt.Sprintf("%s (%.6f ETH available, %.6f ETH is yours plus a refund of %.6f ETH)", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(minipool.Balance), 6), math.RoundDown(eth.WeiToEth(minipool.NodeShare), 6), math.RoundDown(eth.WeiToEth(minipool.Refund), 6))
			}
		}
		selected, _ := cliutils.Select("Please select a minipool to close:", options)

		// Get minipools
		if selected == 0 {
			selectedMinipools = closableMinipools
		} else {
			selectedMinipools = []api.MinipoolCloseDetails{closableMinipools[selected-1]}
		}

	} else {

		// Get matching minipools
		if c.String("minipool") == "all" {
			selectedMinipools = closableMinipools
		} else {
			selectedAddress := common.HexToAddress(c.String("minipool"))
			for _, minipool := range closableMinipools {
				if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
					selectedMinipools = []api.MinipoolCloseDetails{minipool}
					break
				}
			}
			if selectedMinipools == nil {
				return fmt.Errorf("The minipool %s is not available for closing.", selectedAddress.Hex())
			}
		}

	}

	// Force confirmation of slashable minipools
	eight := eth.EthToWei(8)
	yellowThreshold := eth.EthToWei(31.5)
	thirtyTwo := eth.EthToWei(32)
	for _, minipool := range selectedMinipools {
		// Dissolved minipools can always be closed
		if minipool.MinipoolStatus == types.Dissolved {
			continue
		}
		// Check the distributableBalance, minus any refunds
		distributableBalance := big.NewInt(0).Sub(minipool.Balance, minipool.Refund)
		// If it's under 8, it shouldn't be closed, and must be distributed.
		if distributableBalance.Cmp(eight) < 0 {
			fmt.Printf("Cannot close minipool %s: it has an effective balance of %.6f ETH which is too low to close the minipool. Please run `rocketpool minipool distribute-balance` on it instead.\n", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(distributableBalance), 6))
			return nil
		}

		// If there isn't enough eth to pay back rETH holders, warn that RPL and ETH will both be penalized
		if distributableBalance.Cmp(minipool.UserDepositBalance) < 0 {
			// Less than the user deposit balance, ETH + RPL will be slashed
			fmt.Printf("%sWARNING: Minipool %s has a distributable balance of %.6f ETH which is lower than the amount borrowed from the staking pool (%.6f ETH).\nPlease visit the Rocket Pool Discord's #support channel (https://discord.gg/rocketpool) if you are not expecting this.%s\n", colorRed, minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(distributableBalance), 6), math.RoundDown(eth.WeiToEth(minipool.UserDepositBalance), 6), colorReset)
			if !c.Bool("confirm-slashing") {
				fmt.Printf("\n%sIf you are *sure* you want to close the minipool anyway, rerun this command with the `--confirm-slashing` flag. Doing so WILL RESULT in both your ETH bond and your RPL collateral being slashed.%s\n", colorRed, colorReset)
				return nil
			}
			if !cliutils.ConfirmWithIAgree(fmt.Sprintf("\n%sYou have the `--confirm-slashing` flag enabled. Closing this minipool WILL RESULT in the complete loss of your initial ETH bond and enough of your RPL stake to cover the losses to the staking pool. Please confirm you understand this and want to continue closing the minipool.%s", colorRed, colorReset)) {
				fmt.Println("Cancelled.")
				return nil
			}
			// User has confirmed they know they are about to be slashed
			continue
		}

		if distributableBalance.Cmp(yellowThreshold) < 0 {
			// More than the user deposit balance but less than 31.5, ETH will be slashed with a red warning
			if !cliutils.ConfirmWithIAgree(fmt.Sprintf("%sWARNING: Minipool %s has a distributable balance of %.6f ETH. Closing it in this state WILL RESULT in a loss of ETH. You will only receive %.6f ETH back. Please confirm you understand this and want to continue closing the minipool.%s", colorRed, minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(distributableBalance), 6), math.RoundDown(eth.WeiToEth(minipool.NodeShare), 6), colorReset)) {
				fmt.Println("Cancelled.")
				return nil
			}
			// User has confirmed they know they are about to be slashed
			continue
		}
		if distributableBalance.Cmp(thirtyTwo) < 0 {
			// More than 31.5 but less than 32, ETH will be slashed with a yellow warning
			if !cliutils.Confirm(fmt.Sprintf("%sWARNING: Minipool %s has a distributable balance of %.6f ETH. Closing it in this state WILL RESULT in a loss of ETH. You will only receive %.6f ETH back. Please confirm you understand this and want to continue closing the minipool.%s", colorYellow, minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(distributableBalance), 6), math.RoundDown(eth.WeiToEth(minipool.NodeShare), 6), colorReset)) {
				fmt.Println("Cancelled.")
				return nil
			}
			// User has confirmed they know they are about to be slashed
			continue
		}

		// Node Operator has greater than 32 eth in the minpool contract and can safely close the minipool
		continue
	}

	// Get the total gas limit estimate
	var gasInfo rocketpoolapi.GasInfo
	for _, minipool := range selectedMinipools {
		gasInfo.EstGasLimit += minipool.GasInfo.EstGasLimit
		gasInfo.SafeGasLimit += minipool.GasInfo.SafeGasLimit
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(gasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to close %d minipools?", len(selectedMinipools)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Close minipools
	for _, minipool := range selectedMinipools {

		response, err := rp.CloseMinipool(minipool.Address)
		if err != nil {
			fmt.Printf("Could not close minipool %s: %s.\n", minipool.Address.Hex(), err.Error())
			continue
		}

		fmt.Printf("Closing minipool %s...\n", minipool.Address.Hex())
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			fmt.Printf("Could not close minipool %s: %s.\n", minipool.Address.Hex(), err.Error())
		} else {
			fmt.Printf("Successfully closed minipool %s.\n", minipool.Address.Hex())
		}
	}

	// Return
	return nil

}
