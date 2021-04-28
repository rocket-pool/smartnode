package odao

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func replace(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Check if node can replace its position in the oracle DAO
    canReplace, err := rp.CanReplaceTNDAOMember()
    if err != nil {
        return err
    }
    if !canReplace.CanReplace {
        fmt.Println("Cannot replace the node's position in the oracle DAO:")
        if canReplace.ProposalExpired {
            fmt.Println("The proposal to replace your node's position in the oracle DAO does not exist or has expired.")
        }
        if canReplace.MemberAlreadyExists {
            fmt.Println("The replacing node is already a member of the oracle DAO.")
        }
        return nil
    }

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to replace your node's position in the oracle DAO? This action cannot be undone!")) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Replace node's position in the oracle DAO
    response, err := rp.ReplaceTNDAOMember()
    if err != nil {
        return err
    }

    fmt.Printf("Replacing position...\n")
    cliutils.PrintTransactionHash(c, response.TxHash)
    if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
        return err
    }

    // Log & return
    fmt.Println("Successfully replaced the node's position in the oracle DAO.")
    return nil

}

