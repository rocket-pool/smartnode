package minipool

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"

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
	rescueMinipoolFlag string = "minipool"
	rescueAmountFlag   string = "amount"
)

func rescueDissolved(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	fmt.Println("This command will allow you to manually deposit the remaining ETH for any dissolved minipools, activating them so you can exit them and retrieve your minipool's funds.\nPlease read our guide at https://docs.rocketpool.net/guides/node/rescue-dissolved.html to fully read about the process before continuing.\n")

	// Get minipool statuses
	details, err := rp.Api.Minipool.GetRescueDissolvedDetails()
	if err != nil {
		return err
	}

	// Validate the amount
	var depositAmount *big.Int
	if c.String(rescueAmountFlag) != "" {
		depositAmountEth, err := strconv.ParseFloat(c.String("amount"), 64)
		if err != nil {
			return fmt.Errorf("Invalid deposit amount '%s': %w", c.String("amount"), err)
		}
		if depositAmountEth < 1 {
			return fmt.Errorf("The minimum amount you can deposit to the Beacon deposit contract is 1 ETH.")
		}
		depositAmount = eth.EthToWei(depositAmountEth)
	}

	// Categorize minipools
	rescuableMinipools := []api.MinipoolRescueDissolvedDetails{}
	versionTooLowMinipools := []api.MinipoolRescueDissolvedDetails{}
	balanceCompletedMinipools := []api.MinipoolRescueDissolvedDetails{}
	invalidBeaconStateMinipools := []api.MinipoolRescueDissolvedDetails{}
	fullDepositAmount := eth.EthToWei(32)
	for _, mp := range details.Data.Details {
		if mp.IsFinalized {
			// Ignore minipools that are already closed
			continue
		}
		if mp.MinipoolState != types.MinipoolStatus_Dissolved {
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
		fmt.Printf("%sWARNING: The following minipools are using an old delegate and cannot be safely rescued:\n", terminal.ColorYellow)
		for _, mp := range versionTooLowMinipools {
			fmt.Printf("\t%s\n", mp.Address)
		}
		fmt.Printf("\nPlease upgrade the delegate for these minipools using `rocketpool minipool delegate-upgrade` before rescuing them.%s\n\n", terminal.ColorReset)
	}
	if len(balanceCompletedMinipools) > 0 {
		fmt.Printf("%NOTE: The following minipools already have 32 ETH or more deposited:\n", terminal.ColorYellow)
		for _, mp := range balanceCompletedMinipools {
			fmt.Printf("\t%s\n", mp.Address)
		}
		fmt.Printf("\nThese minipools don't need to be rescued.%s\n\n", terminal.ColorReset)
	}
	if len(invalidBeaconStateMinipools) > 0 {
		fmt.Printf("%NOTE: The following minipools have an invalid state on the Beacon Chain (expected 'initialized_pending'):\n", terminal.ColorYellow)
		for _, mp := range invalidBeaconStateMinipools {
			fmt.Printf("\t%s (%s)\n", mp.Address, mp.BeaconState)
		}
		fmt.Printf("\nThese minipools cannot currently be rescued.%s\n\n", terminal.ColorReset)
	}

	// Check for rescuable minipools
	if len(rescuableMinipools) == 0 {
		fmt.Println("No minipools can currently be rescued.")
		return nil
	}

	fmt.Printf("%sNOTE: the amounts required for completion below use the validator balances according to the Beacon Chain.\nIf you have recently sent a rescue deposit to this minipool, please wait until it has been registered with the Beacon Chain for these remaining amounts to be accurate.%s\n\n", terminal.ColorYellow, terminal.ColorReset)

	// Get selected minipools
	var selectedMinipool *api.MinipoolRescueDissolvedDetails
	var rescueAmount *big.Int
	var rescueAmountFloat float64
	if c.String(rescueMinipoolFlag) == "" {
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
		selected, _ := utils.Select("Please select a minipool to refund ETH from:", options)

		// Get minipool
		selectedMinipool = &rescuableMinipools[selected]
		rescueAmount = rescueAmounts[selected]
		rescueAmountFloat = rescueAmountFloats[selected]
	} else {
		// Get matching minipool
		selectedAddress := common.HexToAddress(c.String(rescueMinipoolFlag))
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
	if c.String(rescueAmountFlag) == "" {
		// Prompt for amount selection
		options := []string{
			fmt.Sprintf("All %.6f ETH required to complete the minipool", rescueAmountFloat),
			"1 ETH",
			"A custom amount",
		}
		selected, _ := utils.Select("Please select an amount of ETH to deposit:", options)

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
		inputAmount := utils.Prompt("Please enter an amount of ETH to deposit:", "^\\d+(\\.\\d+)?$", "Invalid amount")
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

	// Build the TX
	response, err := rp.Api.Minipool.RescueDissolved([]common.Address{selectedMinipool.Address}, []*big.Int{depositAmount})
	if err != nil {
		return fmt.Errorf("error during TX generation: %w", err)
	}

	// Rescue minipool
	txInfo := response.Data.TxInfos[0]
	validated, err := tx.HandleTx(c, rp, txInfo,
		fmt.Sprintf("Are you sure you want to deposit %.6f ETH to rescue minipool %s?", math.RoundDown(depositAmountFloat, 6), selectedMinipool.Address.Hex()),
		fmt.Sprintf("rescue of minipool %s", selectedMinipool.Address.Hex()),
		fmt.Sprintf("Depositing ETH to rescue minipool %s...\n", selectedMinipool.Address.Hex()),
	)
	if err != nil {
		return fmt.Errorf("could not rescue minipool %s: %s\n", selectedMinipool.Address.Hex(), err.Error())
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("Successfully deposited to minipool %s.\nPlease watch its status on a chain explorer such as https://beaconcha.in; it may take up to 24 hours for this deposit to be seen by the chain.\n", selectedMinipool.Address.Hex())
	return nil
}
