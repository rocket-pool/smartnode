package node

import (
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func setWithdrawalAddress(c *cli.Context, withdrawalAddress common.Address) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Set node's withdrawal address
    canResponse, err := rp.CanSetNodeWithdrawalAddress(withdrawalAddress)
    if err != nil {
        return err
    }

    // Prompt for a test transaction
    if cliutils.Confirm("Would you like to send a test transaction to make sure you have the correct address?") {
        inputAmount := cliutils.Prompt(fmt.Sprintf("Please enter an amount of ETH to send to %s:", withdrawalAddress), "^\\d+(\\.\\d+)?$", "Invalid amount")
        testAmount, err := strconv.ParseFloat(inputAmount, 64)
        if err != nil {
            return fmt.Errorf("Invalid test amount '%s': %w\n", inputAmount, err)
        }
        amountWei := eth.EthToWei(testAmount)
        response, err := rp.NodeSend(amountWei, "eth", withdrawalAddress)
        if err != nil {
            return err
        }

        fmt.Printf("Sending ETH to %s...\n", withdrawalAddress.Hex())
        cliutils.PrintTransactionHash(rp, response.TxHash)
        if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
            return err
        }

        fmt.Printf("Successfully sent the test transaction. Please verify that your withdrawal address received it before confirming it below.\n\n")
    }

    // Display gas estimate
    rp.PrintGasInfo(canResponse.GasInfo)

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to set your node's withdrawal address to %s? All future ETH & RPL rewards/refunds will be sent here.", withdrawalAddress.Hex()))) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Set node's withdrawal address
    response, err := rp.SetNodeWithdrawalAddress(withdrawalAddress)
    if err != nil {
        return err
    }

    fmt.Printf("Setting withdrawal address...\n")
    cliutils.PrintTransactionHash(rp, response.TxHash)
    if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
        return err
    }

    // Log & return
    fmt.Printf("The node's withdrawal address was successfully set to %s.\n", withdrawalAddress.Hex())
    return nil

}

