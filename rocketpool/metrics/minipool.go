package metrics

import (
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/ethereum/go-ethereum/accounts"
    "github.com/urfave/cli"
    "go.uber.org/multierr"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    apiMinipool "github.com/rocket-pool/smartnode/rocketpool/api/minipool"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/beacon"
    "github.com/rocket-pool/smartnode/shared/utils/hex"
    "github.com/rocket-pool/smartnode/shared/utils/log"
)


// minipool metrics process
type minipoolGauges struct {
    minipoolBalance        *prometheus.GaugeVec
}


type minipoolMetricsProcess struct {
    rp *rocketpool.RocketPool
    bc beacon.Client
    account accounts.Account
    metrics minipoolGauges
    logger log.ColorLogger
}


// Start minipool metrics process
func startMinipoolMetricsProcess(c *cli.Context, interval time.Duration, logger log.ColorLogger) {

    logger.Printlnf("Enter startMinipoolMetricsProcess")
    timer := time.NewTicker(interval)
    var p *minipoolMetricsProcess
    var err error
    // put create process in a loop because it may fail initially
    for ; true; <- timer.C {
        p, err = newMinipoolMetricsProcss(c, logger)
        if p != nil && err == nil {
            break;
        }
        logger.Printlnf("minipoolMetricsProcess retry loop")
    }
    logger.Printlnf("minipoolMetricsProcess created")

    // Update metrics on interval
    for ; true; <- timer.C {
        err = p.updateMetrics()
        if err != nil {
            // print error here instead of exit
            logger.Printlnf("Error in updateMetrics: %w", err)
        }
    }
    logger.Printlnf("Exit startMinipoolMetricsProcess")
}


// Create new minipoolMetricsProcss object
func newMinipoolMetricsProcss(c *cli.Context, logger log.ColorLogger) (*minipoolMetricsProcess, error) {

    // Get services
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    if err := services.RequireBeaconClientSynced(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }
    bc, err := services.GetBeaconClient(c)
    if err != nil { return nil, err }
    account, err := w.GetNodeAccount()
    if err != nil { return nil, err }

    // Initialise metrics
    metrics := minipoolGauges {
        minipoolBalance:    promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:  "rocketpool",
                Subsystem:  "minipool",
                Name:       "balance_eth",
                Help:       "balance of validator",
            },
            []string{"address", "validatorPubkey"},
        ),
    }

    p := &minipoolMetricsProcess {
        rp: rp,
        bc: bc,
        account: account,
        metrics: metrics,
        logger: logger,
    }

    return p, nil
}


// Update minipool metrics
func (p *minipoolMetricsProcess) updateMetrics() error {
    p.logger.Println("Enter minipool updateMetrics")

    err2 := p.updateMinipool()
    err := multierr.Combine(err2)

    p.logger.Printlnf("Exit minipool updateMetrics with %d errors", len(multierr.Errors(err)))
    return err
}


func (p *minipoolMetricsProcess) updateMinipool() error {

    minipools, err := apiMinipool.GetNodeMinipoolDetails(p.rp, p.bc, p.account.Address)
    if err != nil { return err }

    for _, minipool := range minipools {
        address := hex.AddPrefix(minipool.Node.Address.Hex())
        validatorPubkey := hex.AddPrefix(minipool.ValidatorPubkey.Hex())
        var balance float64
        if minipool.Validator.Balance != nil {
            balance = eth.WeiToEth(minipool.Validator.Balance)
        }

        p.metrics.minipoolBalance.With(prometheus.Labels{"address":address, "validatorPubkey":validatorPubkey}).Set(balance)
    }

    return nil
}
