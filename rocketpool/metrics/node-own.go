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
    apiNode "github.com/rocket-pool/smartnode/rocketpool/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/beacon"
    "github.com/rocket-pool/smartnode/shared/utils/hex"
    "github.com/rocket-pool/smartnode/shared/utils/log"
)


// minipool metrics process
type nodeOwnGauges struct {
    minipoolBalance       *prometheus.GaugeVec
    nodeTrusted           prometheus.Gauge
    accountBalances       *prometheus.GaugeVec
    rplStake              *prometheus.GaugeVec
    minipoolLimit         prometheus.Gauge
    minipoolCounts        *prometheus.GaugeVec
    withdrawalBalances    *prometheus.GaugeVec
}


type nodeOwnMetricsProcess struct {
    rp *rocketpool.RocketPool
    bc beacon.Client
    account accounts.Account
    metrics nodeOwnGauges
    logger log.ColorLogger
}


// Start minipool metrics process
func startNodeOwnMetricsProcess(c *cli.Context, interval time.Duration, logger log.ColorLogger) {

    logger.Printlnf("Enter startNodeOwnMetricsProcess")
    timer := time.NewTicker(interval)
    var p *nodeOwnMetricsProcess
    var err error
    // put create process in a loop because it may fail initially
    for ; true; <- timer.C {
        p, err = newMinipoolMetricsProcss(c, logger)
        if p != nil && err == nil {
            break;
        }
        logger.Printlnf("nodeOwnMetricsProcess retry loop: %w", err)
    }
    logger.Printlnf("nodeOwnMetricsProcess created")

    // Update metrics on interval
    for ; true; <- timer.C {
        err = p.updateMetrics()
        if err != nil {
            // print error here instead of exit
            logger.Printlnf("Error in updateMetrics: %w", err)
        }
    }
    logger.Printlnf("Exit startNodeOwnMetricsProcess")
}


// Create new minipoolMetricsProcss object
func newMinipoolMetricsProcss(c *cli.Context, logger log.ColorLogger) (*nodeOwnMetricsProcess, error) {

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
    metrics := nodeOwnGauges {
        minipoolBalance:      promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:    "rocketpool",
                Subsystem:    "node_own",
                Name:         "minipool_balance_eth",
                Help:         "balance of validator for own node",
            },
            []string{"address", "validatorPubkey"},
        ),
        nodeTrusted:          promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:        "rocketpool",
            Subsystem:        "node_own",
            Name:             "trusted_bool",
            Help:             "whether this node is oracle node",
        }),
        accountBalances:      promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:    "rocketpool",
                Subsystem:    "node_own",
                Name:         "account_balance",
                Help:         "account balances for own node",
            },
            []string{"token"},
        ),
        rplStake:             promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:    "rocketpool",
                Subsystem:    "node_own",
                Name:         "stake_rpl",
                Help:         "amounts of stake in RPL for own node",
            },
            []string{"status"},
        ),
        minipoolLimit:        promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:        "rocketpool",
            Subsystem:        "node_own",
            Name:             "minipool_limit_count",
            Help:             "minipool limit based on RPL stake for own node",
        }),
        minipoolCounts:       promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:    "rocketpool",
                Subsystem:    "node_own",
                Name:         "minipool_count",
                Help:         "counts of minipools for own node",
            },
            []string{"status"},
        ),
        withdrawalBalances:   promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:    "rocketpool",
                Subsystem:    "node_own",
                Name:         "withdraw_balance",
                Help:         "balances available for withdraw for own node",
            },
            []string{"token"},
        ),
    }

    p := &nodeOwnMetricsProcess {
        rp: rp,
        bc: bc,
        account: account,
        metrics: metrics,
        logger: logger,
    }

    return p, nil
}


// Update minipool metrics
func (p *nodeOwnMetricsProcess) updateMetrics() error {
    p.logger.Println("Enter node-own updateMetrics")

    err1 := p.updateNode()
    err2 := p.updateMinipool()
    err := multierr.Combine(err1, err2)

    p.logger.Printlnf("Exit node-own updateMetrics with %d errors", len(multierr.Errors(err)))
    return err
}


func (p *nodeOwnMetricsProcess) updateNode() (error) {

    // Response
    nodeStatus, err := apiNode.GetStatus(p.rp, p.account)
    if err != nil {
        return err
    }

    p.metrics.nodeTrusted.Set(float64(B2i(nodeStatus.Trusted)))
    p.metrics.accountBalances.With(prometheus.Labels{"token":"ETH"}).Set(eth.WeiToEth(nodeStatus.AccountBalances.ETH))
    p.metrics.accountBalances.With(prometheus.Labels{"token":"RETH"}).Set(eth.WeiToEth(nodeStatus.AccountBalances.RETH))
    p.metrics.accountBalances.With(prometheus.Labels{"token":"RPL"}).Set(eth.WeiToEth(nodeStatus.AccountBalances.RPL))
    p.metrics.rplStake.With(prometheus.Labels{"status":"current"}).Set(eth.WeiToEth(nodeStatus.RplStake))
    p.metrics.rplStake.With(prometheus.Labels{"status":"effective"}).Set(eth.WeiToEth(nodeStatus.EffectiveRplStake))
    p.metrics.rplStake.With(prometheus.Labels{"status":"minimum"}).Set(eth.WeiToEth(nodeStatus.MinimumRplStake))
    p.metrics.minipoolLimit.Set(float64(nodeStatus.MinipoolLimit))
    p.metrics.minipoolCounts.With(prometheus.Labels{"status":"closeAvailable"}).Set(float64(nodeStatus.MinipoolCounts.CloseAvailable))
    p.metrics.minipoolCounts.With(prometheus.Labels{"status":"dissolved"}).Set(float64(nodeStatus.MinipoolCounts.Dissolved))
    p.metrics.minipoolCounts.With(prometheus.Labels{"status":"initialized"}).Set(float64(nodeStatus.MinipoolCounts.Initialized))
    p.metrics.minipoolCounts.With(prometheus.Labels{"status":"prelaunch"}).Set(float64(nodeStatus.MinipoolCounts.Prelaunch))
    p.metrics.minipoolCounts.With(prometheus.Labels{"status":"refundAvailable"}).Set(float64(nodeStatus.MinipoolCounts.RefundAvailable))
    p.metrics.minipoolCounts.With(prometheus.Labels{"status":"staking"}).Set(float64(nodeStatus.MinipoolCounts.Staking))
    p.metrics.minipoolCounts.With(prometheus.Labels{"status":"total"}).Set(float64(nodeStatus.MinipoolCounts.Total))
    p.metrics.minipoolCounts.With(prometheus.Labels{"status":"withdrawable"}).Set(float64(nodeStatus.MinipoolCounts.Withdrawable))
    p.metrics.minipoolCounts.With(prometheus.Labels{"status":"withdrawalAvailable"}).Set(float64(nodeStatus.MinipoolCounts.WithdrawalAvailable))
    if nodeStatus.WithdrawalBalances.ETH != nil {
        p.metrics.withdrawalBalances.With(prometheus.Labels{"token":"ETH"}).Set(eth.WeiToEth(nodeStatus.WithdrawalBalances.ETH))
        p.metrics.withdrawalBalances.With(prometheus.Labels{"token":"RETH"}).Set(eth.WeiToEth(nodeStatus.WithdrawalBalances.RETH))
        p.metrics.withdrawalBalances.With(prometheus.Labels{"token":"RPL"}).Set(eth.WeiToEth(nodeStatus.WithdrawalBalances.RPL))
    }

    return nil
}


func (p *nodeOwnMetricsProcess) updateMinipool() error {

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

