package minipool

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func rescueDissolved(minipool string, amount string, noSend bool, yes bool) error {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get minipool statuses
	details, err := rp.GetMinipoolRescueDissolvedDetailsForNode()
	if err != nil {
		return err
	}

	fmt.Println("This command will allow you to manually deposit the remaining ETH for any dissolved minipools, activating them so you can exit them and retrieve your minipool's funds.")
	fmt.Println("Please read our guide at https://docs.rocketpool.net/node-staking/rescue-dissolved to fully read about the process before continuing.")
	fmt.Println()

	// Validate the amount
	var depositAmount *big.Int
	if amount != "" {
		depositAmountEth, err := strconv.ParseFloat(amount, 64)
		if err != nil {
			return fmt.Errorf("Invalid deposit amount '%s': %w", amount, err)
		}
		if depositAmountEth < 1 {
			return fmt.Errorf("The minimum amount you can deposit to the Beacon deposit contract is 1 ETH.")
		}
		depositAmount = eth.EthToWei(depositAmountEth)
	}

	rescuableMinipools := []api.MinipoolRescueDissolvedDetails{}
	versionTooLowMinipools := []api.MinipoolRescueDissolvedDetails{}
	balanceCompletedMinipools := []api.MinipoolRescueDissolvedDetails{}
	invalidBeaconStateMinipools := []api.MinipoolRescueDissolvedDetails{}

	fullDepositAmount := eth.EthToWei(32)
	for _, mp := range details.Details {
		if mp.IsFinalized {
			// Ignore minipools that are already closed
			continue
		}
		if mp.MinipoolStatus != types.Dissolved {
			// Ignore minipools that are aren't dissolved
			continue
		}
		if mp.MinipoolVersion < 3 {
			versionTooLowMinipools = append(versionTooLowMinipools, mp)
			continue
		}
		if mp.BeaconBalance.Cmp(fullDepositAmount) >= 0 {
			balanceCompletedMinipools = append(balanceCompletedMinipools, mp)
			continue
		}
		if mp.BeaconState != beacon.ValidatorState_PendingInitialized {
			invalidBeaconStateMinipools = append(invalidBeaconStateMinipools, mp)
			continue
		}

		// Passed all the status checks, so it's eligible for rescue
		rescuableMinipools = append(rescuableMinipools, mp)
	}

	// Print ineligible ones
	if len(versionTooLowMinipools) > 0 {
		color.YellowPrintf("WARNING: The following minipools are using an old delegate and cannot be safely rescued:\n")
		for _, mp := range versionTooLowMinipools {
			color.YellowPrintf("\t%s\n", mp.Address)
		}
		fmt.Println()
		color.YellowPrintln("Please upgrade the delegate for these minipools using `rocketpool minipool delegate-upgrade` before rescuing them.")
		fmt.Println()
	}
	if len(balanceCompletedMinipools) > 0 {
		color.YellowPrintf("NOTE: The following minipools already have 32 ETH or more deposited:\n")
		for _, mp := range balanceCompletedMinipools {
			color.YellowPrintf("\t%s\n", mp.Address)
		}
		fmt.Println()
		color.YellowPrintln("These minipools don't need to be rescued.")
		fmt.Println()
	}
	if len(invalidBeaconStateMinipools) > 0 {
		color.YellowPrintf("NOTE: The following minipools have an invalid state on the Beacon Chain (expected 'initialized_pending'):\n")
		for _, mp := range invalidBeaconStateMinipools {
			color.YellowPrintf("\t%s (%s)\n", mp.Address, mp.BeaconState)
		}
		fmt.Println()
		color.YellowPrintln("These minipools cannot currently be rescued.")
		fmt.Println()
	}

	// Check for rescuable minipools
	if len(rescuableMinipools) == 0 {
		fmt.Println("No minipools can currently be rescued.")
		return nil
	}

	color.YellowPrintln("NOTE: the amounts required for completion below use the validator balances according to the Beacon Chain.")
	color.YellowPrintln("If you have recently sent a rescue deposit to this minipool, please wait until it has been registered with the Beacon Chain for these remaining amounts to be accurate.")
	fmt.Println()

	// Get selected minipools
	var selectedMinipool *api.MinipoolRescueDissolvedDetails
	var rescueAmount *big.Int
	var rescueAmountFloat float64
	if minipool == "" {

		// Prompt for minipool selection
		options := make([]string, len(rescuableMinipools))
		rescueAmounts := make([]*big.Int, len(rescuableMinipools))
		rescueAmountFloats := make([]float64, len(rescuableMinipools))

		for mi, minipool := range rescuableMinipools {
			localRescueAmount := big.NewInt(0)
			localRescueAmount.Sub(fullDepositAmount, minipool.BeaconBalance)
			rescueAmounts[mi] = localRescueAmount
			rescueAmountFloats[mi] = math.RoundDown(eth.WeiToEth(localRescueAmount), 6)
			options[mi] = fmt.Sprintf("%s (requires %.6f more ETH)", minipool.Address.Hex(), rescueAmountFloats[mi])
		}
		selected, _ := prompt.Select("Please select a minipool to refund ETH from:", options)

		// Get minipool
		selectedMinipool = &rescuableMinipools[selected]
		rescueAmount = rescueAmounts[selected]
		rescueAmountFloat = rescueAmountFloats[selected]

	} else {

		// Get matching minipool
		selectedAddress := common.HexToAddress(minipool)
		for i, minipool := range rescuableMinipools {
			if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
				selectedMinipool = &rescuableMinipools[i]
				rescueAmount = big.NewInt(0)
				rescueAmount.Sub(fullDepositAmount, selectedMinipool.BeaconBalance)
				rescueAmountFloat = math.RoundDown(eth.WeiToEth(rescueAmount), 6)
				break
			}
		}
		if selectedMinipool == nil {
			return fmt.Errorf("The minipool %s is not available for rescue.", selectedAddress.Hex())
		}

	}

	fmt.Println()

	// Get the amount to deposit
	var depositAmountFloat float64
	if amount == "" {

		// Prompt for amount selection
		options := []string{
			fmt.Sprintf("All %.6f ETH required to complete the minipool", rescueAmountFloat),
			"1 ETH",
			"A custom amount",
		}
		selected, _ := prompt.Select("Please select an amount of ETH to deposit:", options)

		switch selected {
		case 0:
			depositAmount = rescueAmount
			depositAmountFloat = rescueAmountFloat
		case 1:
			depositAmount = eth.EthToWei(1)
			depositAmountFloat = 1
		}

	}

	// Prompt for custom amount
	if depositAmount == nil {
		inputAmount := prompt.Prompt("Please enter an amount of ETH to deposit:", "^\\d+(\\.\\d+)?$", "Invalid amount")
		depositAmountEth, err := strconv.ParseFloat(inputAmount, 64)
		if err != nil {
			return fmt.Errorf("Invalid deposit amount '%s': %w", inputAmount, err)
		}
		if depositAmountEth < 1 {
			return fmt.Errorf("The minimum amount you can deposit to the Beacon deposit contract is 1 ETH.")
		}
		depositAmount = eth.EthToWei(depositAmountEth)
		depositAmountFloat = depositAmountEth
	}

	// Assign max fee
	err = gas.AssignMaxFeeAndLimit(selectedMinipool.GasInfo, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if prompt.Declined(yes, "Are you sure you want to deposit %.6f ETH to rescue minipool %s?", math.RoundDown(depositAmountFloat, 6), selectedMinipool.Address.Hex()) {
		fmt.Println("Cancelled.")
		return nil
	}

	submit := !noSend

	// Refund minipool
	response, err := rp.RescueDissolvedMinipool(selectedMinipool.Address, depositAmount, submit)
	if err != nil {
		return fmt.Errorf("Could not rescue minipool %s: %s.\n", selectedMinipool.Address.Hex(), err.Error())
	}

	fmt.Printf("Depositing ETH to rescue minipool %s...\n", selectedMinipool.Address.Hex())
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return fmt.Errorf("Could not rescue minipool %s: %s.\n", selectedMinipool.Address.Hex(), err.Error())
	}
	fmt.Printf("Successfully deposited to minipool %s.\n", selectedMinipool.Address.Hex())
	fmt.Println("Please watch its status on a chain explorer such as https://beaconcha.in; it may take up to 24 hours for this deposit to be seen by the chain.")

	return nil

}
