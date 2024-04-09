package node

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

const (
	setPrimaryWithdrawalAddressForceFlag string = "force"
)

func setPrimaryWithdrawalAddress(c *cli.Context, withdrawalAddressOrENS string) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	var withdrawalAddress common.Address
	var withdrawalAddressString string
	if strings.Contains(withdrawalAddressOrENS, ".") {
		response, err := rp.Api.Node.ResolveEns(common.Address{}, withdrawalAddressOrENS)
		if err != nil {
			return err
		}
		withdrawalAddress = response.Data.Address
		withdrawalAddressString = fmt.Sprintf("%s (%s)", withdrawalAddressOrENS, withdrawalAddress.Hex())
	} else {
		withdrawalAddress, err = input.ValidateAddress("withdrawal address", withdrawalAddressOrENS)
		if err != nil {
			return err
		}
		withdrawalAddressString = withdrawalAddress.Hex()
	}

	// Print the "pending" disclaimer
	var confirm bool
	fmt.Println("You are about to change your primary withdrawal address. All future ETH rewards/refunds will be sent there.\nIf you haven't set your RPL withdrawal address, RPL rewards will be sent there as well.")
	if !c.Bool(setPrimaryWithdrawalAddressForceFlag) {
		confirm = false
		fmt.Println("By default, this will put your new primary withdrawal address into a \"pending\" state.")
		fmt.Println("Rocket Pool will continue to use your old primary withdrawal address until you confirm that you own the new address via the Rocket Pool website.")
		fmt.Println("You will need to use a web3-compatible wallet (such as MetaMask) with your new address to confirm it.")
		fmt.Printf("%sIf you cannot use such a wallet, or if you want to bypass this step and force Rocket Pool to use the new address immediately, please re-run this command with the \"--%s\" flag.\n\n%s", terminal.ColorYellow, setPrimaryWithdrawalAddressForceFlag, terminal.ColorReset)
	} else {
		confirm = true
		fmt.Printf("%sYou have specified the \"--%s\" option, so your new address will take effect immediately.\n", terminal.ColorRed, setPrimaryWithdrawalAddressForceFlag)
		fmt.Printf("Please ensure that you have the correct address - if you do not control the new address, you will not be able to change this once set!%s\n\n", terminal.ColorReset)
	}

	// Check if the withdrawal address can be set
	setResponse, err := rp.Api.Node.SetPrimaryWithdrawalAddress(withdrawalAddress, confirm)
	if err != nil {
		return err
	}

	if confirm {
		// Prompt for a test transaction
		if utils.Confirm("Would you like to send a test transaction to make sure you have the correct address?") {
			inputAmount := utils.Prompt(fmt.Sprintf("Please enter an amount of ETH to send to %s:", withdrawalAddressString), "^\\d+(\\.\\d+)?$", "Invalid amount")
			testAmount, err := strconv.ParseFloat(inputAmount, 64)
			if err != nil {
				return fmt.Errorf("invalid test amount '%s': %w", inputAmount, err)
			}
			amountWei := eth.EthToWei(testAmount)
			sendResponse, err := rp.Api.Node.Send(amountWei, "eth", withdrawalAddress)
			if err != nil {
				return err
			}

			// Make sure they can send the proper amount
			if !sendResponse.Data.CanSend {
				fmt.Println("Cannot send test transaction:")
				if sendResponse.Data.InsufficientBalance {
					fmt.Printf("You do not have %.6f ETH in your node wallet.\n", testAmount)
				}
				return nil
			}

			// Run the TX
			validated, err := tx.HandleTx(c, rp, sendResponse.Data.TxInfo,
				fmt.Sprintf("Please confirm you want to send %f ETH to %s.", testAmount, withdrawalAddressString),
				fmt.Sprintf("sending ETH to %s", withdrawalAddressString),
				fmt.Sprintf("Sending ETH to %s...\n", withdrawalAddressString),
			)
			if err != nil {
				return err
			}
			if validated {
				fmt.Println("Successfully sent the test transaction.\nPlease verify that your primary withdrawal address received it before confirming it below.")
				fmt.Println()
			}

			fmt.Println()
		}
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, setResponse.Data.TxInfo,
		fmt.Sprintf("Are you sure you want to set your node's primary withdrawal address to %s?", withdrawalAddressString),
		"setting node's primary withdrawal address",
		fmt.Sprintf("Setting primary withdrawal address to %s...", withdrawalAddressString),
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	if !c.Bool(setPrimaryWithdrawalAddressForceFlag) {
		stakeUrl := ""
		config, _, err := rp.LoadConfig()
		if err == nil {
			rs := config.GetRocketPoolResources()
			stakeUrl = rs.StakeUrl
		}
		if stakeUrl != "" {
			fmt.Printf("The node's primary withdrawal address update to %s is now pending.\n"+
				"To confirm it, please visit the Rocket Pool website (%s).", withdrawalAddressString, stakeUrl)
		} else {
			fmt.Printf("The node's primary withdrawal address update to %s is now pending.\n"+
				"To confirm it, please visit the Rocket Pool website.", withdrawalAddressString)
		}
	} else {
		fmt.Printf("The node's primary withdrawal address was successfully set to %s.\n", withdrawalAddressString)
	}
	return nil
}
