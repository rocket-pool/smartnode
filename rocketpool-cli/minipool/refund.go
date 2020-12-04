package minipool

import (
    "bytes"
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/types/api"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
    "github.com/rocket-pool/smartnode/shared/utils/math"
)


func refundMinipools(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
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

    // Get selected minipools
    var selectedMinipools []api.MinipoolDetails
    if c.String("minipool") == "" {

        // Prompt for minipool selection
        options := make([]string, len(refundableMinipools) + 1)
        options[0] = "All available minipools"
        for mi, minipool := range refundableMinipools {
            options[mi + 1] = fmt.Sprintf("%s (%.6f ETH to claim)", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(minipool.Node.RefundBalance), 6))
        }
        selected, _ := cliutils.Select("Please select a minipool to refund ETH from:", options)

        // Get minipools
        if selected == 0 {
            selectedMinipools = refundableMinipools
        } else {
            selectedMinipools = []api.MinipoolDetails{refundableMinipools[selected - 1]}
        }

    } else {

        // Get matching minipools
        if c.String("minipool") == "all" {
            selectedMinipools = refundableMinipools
        } else {
            selectedAddress := common.HexToAddress(c.String("minipool"))
            for _, minipool := range refundableMinipools {
                if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
                    selectedMinipools = []api.MinipoolDetails{minipool}
                    break
                }
            }
            if selectedMinipools == nil {
                return fmt.Errorf("The minipool %s is not available for refund.", selectedAddress.Hex())
            }
        }

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

