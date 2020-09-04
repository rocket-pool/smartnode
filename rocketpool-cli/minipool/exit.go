package minipool

import (
    "fmt"

    "github.com/rocket-pool/rocketpool-go/types"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/types/api"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func exitMinipools(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get minipool statuses
    status, err := rp.MinipoolStatus()
    if err != nil {
        return err
    }

    // Get staking minipools
    stakingMinipools := []api.MinipoolDetails{}
    for _, minipool := range status.Minipools {
        if minipool.Status.Status == types.Staking {
            stakingMinipools = append(stakingMinipools, minipool)
        }
    }

    // Check for staking minipools
    if len(stakingMinipools) == 0 {
        fmt.Println("No minipools can be exited.")
        return nil
    }

    // Prompt for minipool selection
    options := make([]string, len(stakingMinipools) + 1)
    options[0] = "All available minipools"
    for mi, minipool := range stakingMinipools {
        options[mi + 1] = fmt.Sprintf("%s (staking since %s)", minipool.Address.Hex(), minipool.Status.StatusTime.Format(TimeFormat))
    }
    selected, _ := cliutils.Select("Please select a minipool to exit:", options)

    // Get selected minipools
    var selectedMinipools []api.MinipoolDetails
    if selected == 0 {
        selectedMinipools = stakingMinipools
    } else {
        selectedMinipools = []api.MinipoolDetails{stakingMinipools[selected - 1]}
    }

    // Prompt for confirmation
    if !cliutils.Confirm(fmt.Sprintf("Are you sure you want to exit %d minipool(s)? This action cannot be undone!", len(selectedMinipools))) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Exit minipools
    for _, minipool := range selectedMinipools {
        if _, err := rp.ExitMinipool(minipool.Address); err != nil {
            fmt.Printf("Could not exit minipool %s: %s.\n", minipool.Address.Hex(), err)
        } else {
            fmt.Printf("Successfully exited minipool %s.\n", minipool.Address.Hex())
        }
    }

    // Return
    return nil

}

