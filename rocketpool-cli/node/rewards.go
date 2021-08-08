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

    defaultTime := time.Time{}
    if rewards.NodeRegistrationTime == defaultTime {
        fmt.Printf("This node is not currently registered (could not find a registration event for this node).\n")
        return nil;
    }

    nextRewardsTime := rewards.LastCheckpoint.Add(rewards.RewardsInterval)
    nextRewardsTimeString := cliutils.GetDateTimeString(uint64(nextRewardsTime.Unix()))
    timeToCheckpointString := nextRewardsTime.Sub(time.Now()).Round(time.Second).String()
    timeSinceRegistration := time.Now().Sub(rewards.NodeRegistrationTime)
    docsUrl := "https://docs.rocketpool.net/guides/node/rewards.html#claiming-rpl-rewards"

    rplApy := rewards.CumulativeRewards / timeSinceRegistration.Hours() * (24*365) // Assume 365 days in a year, 24 hours per day
    rplTrustedApy := rewards.CumulativeTrustedRewards / timeSinceRegistration.Hours() * (24*365)

    fmt.Printf("The current rewards cycle started on %s.\n", cliutils.GetDateTimeString(uint64(rewards.LastCheckpoint.Unix())))
    fmt.Printf("It will end on %s (%s from now).\n\n", nextRewardsTimeString, timeToCheckpointString)
    
    fmt.Printf("Your estimated RPL staking rewards for this cycle: %f RPL (this may change based on network activity).\n", rewards.EstimatedRewards)
    fmt.Printf("Your node has received %f RPL staking rewards in total.\n", rewards.CumulativeRewards)
    fmt.Printf("Based on your current effective stake, this is approximately %f APY.", rplApy)

    if rewards.Trusted {
        fmt.Println()
        fmt.Printf("You will receive an estimated %f RPL in rewards for Oracle DAO duties (this may change based on network activity).\n", rewards.EstimatedTrustedRewards)
        fmt.Printf("Your node has received %f RPL Oracle DAO rewards in total.\n", rewards.CumulativeTrustedRewards)
        fmt.Printf("Based on your current effective stake, this is approximately %f APY.", rplTrustedApy)
    }

    fmt.Println()
    fmt.Println("These rewards will be claimed automatically when the checkpoint ends, unless you have disabled auto-claims.")
    fmt.Printf("Refer to the Claiming Node Operator Rewards guide at %s for more information.", docsUrl)

    // Return
    return nil

}

