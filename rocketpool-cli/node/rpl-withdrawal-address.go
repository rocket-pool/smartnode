package node

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func setRPLWithdrawalAddress(c *cli.Context, withdrawalAddressOrENS string) error {

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
	fmt.Println("You are about to change your RPL withdrawal address. All future RPL rewards/refunds will be sent there.")
	if !c.Bool("force") {
		confirm = false
		fmt.Println("By default, this will put your new RPL withdrawal address into a \"pending\" state.")
		fmt.Println("Rocket Pool will continue to use your old RPL withdrawal address (or your primary withdrawal address if your RPL withdrawal address has not been set) until you confirm that you own the new address via the Rocket Pool website.")
		fmt.Println("You will need to use a web3-compatible wallet (such as MetaMask) with your new address to confirm it.")
		fmt.Printf("%sIf you cannot use such a wallet, or if you want to bypass this step and force Rocket Pool to use the new address immediately, please re-run this command with the \"--force\" flag.\n\n%s", colorYellow, colorReset)
	} else {
		confirm = true
		fmt.Printf("%sYou have specified the \"--force\" option, so your new address will take effect immediately.\n", colorRed)
		fmt.Printf("Please ensure that you have the correct address - if you do not control the new address, you will not be able to change this once set!%s\n\n", colorReset)
	}

	// Check if the withdrawal address can be set
	canResponse, err := rp.CanSetNodeRPLWithdrawalAddress(withdrawalAddress, confirm)
	if err != nil {
		return err
	}

	// Check if it can be set
	if !canResponse.CanSet {
		fmt.Println("Cannot set RPL withdrawal address:")
		if canResponse.RPLAddressDiffers {
			fmt.Println("The RPL withdrawal address has already been set to something other than the node address. Setting it can only be called from the RPL withdrawal address.")
		} else if canResponse.PrimaryAddressDiffers {
			fmt.Println("The primary withdrawal address has already been set to something other than the node address. Setting the RPL withdrawal address can only be called from the primary withdrawal address.")
		}
		return nil
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

			fmt.Printf("Successfully sent the test transaction.\nPlease verify that your RPL withdrawal address received it before confirming it below.\n\n")
		}
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if canResponse.RPLStake.Cmp(common.Big0) == 1 {
		fmt.Printf("%sNOTE: You currently have %.6f RPL staked. Withdrawing it will *no longer* send it to your primary withdrawal address. It will be sent to the new RPL withdrawal address instead. Please verify you have control over that address before confirming this!%s\n", colorYellow, eth.WeiToEth(canResponse.RPLStake), colorReset)
	}
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to set your node's RPL withdrawal address to %s?", withdrawalAddressString))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Set node's withdrawal address
	response, err := rp.SetNodeRPLWithdrawalAddress(withdrawalAddress, confirm)
	if err != nil {
		return err
	}

	fmt.Printf("Setting RPL withdrawal address...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	if !c.Bool("force") {
		nodeManagerUrl := ""
		config, _, err := rp.LoadConfig()
		if err == nil {
			nodeManagerUrl = config.Smartnode.GetNodeManagerUrl() + "/rpl-withdrawal-address"
		}
		if nodeManagerUrl != "" {
			fmt.Printf("The node's RPL withdrawal address update to %s is now pending.\n"+
				"To confirm it, please visit the Rocket Pool website (%s).", withdrawalAddressString, nodeManagerUrl)
		} else {
			fmt.Printf("The node's RPL withdrawal address update to %s is now pending.\n"+
				"To confirm it, please visit the Rocket Pool website.", withdrawalAddressString)
		}
	} else {
		fmt.Printf("The node's RPL withdrawal address was successfully set to %s.\n", withdrawalAddressString)
	}
	return nil

}

func confirmRPLWithdrawalAddress(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if the withdrawal address can be confirmed
	canResponse, err := rp.CanConfirmNodeRPLWithdrawalAddress()
	if err != nil {
		return err
	}

	// Check if it can be set
	if !canResponse.CanSet {
		fmt.Println("Cannot confirm new RPL withdrawal address: your node address is not the new pending RPL withdrawal address. Confirmation can only be done if it is set to your node address.")
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to confirm your node's address as the new RPL withdrawal address?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Confirm node's withdrawal address
	response, err := rp.ConfirmNodeRPLWithdrawalAddress()
	if err != nil {
		return err
	}

	fmt.Printf("Confirming new RPL withdrawal address...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("The node's RPL withdrawal address was successfully set to the node address.\n")
	return nil

}
