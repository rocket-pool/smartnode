package minipool

import (
    "fmt"

    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func closeMinipools(c *cli.Context) error {

    // Get services
    rp, err := services.GetRocketPoolClient(c)
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

    // Prompt for minipool selection
    options := make([]string, len(closableMinipools) + 1)
    options[0] = "All available minipools"
    for mi, minipool := range closableMinipools {
        options[mi + 1] = fmt.Sprintf("%s (%.2f ETH to claim)", minipool.Address.Hex(), eth.WeiToEth(minipool.Node.DepositBalance))
    }
    selected, _ := cliutils.Select("Please select a minipool to close:", options)

    // Get selected minipools
    var selectedMinipools []api.MinipoolDetails
    if selected == 0 {
        selectedMinipools = closableMinipools
    } else {
        selectedMinipools = []api.MinipoolDetails{closableMinipools[selected - 1]}
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

