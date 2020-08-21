package minipool

import (
    "fmt"

    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func withdrawMinipools(c *cli.Context) error {

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Get minipool statuses
    status, err := rp.MinipoolStatus()
    if err != nil {
        return err
    }

    // Get withdrawable minipools
    withdrawableMinipools := []api.MinipoolDetails{}
    for _, minipool := range status.Minipools {
        if minipool.WithdrawalAvailable {
            withdrawableMinipools = append(withdrawableMinipools, minipool)
        }
    }

    // Check for withdrawable minipools
    if len(withdrawableMinipools) == 0 {
        fmt.Println("No minipools can be withdrawn from.")
        return nil
    }

    // Prompt for minipool selection
    options := make([]string, len(withdrawableMinipools) + 1)
    options[0] = "All available minipools"
    for mi, minipool := range withdrawableMinipools {
        options[mi + 1] = fmt.Sprintf("%s (%.2f nETH to claim)", minipool.Address.Hex(), eth.WeiToEth(minipool.Balances.NETH))
    }
    selected, _ := cliutils.Select("Please select a minipool to withdraw from:", options)

    // Get selected minipools
    var selectedMinipools []api.MinipoolDetails
    if selected == 0 {
        selectedMinipools = withdrawableMinipools
    } else {
        selectedMinipools = []api.MinipoolDetails{withdrawableMinipools[selected - 1]}
    }

    // Withdraw minipools
    for _, minipool := range selectedMinipools {
        if _, err := rp.CloseMinipool(minipool.Address); err != nil {
            fmt.Printf("Could not withdraw from minipool %s: %s.\n", minipool.Address.Hex(), err)
        } else {
            fmt.Printf("Successfully withdrew from minipool %s.\n", minipool.Address.Hex())
        }
    }

    // Return
    return nil

}

