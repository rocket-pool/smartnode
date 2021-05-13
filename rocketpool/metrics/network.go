package metrics

import (
    "math/big"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"
    "go.uber.org/multierr"

    "github.com/rocket-pool/rocketpool-go/deposit"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/network"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/types"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    apiNetwork "github.com/rocket-pool/smartnode/rocketpool/api/network"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/beacon"
    "github.com/rocket-pool/smartnode/shared/utils/log"
)


type networkGauges struct {
    nodeCount                        prometheus.Gauge
    minipoolCount                    prometheus.Gauge
    minipoolQueue                    *prometheus.GaugeVec
    networkFees                      *prometheus.GaugeVec
    rplPriceBlock                    prometheus.Gauge
    rplPrice                         prometheus.Gauge
    networkBlock                     prometheus.Gauge
    networkBalances                  *prometheus.GaugeVec
}


// network metrics process
type networkMetricsProcess struct {
    rp *rocketpool.RocketPool
    bc beacon.Client
    metrics networkGauges
    logger log.ColorLogger
}


type networkBalances struct {
    Block                   uint64
    TotalETH                *big.Int
    StakingETH              *big.Int
    TotalRETH               *big.Int
    DepositBalance          *big.Int
    DepositExcessBalance    *big.Int
    TotalRplStake           *big.Int
    TotalEffectiveRplStake  *big.Int
}


// Start network metrics process
func startNetworkMetricsProcess(c *cli.Context, interval time.Duration, logger log.ColorLogger) {

    logger.Printlnf("Enter startNetworkMetricsProcess")
    timer := time.NewTicker(interval)
    var p *networkMetricsProcess
    var err error
    // put create process in a loop because it may fail initially
    for ; true; <- timer.C {
        p, err = newNetworkMetricsProcess(c, logger)
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
    logger.Printlnf("Exit startNetworkMetricsProcess")
}


// Create new networkMetricsProcess object
func newNetworkMetricsProcess(c *cli.Context, logger log.ColorLogger) (*networkMetricsProcess, error) {

    logger.Printlnf("Enter newNetworkMetricsProcess")
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    if err := services.RequireBeaconClientSynced(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }
    bc, err := services.GetBeaconClient(c)
    if err != nil { return nil, err }

    // Initialise metrics
    metrics := networkGauges {
        nodeCount:          promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:      "rocketpool",
            Subsystem:      "node",
            Name:           "total_count",
            Help:           "total number of nodes in Rocket Pool",
        }),
        minipoolCount:      promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:      "rocketpool",
            Subsystem:      "minipool",
            Name:           "total_count",
            Help:           "total number of minipools in Rocket Pool",
        }),
        minipoolQueue:      promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:  "rocketpool",
                Subsystem:  "minipool",
                Name:       "queue_count",
                Help:       "number of minipools in queue for assignment",
            },
            []string{"depositType"},
        ),
        networkFees:        promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:  "rocketpool",
                Subsystem:  "network",
                Name:       "fee_rate",
                Help:       "network fees as rate of amount staked",
            },
            []string{"range"},
        ),
        rplPriceBlock:      promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:      "rocketpool",
            Subsystem:      "network",
            Name:           "rpl_price_updated_block",
            Help:           "block of current submitted RPL price",
        }),
        rplPrice:           promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:      "rocketpool",
            Subsystem:      "network",
            Name:           "rpl_price_eth",
            Help:           "RPL price in ETH",
        }),
        networkBlock:       promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:      "rocketpool",
            Subsystem:      "network",
            Name:           "balance_updated_block",
            Help:           "block of current submitted balances",
        }),
        networkBalances:    promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:  "rocketpool",
                Subsystem:  "network",
                Name:       "balance_eth",
                Help:       "network balances and supplies in given category",
            },
            []string{"category"},
        ),
    }

    p := &networkMetricsProcess {
        rp: rp,
        bc: bc,
        //account: account,
        metrics: metrics,
        logger: logger,
    }

    logger.Printlnf("Exit newNetworkMetricsProcess")
    return p, nil
}


// Update network metrics
func (p *networkMetricsProcess) updateMetrics() error {
    p.logger.Printlnf("Enter network updateMetrics")

    err1 := p.updateCounts()
    err4 := p.updateNetwork()
    err5 := p.updateMinipoolQueue()
    err := multierr.Combine(err1, err4, err5)

    p.logger.Printlnf("Exit network updateMetrics with %d errors", len(multierr.Errors(err)))
    return err
}


func (p *networkMetricsProcess) updateCounts() error {

    nodeCount, err := node.GetNodeCount(p.rp, nil)
    if err != nil { return err }
    p.metrics.nodeCount.Set(float64(nodeCount))

    minipoolCount, err := minipool.GetMinipoolCount(p.rp, nil)
    if err != nil { return err }
    p.metrics.minipoolCount.Set(float64(minipoolCount))

    return nil
}


