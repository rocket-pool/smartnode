package minipool

import (
    "fmt"

    "github.com/rocket-pool/rocketpool-go/types"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/types/api"
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

    // Print & return
    for _, statusName := range types.MinipoolStatuses {
        minipools, ok := statusMinipools[statusName]
        if !ok { continue }
        fmt.Printf("%d %s minipool(s):\n", len(minipools), statusName)
        fmt.Println("")
        for _, minipool := range minipools {
            fmt.Printf("-----------------\n")
            fmt.Printf("\n")
            fmt.Printf("Address:           %s\n", minipool.Address.Hex())
            fmt.Printf("Status updated:    %s\n", minipool.Status.StatusTime.Format(TimeFormat))
            fmt.Printf("Node fee:          %f%%\n", minipool.Node.Fee * 100)
            fmt.Printf("Node deposit:      %.2f ETH\n", eth.WeiToEth(minipool.Node.DepositBalance))
            if minipool.Status.Status == types.Prelaunch || minipool.Status.Status == types.Staking {
            if minipool.User.DepositAssigned {
            fmt.Printf("RP ETH assigned:   %s\n", minipool.User.DepositAssignedTime.Format(TimeFormat))
            fmt.Printf("RP deposit:        %.2f ETH\n", eth.WeiToEth(minipool.User.DepositBalance))
            } else {
            fmt.Printf("RP ETH assigned:   no\n")
            }
            }
            if minipool.Status.Status == types.Staking {
            fmt.Printf("Validator pubkey:  %s\n", minipool.ValidatorPubkey.Hex())
            if minipool.Validator.Exists {
            if minipool.Validator.Active {
            fmt.Printf("Validator active:  yes\n")
            } else {
            fmt.Printf("Validator active:  no\n")
            }
            fmt.Printf("Validator balance: %.2f ETH\n", eth.WeiToEth(minipool.Validator.Balance))
            fmt.Printf("Expected rewards:  %.2f ETH\n", eth.WeiToEth(minipool.Validator.NodeBalance))
            } else {
            fmt.Printf("Validator seen:    no\n")
            }
            }
            if minipool.Status.Status == types.Withdrawable {
            fmt.Printf("Final balance:     %.2f ETH\n", eth.WeiToEth(minipool.Staking.EndBalance))
            }
            fmt.Printf("\n")
        }
        fmt.Println("")
    }
    if len(refundableMinipools) > 0 {
        fmt.Printf("%d minipools have refunds available:\n", len(refundableMinipools))
        for _, minipool := range refundableMinipools {
            fmt.Printf("- %s (%.2f ETH to claim)\n", minipool.Address.Hex(), eth.WeiToEth(minipool.Node.RefundBalance))
        }
        fmt.Println("")
    }
    if len(withdrawableMinipools) > 0 {
        fmt.Printf("%d minipools are ready for withdrawal:\n", len(withdrawableMinipools))
        for _, minipool := range withdrawableMinipools {
            fmt.Printf("- %s (%.2f nETH to claim)\n", minipool.Address.Hex(), eth.WeiToEth(minipool.Balances.NETH))
        }
        fmt.Println("")
    }
    if len(closeableMinipools) > 0 {
        fmt.Printf("%d dissolved minipools can be closed:\n", len(closeableMinipools))
        for _, minipool := range closeableMinipools {
            fmt.Printf("- %s (%.2f ETH to claim)\n", minipool.Address.Hex(), eth.WeiToEth(minipool.Node.DepositBalance))
        }
        fmt.Println("")
    }
    return nil

}

