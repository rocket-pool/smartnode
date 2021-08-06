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

    nextRewardsTime := time.Now().Add(rewards.TimeToCheckpoint)
    nextRewardsTimeString := cliutils.GetDateTimeString(uint64(nextRewardsTime.Unix()))
    timeToCheckpointString := rewards.TimeToCheckpoint.String()

    fmt.Printf("Next rewards checkpoint is schedule for %s (%s from now).\n", nextRewardsTimeString, timeToCheckpointString)
    fmt.Printf("You will receive an estimated %f RPL in rewards from your RPL stake (this may change based on network activity).\n", rewards.EstimatedRewards)
    fmt.Printf("Your node has received %f RPL staking rewards in total.\n", rewards.CumulativeRewards)

    if rewards.Trusted {
        fmt.Println()
        fmt.Printf("You will receive an estimated %f RPL in rewards for Oracle DAO duties (this may change based on network activity).\n", rewards.EstimatedTrustedRewards)
        fmt.Printf("Your node has received %f RPL Oracle DAO rewards in total.\n", rewards.CumulativeTrustedRewards)
    }

    // Return
    return nil

}

