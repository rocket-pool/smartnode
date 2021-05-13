package metrics

import (
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/urfave/cli"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/beacon"
    "github.com/rocket-pool/smartnode/shared/utils/log"
)


type tokensGauges struct {
    tokenSupply                    *prometheus.GaugeVec
    something                      prometheus.Gauge
}


// tokens metrics process
type tokensMetricsProcess struct {
    rp *rocketpool.RocketPool
    bc beacon.Client
    metrics tokensGauges
    logger log.ColorLogger
}




// Start tokens metrics process
func startTokensMetricsProcess(c *cli.Context, interval time.Duration, logger log.ColorLogger) {

    logger.Printlnf("Enter startTokensMetricsProcess")
    timer := time.NewTicker(interval)
    var p *tokensMetricsProcess
    var err error
    // put create process in a loop because it may fail initially
    for ; true; <- timer.C {
        p, err = newTokensMetricsProcess(c, logger)
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
    logger.Printlnf("Exit startTokensMetricsProcess")
}


// Create new tokensMetricsProcess object
func newTokensMetricsProcess(c *cli.Context, logger log.ColorLogger) (*tokensMetricsProcess, error) {

    logger.Printlnf("Enter newTokensMetricsProcess")
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    if err := services.RequireBeaconClientSynced(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }
    bc, err := services.GetBeaconClient(c)
    if err != nil { return nil, err }

    // Initialise metrics
    metrics := tokensGauges {
        tokenSupply:      promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:  "rocketpool",
                Subsystem:  "tokens",
                Name:       "supply_count",
                Help:       "total supply of token",
            },
            []string{"token"},
        ),
        something:      promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:      "rocketpool",
            Subsystem:      "tokens",
            Name:           "something",
            Help:           "something",
        }),
    }

    p := &tokensMetricsProcess {
        rp: rp,
        bc: bc,
        //account: account,
        metrics: metrics,
        logger: logger,
    }

    logger.Printlnf("Exit newTokensMetricsProcess")
    return p, nil
}


// Update tokens metrics
func (p *tokensMetricsProcess) updateMetrics() error {
    p.logger.Printlnf("Enter tokens updateMetrics")

    rethSupply, err := tokens.GetRETHTotalSupply(p.rp, nil)
    rplFixedSupply, err := tokens.GetFixedSupplyRPLTotalSupply(p.rp, nil)
    rplSupply, err := tokens.GetRPLTotalSupply(p.rp, nil)
    if err != nil { return err }

    p.metrics.tokenSupply.With(prometheus.Labels{"token":"rETH"}).Set(eth.WeiToEth(rethSupply))
    p.metrics.tokenSupply.With(prometheus.Labels{"token":"fixed-supply RPL"}).Set(eth.WeiToEth(rplFixedSupply))
    p.metrics.tokenSupply.With(prometheus.Labels{"token":"RPL"}).Set(eth.WeiToEth(rplSupply))

    p.logger.Printlnf("Exit tokens updateMetrics")
    return err
}



