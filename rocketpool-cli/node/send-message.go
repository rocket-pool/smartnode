package node

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

func sendMessage(c *cli.Context, toAddressOrENS string, message []byte) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get the address
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

	// Get the gas estimate
	canSend, err := rp.CanSendMessage(toAddress, message)
	if err != nil {
		return err
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canSend.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to send a message to %s?", toAddressString))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Send tokens
	response, err := rp.SendMessage(toAddress, message)
	if err != nil {
		return err
	}

	fmt.Printf("Sending message to %s...\n", toAddressString)
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully sent message to %s.\n", toAddressString)
	return nil
}
