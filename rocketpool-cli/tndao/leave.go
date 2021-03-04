package tndao

import (
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func leave(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
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
        if cliutils.Confirm(fmt.Sprintf("Would you like to refund your RPL bond to your node account (%s)?", wallet.AccountAddress.Hex())) {
            bondRefundAddress = wallet.AccountAddress
        } else {

            // Prompt for custom address
            inputAddress := cliutils.Prompt("Please enter the address to refund your RPL bond to:", "^0x[0-9a-fA-F]{40}$", "Invalid address")
            bondRefundAddress = common.HexToAddress(inputAddress)

        }

    }

    // Check if node can leave the trusted node DAO
    canLeave, err := rp.CanLeaveTNDAO()
    if err != nil {
        return err
    }
    if !canLeave.CanLeave {
        fmt.Println("Cannot leave the trusted node DAO:")
        if canLeave.ProposalExpired {
            fmt.Println("The proposal for you to leave the trusted node DAO does not exist or has expired.")
        }
        if canLeave.InsufficientMembers {
            fmt.Println("There are not enough members in the trusted node DAO to allow a member to leave.")
        }
        return nil
    }

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to leave the trusted node DAO and refund your RPL bond to %s? This action cannot be undone!", bondRefundAddress.Hex()))) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Leave the trusted node DAO
    if _, err := rp.LeaveTNDAO(bondRefundAddress); err != nil {
        return err
    }

    // Log & return
    fmt.Println("Successfully left the trusted node DAO.")
    return nil

}

