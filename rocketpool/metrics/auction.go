package metrics

import (
    "math/big"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/rocketpool-go/auction"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/log"
)


// minipool metrics process
type auctionGauges struct {
    lotCount        prometheus.Gauge
    balances        *prometheus.GaugeVec
}


type auctionMetricsProcess struct {
    rp *rocketpool.RocketPool
    metrics auctionGauges
    logger log.ColorLogger
}


// Start minipool metrics process
func startAuctionMetricsProcess(c *cli.Context, interval time.Duration, logger log.ColorLogger) {

    logger.Printlnf("Enter startAuctionMetricsProcess")
    timer := time.NewTicker(interval)
    var p *auctionMetricsProcess
    var err error
    // put create process in a loop because it may fail initially
    for ; true; <- timer.C {
        p, err = newAuctionMetricsProcss(c, logger)
        if p != nil && err == nil {
            break;
        }
    }

    // Update metrics on interval
    for ; true; <- timer.C {
        err = p.updateMetrics()
        if err != nil {
            // print error here instead of exit
            logger.Printlnf("Error in updateMetrics: %w", err)
        }
    }
    logger.Printlnf("Exit startAuctionMetricsProcess")
}


// Create new minipoolMetricsProcss object
func newAuctionMetricsProcss(c *cli.Context, logger log.ColorLogger) (*auctionMetricsProcess, error) {

    // Get services
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    if err := services.RequireBeaconClientSynced(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Initialise metrics
    metrics := auctionGauges {
        lotCount:           promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:      "rocketpool",
            Subsystem:      "auction",
            Name:           "lot_count",
            Help:           "number of lots in auction Rocket Pool",
        }),
        balances:           promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:  "rocketpool",
                Subsystem:  "auction",
                Name:       "balances_rpl",
                Help:       "the total RPL balance of the auction contract",
            },
            []string{"category"},
        ),
    }

    p := &auctionMetricsProcess {
        rp: rp,
        metrics: metrics,
        logger: logger,
    }

    return p, nil
}


// Update minipool metrics
func (p *auctionMetricsProcess) updateMetrics() error {
    p.logger.Println("Enter auction updateMetrics")

    var lotCount uint64
    var totalBalance, allottedBalance, remainingBalance *big.Int

    // Sync
    var wg errgroup.Group

    // Get data
    wg.Go(func() error {
        var err error
        lotCount, err = auction.GetLotCount(p.rp, nil)
        return err
    })

    wg.Go(func() error {
        var err error
        totalBalance, err = auction.GetTotalRPLBalance(p.rp, nil)
        return err
    })

    wg.Go(func() error {
        var err error
        allottedBalance, err = auction.GetAllottedRPLBalance(p.rp, nil)
        return err
    })

    wg.Go(func() error {
        var err error
        remainingBalance, err = auction.GetRemainingRPLBalance(p.rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }

    p.metrics.lotCount.Set(float64(lotCount))
    p.metrics.balances.With(prometheus.Labels{"category":"TotalRPL"}).Set(eth.WeiToEth(totalBalance))
    p.metrics.balances.With(prometheus.Labels{"category":"AllottedRPL"}).Set(eth.WeiToEth(allottedBalance))
    p.metrics.balances.With(prometheus.Labels{"category":"RemainingRPL"}).Set(eth.WeiToEth(remainingBalance))

    return nil
}

