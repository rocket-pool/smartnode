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


func withdrawMinipools(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
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

    // Get selected minipools
    var selectedMinipools []api.MinipoolDetails
    if c.String("minipool") == "" {

        // Prompt for minipool selection
        options := make([]string, len(withdrawableMinipools) + 1)
        options[0] = "All available minipools"
        for mi, minipool := range withdrawableMinipools {
            options[mi + 1] = fmt.Sprintf("%s (%.6f nETH to claim)", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(minipool.Balances.NETH), 6))
        }
        selected, _ := cliutils.Select("Please select a minipool to withdraw from:", options)

        // Get minipools
        if selected == 0 {
            selectedMinipools = withdrawableMinipools
        } else {
            selectedMinipools = []api.MinipoolDetails{withdrawableMinipools[selected - 1]}
        }

    } else {

        // Get matching minipools
        if c.String("minipool") == "all" {
            selectedMinipools = withdrawableMinipools
        } else {
            selectedAddress := common.HexToAddress(c.String("minipool"))
            for _, minipool := range withdrawableMinipools {
                if bytes.Equal(minipool.Address.Bytes(), selectedAddress.Bytes()) {
                    selectedMinipools = []api.MinipoolDetails{minipool}
                    break
                }
            }
            if selectedMinipools == nil {
                return fmt.Errorf("The minipool %s is not available for withdrawal.", selectedAddress.Hex())
            }
        }

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

