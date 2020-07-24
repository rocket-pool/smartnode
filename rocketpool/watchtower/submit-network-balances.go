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
var submitNetworkBalancesInterval, _ = time.ParseDuration("384s") // 1 epoch


// Start submit network balances task
func startSubmitNetworkBalances(c *cli.Context) error {

    // Get services
    if err := services.WaitNodeRegistered(c, true); err != nil { return err }
    am, err := services.GetAccountManager(c)
    if err != nil { return err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return err }

    // Submit network balances at interval
    go (func() {
        for {
            if err := submitNetworkBalances(c, am, rp); err != nil {
                log.Println(err)
            }
            time.Sleep(submitNetworkBalancesInterval)
        }
    })()

    // Return
    return nil

}


// Submit network balances
func submitNetworkBalances(c *cli.Context, am *accounts.AccountManager, rp *rocketpool.RocketPool) error {

    // TODO: implement
    log.Println("Submitting network balances not implemented...")
    return nil

}

