package minipool

import (
    "fmt"

    "github.com/rocket-pool/rocketpool-go/types"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/types/api"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func dissolveMinipools(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get minipool statuses
    status, err := rp.MinipoolStatus()
    if err != nil {
        return err
    }

    // Get initialized minipools
    initializedMinipools := []api.MinipoolDetails{}
    for _, minipool := range status.Minipools {
        if minipool.Status.Status == types.Initialized {
            initializedMinipools = append(initializedMinipools, minipool)
        }
    }

    // Check for initialized minipools
    if len(initializedMinipools) == 0 {
        fmt.Println("No minipools can be dissolved.")
        return nil
    }

    // Prompt for minipool selection
    options := make([]string, len(initializedMinipools) + 1)
    options[0] = "All available minipools"
    for mi, minipool := range initializedMinipools {
        options[mi + 1] = fmt.Sprintf("%s (%.2f ETH deposited)", minipool.Address.Hex(), eth.WeiToEth(minipool.Node.DepositBalance))
    }
    selected, _ := cliutils.Select("Please select a minipool to dissolve:", options)

    // Get selected minipools
    var selectedMinipools []api.MinipoolDetails
    if selected == 0 {
        selectedMinipools = initializedMinipools
    } else {
        selectedMinipools = []api.MinipoolDetails{initializedMinipools[selected - 1]}
    }

    // Prompt for confirmation
    if !cliutils.Confirm(fmt.Sprintf("Are you sure you want to dissolve %d minipool(s)? This action cannot be undone!", len(selectedMinipools))) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Dissolve and close minipools
    for _, minipool := range selectedMinipools {
        if _, err := rp.DissolveMinipool(minipool.Address); err != nil {
            fmt.Printf("Could not dissolve minipool %s: %s.\n", minipool.Address.Hex(), err)
            continue
        } else {
            fmt.Printf("Successfully dissolved minipool %s.\n", minipool.Address.Hex())
        }
        if _, err := rp.CloseMinipool(minipool.Address); err != nil {
            fmt.Printf("Could not close minipool %s: %s.\n", minipool.Address.Hex(), err)
        } else {
            fmt.Printf("Successfully closed minipool %s.\n", minipool.Address.Hex())
        }
    }

    // Return
    return nil

}

