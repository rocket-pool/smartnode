package network

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)


func getStats(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get network stats
    response, err := rp.NetworkStats()
    if err != nil {
        return err
    }
    activeMinipools :=  response.InitializedMinipoolCount +
                        response.PrelaunchMinipoolCount +
                        response.StakingMinipoolCount +
                        response.WithdrawableMinipoolCount +
                        response.DissolvedMinipoolCount

    // Print & return
    fmt.Println("========== General Stats ==========")
    fmt.Printf("Total Value Locked:      %f ETH\n", response.TotalValueLocked)
    fmt.Printf("Staking Pool Balance:    %f ETH\n", response.DepositPoolBalance)
    fmt.Printf("Minipool Queue Demand:   %f ETH\n", response.MinipoolCapacity)
    fmt.Printf("Staking Pool ETH Used:   %f%%\n\n", response.StakerUtilization * 100)

    fmt.Println("============== Nodes ==============")
    fmt.Printf("Current Commission Rate: %f%%\n", response.NodeFee * 100)
    fmt.Printf("Node Count:              %d\n", response.NodeCount)
    fmt.Printf("Active Minipools:        %d\n", activeMinipools)
    fmt.Printf("    Initialized:         %d\n", response.InitializedMinipoolCount)
    fmt.Printf("    Prelaunch:           %d\n", response.PrelaunchMinipoolCount)
    fmt.Printf("    Staking:             %d\n", response.StakingMinipoolCount)
    fmt.Printf("    Withdrawable:        %d\n", response.WithdrawableMinipoolCount)
    fmt.Printf("    Dissolved:           %d\n", response.DissolvedMinipoolCount)
    fmt.Printf("Inactive Minipools:      %d\n\n", response.FinalizedMinipoolCount)

    fmt.Println("============== Tokens =============")
    fmt.Printf("rETH Price (ETH / rETH): %f ETH\n", response.RethPrice)
    fmt.Printf("RPL Price (ETH / RPL):   %f ETH\n", response.RplPrice)
    fmt.Printf("Total RPL staked:        %f RPL\n", response.TotalRplStaked)
    fmt.Printf("Effective RPL staked:    %f RPL\n", response.EffectiveRplStaked)

    return nil

}

