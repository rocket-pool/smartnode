package watchtower

import (
    "fmt"

    "github.com/rocket-pool/rocketpool-go/network"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/wallet"
    "github.com/rocket-pool/smartnode/shared/utils/log"
)


// Process withdrawals task
type processWithdrawals struct {
    c *cli.Context
    log log.ColorLogger
    w *wallet.Wallet
    rp *rocketpool.RocketPool
}


// Create process withdrawals task
func newProcessWithdrawals(c *cli.Context, logger log.ColorLogger) (*processWithdrawals, error) {

    // Get services
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Return task
    return &processWithdrawals{
        c: c,
        log: logger,
        w: w,
        rp: rp,
    }, nil

}


// Process withdrawals
func (t *processWithdrawals) run() error {

    // Wait for eth clients to sync
    if err := services.WaitEthClientSynced(t.c, true); err != nil {
        return err
    }
    if err := services.WaitBeaconClientSynced(t.c, true); err != nil {
        return err
    }

    // Get node account
    nodeAccount, err := t.w.GetNodeAccount()
    if err != nil {
        return err
    }

    // Data
    var wg errgroup.Group
    var nodeTrusted bool
    var processWithdrawalsEnabled bool

    // Get data
    wg.Go(func() error {
        var err error
        nodeTrusted, err = node.GetNodeTrusted(t.rp, nodeAccount.Address, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        processWithdrawalsEnabled, err = settings.GetProcessWithdrawalsEnabled(t.rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }

    // Check node trusted status & settings
    if !(nodeTrusted && processWithdrawalsEnabled) {
        return nil
    }

    t.log.Println("Checking for withdrawals to process...")

    // Get minipool withdrawable details
    withdrawalDetails, err := minipool.GetUnprocessedMinipools(t.rp, nil)
    if err != nil {
        return err
    }
    if len(withdrawalDetails) == 0 {
        return nil
    }

    t.log.Printlnf("%d minipools are unprocessed...", len(withdrawalDetails))

    // Submit minipools withdrawable status
    for _, withdrawal := range withdrawalDetails {
        // Logic: should only process withdraw when minipool.WithDraw() has been called
        if !withdrawal.Exists && !withdrawal.WithdrawalProcessed {
            // TODO TODO TODO TODO TODO TODO TODO TODO
            // _real_ ETH2 withdraw logic goes here
            // no clients have any implementation yet

            _, err := network.ProcessWithdrawal(t.rp, withdrawal.Pubkey, nil)
            if err != nil {
                t.log.Println(fmt.Errorf("Could not process withdrawal for minipool %s: %w", withdrawal.Pubkey.Hex(), err))
            }
        }
    }

    return nil

}

