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
	setRplWithdrawalAddressForceFlag string = "force"
)

func setRplWithdrawalAddress(c *cli.Context, withdrawalAddressOrEns string) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	var withdrawalAddress common.Address
	var withdrawalAddressString string
	if strings.Contains(withdrawalAddressOrEns, ".") {
		response, err := rp.Api.Node.ResolveEns(common.Address{}, withdrawalAddressOrEns)
		if err != nil {
			return err
		}
		withdrawalAddress = response.Data.Address
		withdrawalAddressString = fmt.Sprintf("%s (%s)", withdrawalAddressOrEns, withdrawalAddress.Hex())
	} else {
		withdrawalAddress, err = input.ValidateAddress("withdrawal address", withdrawalAddressOrEns)
		if err != nil {
			return err
		}
		withdrawalAddressString = withdrawalAddress.Hex()
	}

	// Print the "pending" disclaimer
	var confirm bool
	fmt.Println("You are about to change your RPL withdrawal address. All future RPL rewards/refunds will be sent there.")
	if !c.Bool(setRplWithdrawalAddressForceFlag) {
		confirm = false
		fmt.Println("By default, this will put your new RPL withdrawal address into a \"pending\" state.")
		fmt.Println("Rocket Pool will continue to use your old RPL withdrawal address (or your primary withdrawal address if your RPL withdrawal address has not been set) until you confirm that you own the new address via the Rocket Pool website.")
		fmt.Println("You will need to use a web3-compatible wallet (such as MetaMask) with your new address to confirm it.")
		fmt.Printf("%sIf you cannot use such a wallet, or if you want to bypass this step and force Rocket Pool to use the new address immediately, please re-run this command with the \"--%s\" flag.\n\n%s", terminal.ColorYellow, setRplWithdrawalAddressForceFlag, terminal.ColorReset)
	} else {
		confirm = true
		fmt.Printf("%sYou have specified the \"--%s\" option, so your new address will take effect immediately.\n", terminal.ColorRed, setRplWithdrawalAddressForceFlag)
		fmt.Printf("Please ensure that you have the correct address - if you do not control the new address, you will not be able to change this once set!%s\n\n", terminal.ColorReset)
	}

	// Check if the withdrawal address can be set
	response, err := rp.Api.Node.SetRplWithdrawalAddress(withdrawalAddress, confirm)
	if err != nil {
		return err
	}

	// Check if it can be set
	if !response.Data.CanSet {
		fmt.Println("Cannot set RPL withdrawal address:")
		if response.Data.RplAddressDiffers {
			fmt.Println("The RPL withdrawal address has already been set to something other than the node address. Setting it can only be called from the RPL withdrawal address.")
		} else if response.Data.PrimaryAddressDiffers {
			fmt.Println("The primary withdrawal address has already been set to something other than the node address. Setting the RPL withdrawal address can only be called from the primary withdrawal address.")
		}
		return nil
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
				fmt.Printf("Successfully sent the test transaction.\nPlease verify that your RPL withdrawal address received it before confirming it below.")
				fmt.Println()
			}

			fmt.Println()
		}
	}

	// Note about existing RPL
	if response.Data.RplStake.Cmp(common.Big0) == 1 {
		fmt.Printf("%sNOTE: You currently have %.6f RPL staked. Withdrawing it will *no longer* send it to your primary withdrawal address. It will be sent to the new RPL withdrawal address instead. Please verify you have control over that address before confirming this!%s\n", terminal.ColorYellow, eth.WeiToEth(response.Data.RplStake), terminal.ColorReset)
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf("Are you sure you want to set your node's RPL withdrawal address to %s?", withdrawalAddressString),
		"setting RPL withdrawal address",
		"Setting RPL withdrawal address...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	if !c.Bool(setRplWithdrawalAddressForceFlag) {
		stakeUrl := ""
		config, _, err := rp.LoadConfig()
		if err == nil {
			rs := config.GetRocketPoolResources()
			stakeUrl = rs.StakeUrl
		}
		if stakeUrl != "" {
			fmt.Printf("The node's RPL withdrawal address update to %s is now pending.\n"+
				"To confirm it, please visit the Rocket Pool website (%s).", withdrawalAddressString, stakeUrl)
		} else {
			fmt.Printf("The node's RPL withdrawal address update to %s is now pending.\n"+
				"To confirm it, please visit the Rocket Pool website.", withdrawalAddressString)
		}
	} else {
		fmt.Printf("The node's RPL withdrawal address was successfully set to %s.\n", withdrawalAddressString)
	}
	return nil
}
