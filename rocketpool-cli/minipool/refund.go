package minipool

import (
    "fmt"

    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func refundMinipools(c *cli.Context) error {

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Get minipool statuses
    status, err := rp.MinipoolStatus()
    if err != nil {
        return err
    }

    // Get refundable minipools
    refundableMinipools := []api.MinipoolDetails{}
    for _, minipool := range status.Minipools {
        if minipool.RefundAvailable {
            refundableMinipools = append(refundableMinipools, minipool)
        }
    }

    // Check for refundable minipools
    if len(refundableMinipools) == 0 {
        fmt.Println("No minipools have refunds available.")
        return nil
    }

    // Prompt for minipool selection
    options := make([]string, len(refundableMinipools) + 1)
    options[0] = "All available minipools"
    for mi, minipool := range refundableMinipools {
        options[mi + 1] = fmt.Sprintf("%s (%.2f ETH to claim)", minipool.Address.Hex(), eth.WeiToEth(minipool.Node.RefundBalance))
    }
    selected, _ := cliutils.Select("Please select a minipool to refund ETH from:", options)

    // Get selected minipools
    var selectedMinipools []api.MinipoolDetails
    if selected == 0 {
        selectedMinipools = refundableMinipools
    } else {
        selectedMinipools = []api.MinipoolDetails{refundableMinipools[selected - 1]}
    }

    // Refund minipools
    for _, minipool := range selectedMinipools {
        if _, err := rp.RefundMinipool(minipool.Address); err != nil {
            fmt.Printf("Could not refund ETH from minipool %s: %s.\n", minipool.Address.Hex(), err)
        } else {
            fmt.Printf("Successfully refunded ETH from minipool %s.\n", minipool.Address.Hex())
        }
    }

    // Return
    return nil

}

