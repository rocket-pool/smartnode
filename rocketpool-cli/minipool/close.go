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


func closeMinipools(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get minipool statuses
    status, err := rp.MinipoolStatus()
    if err != nil {
        return err
    }

    // Get closable minipools
    closableMinipools := []api.MinipoolDetails{}
    for _, minipool := range status.Minipools {
        if minipool.CloseAvailable {
            closableMinipools = append(closableMinipools, minipool)
        }
    }

    // Check for closable minipools
    if len(closableMinipools) == 0 {
        fmt.Println("No minipools can be closed.")
        return nil
    }

    // Get selected minipools
    var selectedMinipools []api.MinipoolDetails
    if c.String("minipool") == "" {

        // Prompt for minipool selection
        options := make([]string, len(closableMinipools) + 1)
        options[0] = "All available minipools"
        for mi, minipool := range closableMinipools {
            options[mi + 1] = fmt.Sprintf("%s (%.6f ETH to claim)", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(minipool.Node.DepositBalance), 6))
        }
        selected, _ := cliutils.Select("Please select a minipool to close:", options)

        // Get minipools
        if selected == 0 {
            selectedMinipools = closableMinipools
        } else {
            selectedMinipools = []api.MinipoolDetails{closableMinipools[selected - 1]}
        }

    } else {

        // Get matching minipools
        if c.String("minipool") == "all" {
            selectedMinipools = closableMinipools
        } else {
            selectedAddress := common.HexToAddress(c.String("minipool"))
            for _, minipool := range closableMinipools {
                if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
                    selectedMinipools = []api.MinipoolDetails{minipool}
                    break
                }
            }
            if selectedMinipools == nil {
                return fmt.Errorf("The minipool %s is not available for closing.", selectedAddress.Hex())
            }
        }

    }

    // Close minipools
    for _, minipool := range selectedMinipools {
        if _, err := rp.CloseMinipool(minipool.Address); err != nil {
            fmt.Printf("Could not close minipool %s: %s.\n", minipool.Address.Hex(), err)
        } else {
            fmt.Printf("Successfully closed minipool %s.\n", minipool.Address.Hex())
        }
    }

    // Return
    return nil

}