func (p *networkMetricsProcess) updateNetwork() error {

    nodeFees, err := apiNetwork.GetNodeFee(p.rp)
    if err != nil { return err }

    p.metrics.networkFees.With(prometheus.Labels{"range":"current"}).Set(nodeFees.NodeFee)
    p.metrics.networkFees.With(prometheus.Labels{"range":"min"}).Set(nodeFees.MinNodeFee)
    p.metrics.networkFees.With(prometheus.Labels{"range":"target"}).Set(nodeFees.TargetNodeFee)
    p.metrics.networkFees.With(prometheus.Labels{"range":"max"}).Set(nodeFees.MaxNodeFee)

    rplPrice, err := apiNetwork.GetRplPrice(p.rp)
    if err != nil { return err }

    p.metrics.rplPriceBlock.Set(float64(rplPrice.RplPriceBlock))
    p.metrics.rplPrice.Set(eth.WeiToEth(rplPrice.RplPrice))

    balances, err := getNetworkBalances(p.rp)
    if err != nil { return err }

    p.metrics.networkBlock.Set(float64(balances.Block))
    p.metrics.networkBalances.With(prometheus.Labels{"category":"TotalETH"}).Set(eth.WeiToEth(balances.TotalETH))
    p.metrics.networkBalances.With(prometheus.Labels{"category":"StakingETH"}).Set(eth.WeiToEth(balances.StakingETH))
    p.metrics.networkBalances.With(prometheus.Labels{"category":"TotalRETH"}).Set(eth.WeiToEth(balances.TotalRETH))
    p.metrics.networkBalances.With(prometheus.Labels{"category":"Deposit"}).Set(eth.WeiToEth(balances.DepositBalance))
    p.metrics.networkBalances.With(prometheus.Labels{"category":"DepositExcess"}).Set(eth.WeiToEth(balances.DepositExcessBalance))
    p.metrics.networkBalances.With(prometheus.Labels{"category":"TotalRPL"}).Set(eth.WeiToEth(balances.TotalRplStake))
    p.metrics.networkBalances.With(prometheus.Labels{"category":"TotalEffectiveRPL"}).Set(eth.WeiToEth(balances.TotalEffectiveRplStake))

    return nil
}


func getNetworkBalances(rp *rocketpool.RocketPool) (*networkBalances, error) {
    stuff := networkBalances{}

    // Sync
    var wg errgroup.Group

    // Get data
    wg.Go(func() error {
        block, err := network.GetBalancesBlock(rp, nil)
        if err == nil {
            stuff.Block = block
        }
        return err
    })
    wg.Go(func() error {
        totalETH, err := network.GetTotalETHBalance(rp, nil)
        if err == nil {
            stuff.TotalETH = totalETH
        }
        return err
    })
    wg.Go(func() error {
        stakingETH, err := network.GetStakingETHBalance(rp, nil)
        if err == nil {
            stuff.StakingETH = stakingETH
        }
        return err
    })
    wg.Go(func() error {
        totalRETH, err := network.GetTotalRETHSupply(rp, nil)
        if err == nil {
            stuff.TotalRETH = totalRETH
        }
        return err
    })
    wg.Go(func() error {
        depositBalance, err := deposit.GetBalance(rp, nil)
        if err == nil {
            stuff.DepositBalance = depositBalance
        }
        return err
    })
    wg.Go(func() error {
        depositExcessBalance, err := deposit.GetExcessBalance(rp, nil)
        if err == nil {
            stuff.DepositExcessBalance = depositExcessBalance
        }
        return err
    })
    wg.Go(func() error {
        totalRplStake, err := node.GetTotalRPLStake(rp, nil)
        if err == nil {
            stuff.TotalRplStake = totalRplStake
        }
        return err
    })
    wg.Go(func() error {
        totalEffectiveRplStake, err := node.GetTotalEffectiveRPLStake(rp, nil)
        if err == nil {
            stuff.TotalEffectiveRplStake = totalEffectiveRplStake
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Return response
    return &stuff, nil
}


func (p *networkMetricsProcess) updateMinipoolQueue() error {
    var wg errgroup.Group
    var fullQueueLength, halfQueueLength, emptyQueueLength uint64

    // Get data
    wg.Go(func() error {
        response, err := minipool.GetQueueLength(p.rp, types.Full, nil)
        if err == nil {
            fullQueueLength = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := minipool.GetQueueLength(p.rp, types.Half, nil)
        if err == nil {
            halfQueueLength = response
        }
        return err
    })
    wg.Go(func() error {
        response, err := minipool.GetQueueLength(p.rp, types.Empty, nil)
        if err == nil {
            emptyQueueLength = response
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }
    p.metrics.minipoolQueue.With(prometheus.Labels{"depositType":"Full"}).Set(float64(fullQueueLength))
    p.metrics.minipoolQueue.With(prometheus.Labels{"depositType":"Half"}).Set(float64(halfQueueLength))
    p.metrics.minipoolQueue.With(prometheus.Labels{"depositType":"Empty"}).Set(float64(emptyQueueLength))

    return nil
}

