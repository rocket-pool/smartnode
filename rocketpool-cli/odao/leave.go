package odao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func leave(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get the RPL bond refund address
	var bondRefundAddress common.Address
	if c.String("refund-address") == "node" {

		// Set bond refund address to node address
		wallet, err := rp.WalletStatus()
		if err != nil {
			return err
		}
		bondRefundAddress = wallet.AccountAddress

	} else if c.String("refund-address") != "" {

		// Parse bond refund address
		bondRefundAddress = common.HexToAddress(c.String("refund-address"))

	} else {

		// Get wallet status
		wallet, err := rp.WalletStatus()
		if err != nil {
			return err
		}

		// Prompt for node address
		if prompt.Confirm(fmt.Sprintf("Would you like to refund your RPL bond to your node account (%s)?", wallet.AccountAddress.Hex())) {
			bondRefundAddress = wallet.AccountAddress
		} else {

			// Prompt for custom address
			inputAddress := prompt.Prompt("Please enter the address to refund your RPL bond to:", "^0x[0-9a-fA-F]{40}$", "Invalid address")
			bondRefundAddress = common.HexToAddress(inputAddress)

		}

	}

	// Check if node can leave the oracle DAO
	canLeave, err := rp.CanLeaveTNDAO()
	if err != nil {
		return err
	}
	if !canLeave.CanLeave {
		fmt.Println("Cannot leave the oracle DAO:")
		if canLeave.ProposalExpired {
			fmt.Println("The proposal for you to leave the oracle DAO does not exist or has expired.")
		}
		if canLeave.InsufficientMembers {
			fmt.Println("There are not enough members in the oracle DAO to allow a member to leave.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canLeave.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to leave the oracle DAO and refund your RPL bond to %s? This action cannot be undone!", bondRefundAddress.Hex()))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Leave the oracle DAO
	response, err := rp.LeaveTNDAO(bondRefundAddress)
	if err != nil {
		return err
	}

	fmt.Printf("Leaving oracle DAO...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("Successfully left the oracle DAO.")
	return nil

}
