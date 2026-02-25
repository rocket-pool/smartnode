package node

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func setPrimaryWithdrawalAddress(c *cli.Context, withdrawalAddressOrENS string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	var withdrawalAddress common.Address
	var withdrawalAddressString string
	if strings.Contains(withdrawalAddressOrENS, ".") {
		response, err := rp.ResolveEnsName(withdrawalAddressOrENS)
		if err != nil {
			return err
		}
		withdrawalAddress = response.Address
		withdrawalAddressString = fmt.Sprintf("%s (%s)", withdrawalAddressOrENS, withdrawalAddress.Hex())
	} else {
		withdrawalAddress, err = cliutils.ValidateAddress("withdrawal address", withdrawalAddressOrENS)
		if err != nil {
			return err
		}
		withdrawalAddressString = withdrawalAddress.Hex()
	}

	// Print the "pending" disclaimer
	colorReset := "\033[0m"
	colorRed := "\033[31m"
	colorYellow := "\033[33m"
	var confirm bool
	fmt.Println("You are about to change your primary withdrawal address. All future ETH rewards/refunds will be sent there.\nIf you haven't set your RPL withdrawal address, RPL rewards will be sent there as well.")
	if !c.Bool("force") {
		confirm = false
		fmt.Println("By default, this will put your new primary withdrawal address into a \"pending\" state.")
		fmt.Println("Rocket Pool will continue to use your old primary withdrawal address until you confirm that you own the new address via the Rocket Pool website.")
		fmt.Println("You will need to use a web3-compatible wallet (such as MetaMask) with your new address to confirm it.")
		fmt.Printf("%sIf you cannot use such a wallet, or if you want to bypass this step and force Rocket Pool to use the new address immediately, please re-run this command with the \"--force\" flag.\n\n%s", colorYellow, colorReset)
	} else {
		confirm = true
		fmt.Printf("%sYou have specified the \"--force\" option, so your new address will take effect immediately.\n", colorRed)
		fmt.Printf("Please ensure that you have the correct address - if you do not control the new address, you will not be able to change this once set!%s\n\n", colorReset)
	}

	// Check if the withdrawal address can be set
	canResponse, err := rp.CanSetNodePrimaryWithdrawalAddress(withdrawalAddress, confirm)
	if err != nil {
		return err
	}

	if confirm {
		// Prompt for a test transaction
		if prompt.Confirm("Would you like to send a test transaction to make sure you have the correct address?") {
			inputAmount := prompt.Prompt(fmt.Sprintf("Please enter an amount of ETH to send to %s:", withdrawalAddressString), "^\\d+(\\.\\d+)?$", "Invalid amount")
			testAmount, err := strconv.ParseFloat(inputAmount, 64)
			if err != nil {
				return fmt.Errorf("Invalid test amount '%s': %w\n", inputAmount, err)
			}
			canSendResponse, err := rp.CanNodeSend(testAmount, "eth", withdrawalAddress)
			if err != nil {
				return err
			}

			// Assign max fees
			err = gas.AssignMaxFeeAndLimit(canSendResponse.GasInfo, rp, c.Bool("yes"))
			if err != nil {
				return err
			}

			if !prompt.Confirm(fmt.Sprintf("Please confirm you want to send %f ETH to %s.", testAmount, withdrawalAddressString)) {
				fmt.Println("Cancelled.")
				return nil
			}

			sendResponse, err := rp.NodeSend(testAmount, "eth", withdrawalAddress)
			if err != nil {
				return err
			}

			fmt.Printf("Sending ETH to %s...\n", withdrawalAddressString)
			cliutils.PrintTransactionHash(rp, sendResponse.TxHash)
			if _, err = rp.WaitForTransaction(sendResponse.TxHash); err != nil {
				return err
			}

			fmt.Printf("Successfully sent the test transaction.\nPlease verify that your primary withdrawal address received it before confirming it below.\n\n")
		}
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to set your node's primary withdrawal address to %s?", withdrawalAddressString))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Set node's withdrawal address
	response, err := rp.SetNodePrimaryWithdrawalAddress(withdrawalAddress, confirm)
	if err != nil {
		return err
	}

	fmt.Printf("Setting withdrawal address...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	if !c.Bool("force") {
		nodeManagerUrl := ""
		config, _, err := rp.LoadConfig()
		if err == nil {
			nodeManagerUrl = config.Smartnode.GetNodeManagerUrl() + "/primary-withdrawal-address"
		}
		if nodeManagerUrl != "" {
			fmt.Printf("The node's primary withdrawal address update to %s is now pending.\n"+
				"To confirm it, please visit the Rocket Pool website (%s).", withdrawalAddressString, nodeManagerUrl)
		} else {
			fmt.Printf("The node's primary withdrawal address update to %s is now pending.\n"+
				"To confirm it, please visit the Rocket Pool website.", withdrawalAddressString)
		}
	} else {
		fmt.Printf("The node's primary withdrawal address was successfully set to %s.\n", withdrawalAddressString)
	}
	return nil

}

func confirmPrimaryWithdrawalAddress(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if the withdrawal address can be confirmed
	canResponse, err := rp.CanConfirmNodePrimaryWithdrawalAddress()
	if err != nil {
		return err
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to confirm your node's address as the new primary withdrawal address?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Confirm node's withdrawal address
	response, err := rp.ConfirmNodePrimaryWithdrawalAddress()
	if err != nil {
		return err
	}

	fmt.Printf("Confirming new primary withdrawal address...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("The node's primary withdrawal address was successfully set to the node address.\n")
	return nil

}
