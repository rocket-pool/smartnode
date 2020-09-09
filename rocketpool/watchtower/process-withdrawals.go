package watchtower

import (
    "log"
    "time"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/wallet"
)


// Settings
var processWithdrawalsInterval, _ = time.ParseDuration("5m")


// Process withdrawals task
type processWithdrawals struct {
    c *cli.Context
    w *wallet.Wallet
    rp *rocketpool.RocketPool
}


// Create process withdrawals task
func newProcessWithdrawals(c *cli.Context) (*processWithdrawals, error) {

    // Get services
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Return task
    return &processWithdrawals{
        c: c,
        w: w,
        rp: rp,
    }, nil

}


// Start process withdrawals task
func (t *processWithdrawals) Start() {
    go (func() {
        for {
            if err := t.run(); err != nil {
                log.Println(err)
            }
            time.Sleep(processWithdrawalsInterval)
        }
    })()
}


// Process withdrawals
func (t *processWithdrawals) run() error {

    // Process withdrawals
    // TODO: implement

    // Return
    return nil

}

