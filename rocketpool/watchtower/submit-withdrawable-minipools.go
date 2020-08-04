package watchtower

import (
    "log"
    "time"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/accounts"
)


// Settings
var submitWithdrawableMinipoolsInterval, _ = time.ParseDuration("1m")


// Start submit withdrawable minipools task
func startSubmitWithdrawableMinipools(c *cli.Context) error {

    // Get services
    if err := services.WaitNodeRegistered(c, true); err != nil { return err }
    am, err := services.GetAccountManager(c)
    if err != nil { return err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return err }

    // Submit withdrawable minipools at interval
    go (func() {
        for {
            if err := submitWithdrawableMinipools(c, am, rp); err != nil {
                log.Println(err)
            }
            time.Sleep(submitWithdrawableMinipoolsInterval)
        }
    })()

    // Return
    return nil

}


// Submit withdrawable minipools
func submitWithdrawableMinipools(c *cli.Context, am *accounts.AccountManager, rp *rocketpool.RocketPool) error {

    // Submit withdrawable minipools
    // TODO: implement

    // Return
    return nil

}

