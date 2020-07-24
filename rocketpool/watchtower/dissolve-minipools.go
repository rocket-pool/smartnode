package watchtower

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/rocket-pool/rocketpool-go/types"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/accounts"
)


// Settings
var checkMinipoolsInterval, _ = time.ParseDuration("1m")


// Start dissolve timed out minipools task
func startDissolveTimedOutMinipools(c *cli.Context) error {

    // Get services
    if err := services.WaitNodeRegistered(c, true); err != nil { return err }
    am, err := services.GetAccountManager(c)
    if err != nil { return err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return err }

    // Dissolve timed out minipools at interval
    go (func() {
        for {
            if err := dissolveTimedOutMinipools(c, am, rp); err != nil {
                log.Println(err)
            }
            time.Sleep(checkMinipoolsInterval)
        }
    })()

    // Return
    return nil

}


// Dissolve timed out minipools
func dissolveTimedOutMinipools(c *cli.Context, am *accounts.AccountManager, rp *rocketpool.RocketPool) error {

    // Wait for eth client to sync
    if err := services.WaitClientSynced(c, true); err != nil {
        return err
    }

    // Get node account
    nodeAccount, err := am.GetNodeAccount()
    if err != nil {
        return err
    }

    // Check node trusted status
    nodeTrusted, err := node.GetNodeTrusted(rp, nodeAccount.Address)
    if err != nil {
        return err
    }
    if !nodeTrusted {
        return nil
    }

    // Get timed out minipools
    minipools, err := getTimedOutMinipools(rp)
    if err != nil {
        return err
    }
    if len(minipools) == 0 {
        return nil
    }

    // Log
    log.Printf("%d minipools have timed out and will be dissolved...\n", len(minipools))

    // Dissolve minipools
    for _, mp := range minipools {
        if err := dissolveMinipool(am, mp); err != nil {
            log.Println(fmt.Errorf("Could not dissolve minipool %s: %w", mp.Address.Hex(), err))
        }
    }

    // Return
    return nil

}


// Get timed out minipools
func getTimedOutMinipools(rp *rocketpool.RocketPool) ([]*minipool.Minipool, error) {

    // Data
    var wg1 errgroup.Group
    var addresses []common.Address
    var currentBlock int64
    var launchTimeout int64

    // Get minipool addresses
    wg1.Go(func() error {
        var err error
        addresses, err = minipool.GetMinipoolAddresses(rp)
        return err
    })

    // Get current block
    wg1.Go(func() error {
        header, err := rp.Client.HeaderByNumber(context.Background(), nil)
        if err == nil {
            currentBlock = header.Number.Int64()
        }
        return err
    })

    // Get launch timeout
    wg1.Go(func() error {
        var err error
        launchTimeout, err = settings.GetMinipoolLaunchTimeout(rp)
        return err
    })

    // Wait for data
    if err := wg1.Wait(); err != nil {
        return []*minipool.Minipool{}, err
    }

    // Create minipool contracts
    minipools := make([]*minipool.Minipool, len(addresses))
    for mi, address := range addresses {
        mp, err := minipool.NewMinipool(rp, address)
        if err != nil {
            return []*minipool.Minipool{}, err
        }
        minipools[mi] = mp
    }

    // Data
    var wg2 errgroup.Group
    statuses := make([]minipool.StatusDetails, len(minipools))

    // Load minipool statuses
    for mi, mp := range minipools {
        mi, mp := mi, mp
        wg2.Go(func() error {
            status, err := mp.GetStatusDetails()
            if err == nil { statuses[mi] = status }
            return err
        })
    }

    // Wait for data
    if err := wg2.Wait(); err != nil {
        return []*minipool.Minipool{}, err
    }

    // Filter minipools by status
    timedOutMinipools := []*minipool.Minipool{}
    for mi, mp := range minipools {
        if statuses[mi].Status == types.Prelaunch && (currentBlock - statuses[mi].StatusBlock) >= launchTimeout {
            timedOutMinipools = append(timedOutMinipools, mp)
        }
    }

    // Return
    return timedOutMinipools, nil

}


// Dissolve a minipool
func dissolveMinipool(am *accounts.AccountManager, mp *minipool.Minipool) error {

    // Log
    log.Printf("Dissolving minipool %s...\n", mp.Address.Hex())

    // Get transactor
    opts, err := am.GetNodeAccountTransactor()
    if err != nil {
        return err
    }

    // Dissolve
    if _, err := mp.Dissolve(opts); err != nil {
        return err
    }

    // Log
    log.Printf("Successfully dissolved minipool %s.\n", mp.Address.Hex())

    // Return
    return nil

}

