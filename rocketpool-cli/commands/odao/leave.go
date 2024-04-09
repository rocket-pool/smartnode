package odao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

var leaveRefundAddressFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "refund-address",
	Aliases: []string{"r"},
	Usage:   "The address to refund the node's RPL bond to (or 'node')",
}

func leave(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get the RPL bond refund address
	var bondRefundAddress common.Address
	if c.String(leaveRefundAddressFlag.Name) == "node" {
		// Set bond refund address to node address
		wallet, err := rp.Api.Wallet.Status()
		if err != nil {
			return err
		}
		bondRefundAddress = wallet.Data.WalletStatus.Address.NodeAddress
	} else if c.String(leaveRefundAddressFlag.Name) != "" {
		// Parse bond refund address
		bondRefundAddress = common.HexToAddress(c.String(leaveRefundAddressFlag.Name))
	} else {
		// Get wallet status
		wallet, err := rp.Api.Wallet.Status()
		if err != nil {
			return err
		}

		// Prompt for node address
		address := wallet.Data.WalletStatus.Address.NodeAddress
		if utils.Confirm(fmt.Sprintf("Would you like to refund your RPL bond to your node account (%s)?", address.Hex())) {
			bondRefundAddress = address
		} else {
			// Prompt for custom address
			inputAddress := utils.Prompt("Please enter the address to refund your RPL bond to:", "^0x[0-9a-fA-F]{40}$", "Invalid address")
			bondRefundAddress = common.HexToAddress(inputAddress)
		}
	}

	// Build the TX
	response, err := rp.Api.ODao.Leave(bondRefundAddress)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanLeave {
		fmt.Println("Cannot leave the Oracle DAO:")
		if response.Data.ProposalExpired {
			fmt.Println("The proposal for you to leave the Oracle DAO does not exist or has expired.")
		}
		if response.Data.InsufficientMembers {
			fmt.Println("There are not enough members in the Oracle DAO to allow a member to leave.")
		}
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf("Are you sure you want to leave the oracle DAO and refund your RPL bond to %s? This action cannot be undone!", bondRefundAddress.Hex()),
		"leaving Oracle DAO",
		"Leaving the Oracle DAO...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully left the oracle DAO.")
	return nil
}
