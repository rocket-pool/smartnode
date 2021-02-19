package minipool

import (
    "fmt"

    "github.com/rocket-pool/rocketpool-go/types"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/hex"
    "github.com/rocket-pool/smartnode/shared/utils/math"
)


func getStatus(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get minipool statuses
    status, err := rp.MinipoolStatus()
    if err != nil {
        return err
    }

    // Get minipools by status
    statusMinipools := map[string][]api.MinipoolDetails{}
    refundableMinipools := []api.MinipoolDetails{}
    withdrawableMinipools := []api.MinipoolDetails{}
    closeableMinipools := []api.MinipoolDetails{}
    for _, minipool := range status.Minipools {

        // Add to status list
        statusName := minipool.Status.Status.String()
        if _, ok := statusMinipools[statusName]; !ok {
            statusMinipools[statusName] = []api.MinipoolDetails{}
        }
        statusMinipools[statusName] = append(statusMinipools[statusName], minipool)

        // Add to actionable lists
        if minipool.RefundAvailable {
            refundableMinipools = append(refundableMinipools, minipool)
        }
        if minipool.WithdrawalAvailable {
            withdrawableMinipools = append(withdrawableMinipools, minipool)
        }
        if minipool.CloseAvailable {
            closeableMinipools = append(closeableMinipools, minipool)
        }

    }

    // Print minipool details by status
    if len(status.Minipools) == 0 {
        fmt.Println("The node does not have any minipools yet.")
    }
    for _, statusName := range types.MinipoolStatuses {
        minipools, ok := statusMinipools[statusName]
        if !ok { continue }

        // Minipool status count & description
        fmt.Printf("%d %s minipool(s):\n", len(minipools), statusName)
        if statusName == "Withdrawable" {
            fmt.Println("(Withdrawal may not be available until after withdrawal delay)")
        }
        fmt.Println("")

        // Minipools
        for _, minipool := range minipools {

            // Main details
            fmt.Printf("-----------------\n")
            fmt.Printf("\n")
            fmt.Printf("Address:              %s\n", minipool.Address.Hex())
            fmt.Printf("Status updated:       %s\n", minipool.Status.StatusTime.Format(TimeFormat))
            fmt.Printf("Node fee:             %f%%\n", minipool.Node.Fee * 100)
            fmt.Printf("Node deposit:         %.6f ETH\n", math.RoundDown(eth.WeiToEth(minipool.Node.DepositBalance), 6))

            // RP ETH deposit details - prelaunch & staking minipools
            if minipool.Status.Status == types.Prelaunch || minipool.Status.Status == types.Staking {
                if minipool.User.DepositAssigned {
            fmt.Printf("RP ETH assigned:      %s\n", minipool.User.DepositAssignedTime.Format(TimeFormat))
            fmt.Printf("RP deposit:           %.6f ETH\n", math.RoundDown(eth.WeiToEth(minipool.User.DepositBalance), 6))
                } else {
            fmt.Printf("RP ETH assigned:      no\n")
                }
            }

            // Validator details - staking minipools
            if minipool.Status.Status == types.Staking {
            fmt.Printf("Validator pubkey:     %s\n", hex.AddPrefix(minipool.ValidatorPubkey.Hex()))
                if minipool.Validator.Exists {
                    if minipool.Validator.Active {
            fmt.Printf("Validator active:     yes\n")
                    } else {
            fmt.Printf("Validator active:     no\n")
                    }
            fmt.Printf("Validator balance:    %.6f ETH\n", math.RoundDown(eth.WeiToEth(minipool.Validator.Balance), 6))
            fmt.Printf("Expected rewards:     %.6f ETH\n", math.RoundDown(eth.WeiToEth(minipool.Validator.NodeBalance), 6))
                } else {
            fmt.Printf("Validator seen:       no\n")
                }
            }

            // Withdrawal details - withdrawable minipools
            if minipool.Status.Status == types.Withdrawable {
            fmt.Printf("Final balance:        %.6f ETH\n", math.RoundDown(eth.WeiToEth(minipool.Staking.EndBalance), 6))
                if minipool.WithdrawalAvailable {
            fmt.Printf("Withdrawal available: yes\n")
                } else {
            fmt.Printf("Withdrawal available: in %d blocks\n", minipool.WithdrawalAvailableInBlocks)
                }
            }

            fmt.Printf("\n")
        }

        fmt.Println("")
    }

    // Print actionable minipool details
    if len(refundableMinipools) > 0 {
        fmt.Printf("%d minipools have refunds available:\n", len(refundableMinipools))
        for _, minipool := range refundableMinipools {
            fmt.Printf("- %s (%.6f ETH to claim)\n", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(minipool.Node.RefundBalance), 6))
        }
        fmt.Println("")
    }
    if len(withdrawableMinipools) > 0 {
        fmt.Printf("%d minipools are ready for withdrawal:\n", len(withdrawableMinipools))
        for _, minipool := range withdrawableMinipools {
            fmt.Printf("- %s (%.6f nETH to claim)\n", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(minipool.Balances.NETH), 6))
        }
        fmt.Println("")
    }
    if len(closeableMinipools) > 0 {
        fmt.Printf("%d dissolved minipools can be closed:\n", len(closeableMinipools))
        for _, minipool := range closeableMinipools {
            fmt.Printf("- %s (%.6f ETH to claim)\n", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(minipool.Node.DepositBalance), 6))
        }
        fmt.Println("")
    }

    // Return
    return nil

}

