package node

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func nodeSend(c *cli.Context, amountRaw float64, sendAll bool, token string, toAddressOrENS string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get the recipient
	var toAddress common.Address
	var toAddressString string
	if strings.Contains(toAddressOrENS, ".") {
		response, err := rp.ResolveEnsName(toAddressOrENS)
		if err != nil {
			return err
		}
		toAddress = response.Address
		toAddressString = fmt.Sprintf("%s (%s)", toAddressOrENS, toAddress.Hex())
	} else {
		toAddress, err = cliutils.ValidateAddress("to address", toAddressOrENS)
		if err != nil {
			return err
		}
		toAddressString = toAddress.Hex()
	}

	// Handle "send all" mode
	if sendAll {
		return nodeSendAll(c, rp, token, toAddress, toAddressString)
	}

	// Check tokens can be sent
	canSend, err := rp.CanNodeSend(amountRaw, token, toAddress)
	if err != nil {
		return err
	}
	tokenString := fmt.Sprintf("%s (%s)", canSend.TokenSymbol, token)

	if !canSend.CanSend {
		fmt.Println("Cannot send tokens:")
		if canSend.InsufficientBalance {
			if strings.HasPrefix(token, "0x") {
				fmt.Printf("The node's balance of %s is insufficient.\n", tokenString)
			} else {
				fmt.Printf("The node's %s balance is insufficient.\n", token)
			}
		}
		return nil
	}

	// Prompt for confirmation
	if strings.HasPrefix(token, "0x") {
		fmt.Printf("Token address:   %s\n", token)
		fmt.Printf("Token name:      %s\n", canSend.TokenName)
		fmt.Printf("Token symbol:    %s\n", canSend.TokenSymbol)
		fmt.Printf("Node balance:    %.8f %s\n\n", canSend.Balance, canSend.TokenSymbol)
		fmt.Printf("%sWARNING: Please confirm that the above token is the one you intend to send before confirming below!%s\n\n", colorYellow, colorReset)

		if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to send %.8f of %s to %s? This action cannot be undone!", amountRaw, tokenString, toAddressString))) {
			fmt.Println("Cancelled.")
			return nil
		}
	} else {
		fmt.Printf("Node balance:    %.8f %s\n\n", canSend.Balance, token)
		if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to send %.8f %s to %s? This action cannot be undone!", amountRaw, token, toAddressString))) {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canSend.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Send tokens
	response, err := rp.NodeSend(amountRaw, token, toAddress)
	if err != nil {
		return err
	}

	if strings.HasPrefix(token, "0x") {
		fmt.Printf("Sending %s to %s...\n", tokenString, toAddressString)
	} else {
		fmt.Printf("Sending %s to %s...\n", token, toAddressString)
	}
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	if strings.HasPrefix(token, "0x") {
		fmt.Printf("Successfully sent %.6f of %s to %s.\n", amountRaw, tokenString, toAddressString)
	} else {
		fmt.Printf("Successfully sent %.6f %s to %s.\n", amountRaw, token, toAddressString)
	}
	return nil

}

// nodeSendAll sends the entire balance of the specified token to the recipient.
// For ETH, it reserves enough to cover the estimated maximum gas cost.
func nodeSendAll(c *cli.Context, rp *rocketpool.Client, token string, toAddress common.Address, toAddressString string) error {

	// Query balance and gas info using a zero amount
	canSend, err := rp.CanNodeSend(0, token, toAddress)
	if err != nil {
		return err
	}

	if canSend.Balance <= 0 {
		if strings.HasPrefix(token, "0x") {
			fmt.Printf("The node's balance of %s (%s) is zero, nothing to send.\n", canSend.TokenSymbol, token)
		} else {
			fmt.Printf("The node's %s balance is zero, nothing to send.\n", token)
		}
		return nil
	}

	tokenString := fmt.Sprintf("%s (%s)", canSend.TokenSymbol, token)
	amountRaw := canSend.Balance

	if strings.EqualFold(token, "eth") {
		fmt.Printf("Node balance:    %.8f ETH\n", canSend.Balance)
		fmt.Printf("For sending all ETH, we need to estimate the gas costs first.\n")
		// For ETH, determine gas settings first so we can subtract the gas cost from the balance.
		// This may prompt the user to select a gas price.
		g, err := gas.GetMaxFeeAndLimit(canSend.GasInfo, rp, c.Bool("yes"))
		if err != nil {
			return err
		}

		gasCost := g.GetMaxGasCostEth(canSend.GasInfo)
		amountRaw = canSend.Balance - gasCost

		if amountRaw <= 0 {
			fmt.Printf("The node's ETH balance (%.8f ETH) is not enough to cover the gas cost (%.8f ETH).\n", canSend.Balance, gasCost)
			return nil
		}

		fmt.Printf("Node balance:    %.8f ETH\n", canSend.Balance)
		fmt.Printf("Gas reserve:     %.8f ETH\n", gasCost)
		fmt.Printf("Send amount:     %.8f ETH\n\n", amountRaw)

		if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to send %.8f ETH to %s? This action cannot be undone!", amountRaw, toAddressString))) {
			fmt.Println("Cancelled.")
			return nil
		}

		// Assign the gas settings
		g.Assign(rp)

		// For ETH, use NodeSend with the pre-computed amount (balance minus gas reserve).
		response, err := rp.NodeSend(amountRaw, token, toAddress)
		if err != nil {
			return err
		}
		fmt.Printf("Sending %.8f ETH to %s...\n", amountRaw, toAddressString)
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			return err
		}
		fmt.Printf("Successfully sent %.6f ETH to %s.\n", amountRaw, toAddressString)
		return nil
	}

	// For non-ETH tokens, confirm first, then assign gas settings.
	if strings.HasPrefix(token, "0x") {
		fmt.Printf("Token address:   %s\n", token)
		fmt.Printf("Token name:      %s\n", canSend.TokenName)
		fmt.Printf("Token symbol:    %s\n", canSend.TokenSymbol)
		fmt.Printf("Node balance:    %.8f %s\n\n", canSend.Balance, canSend.TokenSymbol)
		fmt.Printf("%sWARNING: Please confirm that the above token is the one you intend to send before confirming below!%s\n\n", colorYellow, colorReset)

		if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to send all %.8f of %s to %s? This action cannot be undone!", amountRaw, tokenString, toAddressString))) {
			fmt.Println("Cancelled.")
			return nil
		}
	} else {
		fmt.Printf("Node balance:    %.8f %s\n\n", canSend.Balance, token)
		if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to send all %.8f %s to %s? This action cannot be undone!", amountRaw, token, toAddressString))) {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canSend.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Use the exact on-chain balance to avoid float64 rounding errors that
	// would cause "transfer amount exceeds balance".
	response, err := rp.NodeSendAll(token, toAddress)
	if err != nil {
		return err
	}

	if strings.HasPrefix(token, "0x") {
		fmt.Printf("Sending %.8f of %s to %s...\n", amountRaw, tokenString, toAddressString)
	} else {
		fmt.Printf("Sending %.8f %s to %s...\n", amountRaw, token, toAddressString)
	}
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	if strings.HasPrefix(token, "0x") {
		fmt.Printf("Successfully sent %.6f of %s to %s.\n", amountRaw, tokenString, toAddressString)
	} else {
		fmt.Printf("Successfully sent %.6f %s to %s.\n", amountRaw, token, toAddressString)
	}
	return nil
}
