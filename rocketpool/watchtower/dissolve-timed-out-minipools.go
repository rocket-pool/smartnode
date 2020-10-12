package watchtower

import (
    "context"
    "fmt"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/rocket-pool/rocketpool-go/types"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/wallet"
    "github.com/rocket-pool/smartnode/shared/utils/log"
)


// Settings
var dissolveTimedOutMinipoolsInterval, _ = time.ParseDuration("5m")
const MinipoolStatusBatchSize = 50


// Dissolve timed out minipools task
type dissolveTimedOutMinipools struct {
    c *cli.Context
    log log.ColorLogger
    w *wallet.Wallet
    ec *ethclient.Client
    rp *rocketpool.RocketPool
}


// Create dissolve timed out minipools task
func newDissolveTimedOutMinipools(c *cli.Context, logger log.ColorLogger) (*dissolveTimedOutMinipools, error) {

    // Get services
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    ec, err := services.GetEthClient(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Return task
    return &dissolveTimedOutMinipools{
        c: c,
        log: logger,
        w: w,
        ec: ec,
        rp: rp,
    }, nil

}


// Start dissolve timed out minipools task
func (t *dissolveTimedOutMinipools) Start() {
    go (func() {
        for {
            if err := t.run(); err != nil {
                t.log.Println(err)
            }
            time.Sleep(dissolveTimedOutMinipoolsInterval)
        }
    })()
}


// Dissolve timed out minipools
func (t *dissolveTimedOutMinipools) run() error {

    // Wait for eth client to sync
    if err := services.WaitEthClientSynced(t.c, true); err != nil {
        return err
    }

    // Get node account
    nodeAccount, err := t.w.GetNodeAccount()
    if err != nil {
        return err
    }

    // Check node trusted status
    nodeTrusted, err := node.GetNodeTrusted(t.rp, nodeAccount.Address, nil)
    if err != nil {
        return err
    }
    if !nodeTrusted {
        return nil
    }

    // Log
    t.log.Println("Checking for timed out minipools to dissolve...")

    // Get timed out minipools
    minipools, err := t.getTimedOutMinipools()
    if err != nil {
        return err
    }
    if len(minipools) == 0 {
        return nil
    }

    // Log
    t.log.Printlnf("%d minipools have timed out and will be dissolved...", len(minipools))

    // Dissolve minipools
    for _, mp := range minipools {
        if err := t.dissolveMinipool(mp); err != nil {
            t.log.Println(fmt.Errorf("Could not dissolve minipool %s: %w", mp.Address.Hex(), err))
        }
    }

    // Return
    return nil

}


// Get timed out minipools
func (t *dissolveTimedOutMinipools) getTimedOutMinipools() ([]*minipool.Minipool, error) {

    // Data
    var wg1 errgroup.Group
    var addresses []common.Address
    var currentBlock uint64
    var launchTimeout uint64

    // Get minipool addresses
    wg1.Go(func() error {
        var err error
        addresses, err = minipool.GetMinipoolAddresses(t.rp, nil)
        return err
    })

    // Get current block
    wg1.Go(func() error {
        header, err := t.ec.HeaderByNumber(context.Background(), nil)
        if err == nil {
            currentBlock = header.Number.Uint64()
        }
        return err
    })

    // Get launch timeout
    wg1.Go(func() error {
        var err error
        launchTimeout, err = settings.GetMinipoolLaunchTimeout(t.rp, nil)
        return err
    })

    // Wait for data
    if err := wg1.Wait(); err != nil {
        return []*minipool.Minipool{}, err
    }

    // Create minipool contracts
    minipools := make([]*minipool.Minipool, len(addresses))
    for mi, address := range addresses {
        mp, err := minipool.NewMinipool(t.rp, address)
        if err != nil {
            return []*minipool.Minipool{}, err
        }
        minipools[mi] = mp
    }

    // Load minipool statuses in batches
    statuses := make([]minipool.StatusDetails, len(minipools))
    for bsi := 0; bsi < len(addresses); bsi += MinipoolStatusBatchSize {

        // Get batch start & end index
        msi := bsi
        mei := bsi + MinipoolStatusBatchSize
        if mei > len(addresses) { mei = len(addresses) }

        // Log
        t.log.Printlnf("Checking minipools %d - %d of %d for timed out status...", msi + 1, mei, len(addresses))

        // Load statuses
        var wg errgroup.Group
        for mi := msi; mi < mei; mi++ {
            mi := mi
            wg.Go(func() error {
                mp := minipools[mi]
                status, err := mp.GetStatusDetails(nil)
                if err == nil { statuses[mi] = status }
                return err
            })
        }
        if err := wg.Wait(); err != nil {
            return []*minipool.Minipool{}, err
        }

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
func (t *dissolveTimedOutMinipools) dissolveMinipool(mp *minipool.Minipool) error {

    // Log
    t.log.Printlnf("Dissolving minipool %s...", mp.Address.Hex())

    // Get transactor
    opts, err := t.w.GetNodeAccountTransactor()
    if err != nil {
        return err
    }

    // Dissolve
    if _, err := mp.Dissolve(opts); err != nil {
        return err
    }

    // Log
    t.log.Printlnf("Successfully dissolved minipool %s.", mp.Address.Hex())

    // Return
    return nil

}

