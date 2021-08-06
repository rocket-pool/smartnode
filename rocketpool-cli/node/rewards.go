package node

import (
	"fmt"
	"time"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


func getRewards(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get node RPL rewards status
    rewards, err := rp.NodeRewards()
    if err != nil {
        return err
    }

    nextRewardsTime := rewards.LastCheckpoint.Add(rewards.RewardsInterval)
    nextRewardsTimeString := cliutils.GetDateTimeString(uint64(nextRewardsTime.Unix()))
    timeToCheckpointString := nextRewardsTime.Sub(time.Now()).Round(time.Second).String()

    docsUrl := "https://docs.rocketpool.net/guides/node/rewards.html#claiming-rpl-rewards"

    fmt.Printf("The current rewards cycle started on %s.\n", cliutils.GetDateTimeString(uint64(rewards.LastCheckpoint.Unix())))
    fmt.Printf("It will end on %s (%s from now).\n\n", nextRewardsTimeString, timeToCheckpointString)
    
    fmt.Printf("Your estimated RPL staking rewards for this cycle: %f RPL (this may change based on network activity).\n", rewards.EstimatedRewards)
    fmt.Printf("Your node has received %f RPL staking rewards in total.\n", rewards.CumulativeRewards)

    if rewards.Trusted {
        fmt.Println()
        fmt.Printf("You will receive an estimated %f RPL in rewards for Oracle DAO duties (this may change based on network activity).\n", rewards.EstimatedTrustedRewards)
        fmt.Printf("Your node has received %f RPL Oracle DAO rewards in total.\n", rewards.CumulativeTrustedRewards)
    }

    fmt.Println()
    fmt.Println("These rewards will be claimed automatically when the checkpoint ends, unless you have disabled auto-claims.")
    fmt.Printf("Refer to the Claiming Node Operator Rewards guide at %s for more information.", docsUrl)

    // Return
    return nil

}

