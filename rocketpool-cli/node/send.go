package node

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func nodeSend(c *cli.Context, amount float64, token string, toAddressOrENS string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get amount in wei
	amountWei := eth.EthToWei(amount)

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

	// Check tokens can be sent
	canSend, err := rp.CanNodeSend(amountWei, token, toAddress)
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
		fmt.Printf("Node balance:    %.6f %s\n\n", eth.WeiToEth(canSend.Balance), canSend.TokenSymbol)
		fmt.Printf("%sWARNING: Please confirm that the above token is the one you intend to send before confirming below!%s\n\n", colorYellow, colorReset)

		if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to send %.6f of %s to %s? This action cannot be undone!", math.RoundDown(eth.WeiToEth(amountWei), 6), tokenString, toAddressString))) {
			fmt.Println("Cancelled.")
			return nil
		}
	} else {
		fmt.Printf("Node balance:    %.6f %s\n\n", eth.WeiToEth(canSend.Balance), token)
		if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to send %.6f %s to %s? This action cannot be undone!", math.RoundDown(eth.WeiToEth(amountWei), 6), token, toAddressString))) {
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
	response, err := rp.NodeSend(amountWei, token, toAddress)
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
		fmt.Printf("Successfully sent %.6f of %s to %s.\n", math.RoundDown(eth.WeiToEth(amountWei), 6), tokenString, toAddressString)
	} else {
		fmt.Printf("Successfully sent %.6f %s to %s.\n", math.RoundDown(eth.WeiToEth(amountWei), 6), token, toAddressString)
	}
	return nil

}
