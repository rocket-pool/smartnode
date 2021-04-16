package node

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func setWithdrawalAddress(c *cli.Context, withdrawalAddress common.Address) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to set your node's withdrawal address to %s? All future ETH, nETH & RPL rewards/refunds will be sent here.", withdrawalAddress.Hex()))) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Set node's withdrawal address
    response, err := rp.SetNodeWithdrawalAddress(withdrawalAddress)
    if err != nil {
        return err
    }

    fmt.Printf("Setting withdrawal address...\n")
    cliutils.PrintTransactionHash(response.TxHash)
    if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
        return err
    }

    // Log & return
    fmt.Printf("The node's withdrawal address was successfully set to %s.\n", withdrawalAddress.Hex())
    return nil

}

