package odao

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)


func join(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get node status
    status, err := rp.NodeStatus()
    if err != nil {
        return err
    }

    // Check for fixed-supply RPL balance
    if status.AccountBalances.FixedSupplyRPL.Cmp(big.NewInt(0)) > 0 {

        // Confirm swapping RPL
        if (c.Bool("swap") || cliutils.Confirm(fmt.Sprintf("The node has a balance of %.6f old RPL. Would you like to swap it for new RPL before transferring your bond?", math.RoundDown(eth.WeiToEth(status.AccountBalances.FixedSupplyRPL), 6)))) {

            // Approve RPL for swapping
            response, err := rp.NodeSwapRplApprove(status.AccountBalances.FixedSupplyRPL)
            if err != nil {
                return err
            }
            hash := response.ApproveTxHash
            fmt.Printf("Approving old RPL for swap...\n")
            cliutils.PrintTransactionHash(rp, hash)
            
            // Swap RPL
            swapResponse, err := rp.NodeSwapRpl(status.AccountBalances.FixedSupplyRPL, hash)
            if err != nil {
                return err
            }
            fmt.Printf("Swapping old RPL for new RPL...\n")
            cliutils.PrintTransactionHash(rp, swapResponse.SwapTxHash)
            if _, err = rp.WaitForTransaction(swapResponse.SwapTxHash); err != nil {
                return err
            }

            // log
            fmt.Printf("Successfully swapped %.6f old RPL for new RPL.\n", math.RoundDown(eth.WeiToEth(status.AccountBalances.FixedSupplyRPL), 6))
            fmt.Println("")

        }

    }

    // Check if node can join the oracle DAO
    canJoin, err := rp.CanJoinTNDAO()
    if err != nil {
        return err
    }
    if !canJoin.CanJoin {
        fmt.Println("Cannot join the oracle DAO:")
        if canJoin.ProposalExpired {
            fmt.Println("The proposal for you to join the oracle DAO does not exist or has expired.")
        }
        if canJoin.AlreadyMember {
            fmt.Println("The node is already a member of the oracle DAO.")
        }
        if canJoin.InsufficientRplBalance {
            fmt.Println("The node does not have enough RPL to pay the RPL bond.")
        }
        return nil
    }

    // Display gas estimate
    rp.PrintGasInfo(canJoin.GasInfo)
    rp.PrintMultiTxWarning()

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to join the oracle DAO? Your RPL bond will be locked until you leave.")) {
        fmt.Println("Cancelled.")
        return nil
    }
    
    // Approve RPL for joining the ODAO
    response, err := rp.ApproveRPLToJoinTNDAO()
    if err != nil {
        return err
    }
    hash := response.ApproveTxHash
    fmt.Printf("Approving RPL for joining the Oracle DAO...\n")
    cliutils.PrintTransactionHashNoCancel(rp, hash)

    // Join the ODAO
    joinResponse, err := rp.JoinTNDAO(hash)
    if err != nil {
        return err
    }
    fmt.Printf("Joining the ODAO...\n")
    cliutils.PrintTransactionHash(rp, joinResponse.JoinTxHash)
    if _, err = rp.WaitForTransaction(joinResponse.JoinTxHash); err != nil {
        return err
    }

    // Log & return
    fmt.Println("Successfully joined the oracle DAO.")
    return nil

}

