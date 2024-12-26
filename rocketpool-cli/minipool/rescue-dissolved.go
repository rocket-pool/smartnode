package minipool

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
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

func rescueDissolved(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
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
	fmt.Println("Please read our guide at https://docs.rocketpool.net/guides/node/rescue-dissolved.html to fully read about the process before continuing.")
	fmt.Println()

	// Validate the amount
	var depositAmount *big.Int
	if c.String("amount") != "" {
		depositAmountEth, err := strconv.ParseFloat(c.String("amount"), 64)
		if err != nil {
			return fmt.Errorf("Invalid deposit amount '%s': %w", c.String("amount"), err)
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
		fmt.Printf("%sWARNING: The following minipools are using an old delegate and cannot be safely rescued:\n", colorYellow)
		for _, mp := range versionTooLowMinipools {
			fmt.Printf("\t%s\n", mp.Address)
		}
		fmt.Printf("\nPlease upgrade the delegate for these minipools using `rocketpool minipool delegate-upgrade` before rescuing them.%s\n\n", colorReset)
	}
	if len(balanceCompletedMinipools) > 0 {
		fmt.Printf("%sNOTE: The following minipools already have 32 ETH or more deposited:\n", colorYellow)
		for _, mp := range balanceCompletedMinipools {
			fmt.Printf("\t%s\n", mp.Address)
		}
		fmt.Printf("\nThese minipools don't need to be rescued.%s\n\n", colorReset)
	}
	if len(invalidBeaconStateMinipools) > 0 {
		fmt.Printf("%sNOTE: The following minipools have an invalid state on the Beacon Chain (expected 'initialized_pending'):\n", colorYellow)
		for _, mp := range invalidBeaconStateMinipools {
			fmt.Printf("\t%s (%s)\n", mp.Address, mp.BeaconState)
		}
		fmt.Printf("\nThese minipools cannot currently be rescued.%s\n\n", colorReset)
	}

	// Check for rescuable minipools
	if len(rescuableMinipools) == 0 {
		fmt.Println("No minipools can currently be rescued.")
		return nil
	}

	fmt.Printf("%sNOTE: the amounts required for completion below use the validator balances according to the Beacon Chain.\nIf you have recently sent a rescue deposit to this minipool, please wait until it has been registered with the Beacon Chain for these remaining amounts to be accurate.%s\n\n", colorYellow, colorReset)

	// Get selected minipools
	var selectedMinipool *api.MinipoolRescueDissolvedDetails
	var rescueAmount *big.Int
	var rescueAmountFloat float64
	if c.String("minipool") == "" {

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
		selected, _ := cliutils.Select("Please select a minipool to refund ETH from:", options)

		// Get minipool
		selectedMinipool = &rescuableMinipools[selected]
		rescueAmount = rescueAmounts[selected]
		rescueAmountFloat = rescueAmountFloats[selected]

	} else {

		// Get matching minipool
		selectedAddress := common.HexToAddress(c.String("minipool"))
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
	if c.String("amount") == "" {

		// Prompt for amount selection
		options := []string{
			fmt.Sprintf("All %.6f ETH required to complete the minipool", rescueAmountFloat),
			"1 ETH",
			"A custom amount",
		}
		selected, _ := cliutils.Select("Please select an amount of ETH to deposit:", options)

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
		inputAmount := cliutils.Prompt("Please enter an amount of ETH to deposit:", "^\\d+(\\.\\d+)?$", "Invalid amount")
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
	err = gas.AssignMaxFeeAndLimit(selectedMinipool.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to deposit %.6f ETH to rescue minipool %s?", math.RoundDown(depositAmountFloat, 6), selectedMinipool.Address.Hex()))) {
		fmt.Println("Cancelled.")
		return nil
	}

	submit := !c.Bool("no-send")

	// Refund minipool
	response, err := rp.RescueDissolvedMinipool(selectedMinipool.Address, depositAmount, submit)
	if err != nil {
		return fmt.Errorf("Could not rescue minipool %s: %s.\n", selectedMinipool.Address.Hex(), err.Error())
	}

	fmt.Printf("Depositing ETH to rescue minipool %s...\n", selectedMinipool.Address.Hex())
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return fmt.Errorf("Could not rescue minipool %s: %s.\n", selectedMinipool.Address.Hex(), err.Error())
	} else {
		fmt.Printf("Successfully deposited to minipool %s.\nPlease watch its status on a chain explorer such as https://beaconcha.in; it may take up to 24 hours for this deposit to be seen by the chain.\n", selectedMinipool.Address.Hex())
	}

	// Return
	return nil

}
