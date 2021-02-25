package metrics

import (
    "fmt"
    "sort"
    "strconv"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/urfave/cli"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/types"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    apiNode "github.com/rocket-pool/smartnode/rocketpool/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/beacon"
    "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/hex"
    "github.com/rocket-pool/smartnode/shared/utils/log"
)


const (
    BucketInterval = 0.025
)


// node metrics process
type nodeGauges struct {
    nodeScores             *prometheus.GaugeVec
    nodeScoreHist          *prometheus.GaugeVec
    nodeScoreHistSum       prometheus.Gauge
    nodeScoreHistCount     prometheus.Gauge
    nodeMinipoolCounts     *prometheus.GaugeVec
    minipoolCounts         *prometheus.GaugeVec
}


type nodeMetricsProcess struct {
    rp *rocketpool.RocketPool
    bc beacon.Client
    metrics nodeGauges
    logger log.ColorLogger
}


// Start node metrics process
func startNodeMetricsProcess(c *cli.Context, interval time.Duration, logger log.ColorLogger) {

    logger.Printlnf("Enter startNodeMetricsProcess")
    timer := time.NewTicker(interval)
    var p *nodeMetricsProcess
    var err error
    // put create process in a loop because it may fail initially
    for ; true; <- timer.C {
        p, err = newNodeMetricsProcss(c, logger)
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
    logger.Printlnf("Exit startNodeMetricsProcess")
}


// Create new nodeMetricsProcess object
func newNodeMetricsProcss(c *cli.Context, logger log.ColorLogger) (*nodeMetricsProcess, error) {

    // Get services
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    if err := services.RequireBeaconClientSynced(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }
    bc, err := services.GetBeaconClient(c)
    if err != nil { return nil, err }

    // Initialise metrics
    metrics := nodeGauges {
        nodeScores:         promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:  "rocketpool",
                Subsystem:  "node",
                Name:       "score_eth",
                Help:       "sum of rewards/penalties of the top two minipools for this node",
            },
            []string{"address", "rank"},
        ),
        nodeScoreHist: promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:  "rocketpool",
                Subsystem:  "node",
                Name:       "score_hist_eth",
                Help:       "distribution of sum of rewards/penalties of the top two minipools in rocketpool network",
                },
            []string{"le"},
        ),
        nodeScoreHistSum:   promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:      "rocketpool",
            Subsystem:      "node",
            Name:           "score_hist_eth_sum",
        }),
        nodeScoreHistCount: promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:      "rocketpool",
            Subsystem:      "node",
            Name:           "score_hist_eth_count",
        }),
        nodeMinipoolCounts: promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:  "rocketpool",
                Subsystem:  "node",
                Name:       "minipool_count",
                Help:       "number of activated minipools running for node address",
            },
            []string{"address", "timezone"},
        ),
        minipoolCounts: promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:  "rocketpool",
                Subsystem:  "minipool",
                Name:       "count",
                Help:       "minipools counts with various aggregations",
            },
            []string{"status"},
        ),
    }

    p := &nodeMetricsProcess {
        rp: rp,
        bc: bc,
        metrics: metrics,
        logger: logger,
    }

    return p, nil
}


// Update node metrics
func (p *nodeMetricsProcess) updateMetrics() error {
    p.logger.Println("Enter node updateMetrics")

    nodeRanks, err := apiNode.GetNodeLeader(p.rp, p.bc)
    if err != nil { return err }

    p.updateScore(nodeRanks)
    p.updateHistogram(nodeRanks)
    p.updateNodeMinipoolCount(nodeRanks)
    p.updateMinipoolCount(nodeRanks)

    p.logger.Println("Exit node updateMetrics")
    return nil
}


func (p *nodeMetricsProcess) updateScore(nodeRanks []api.NodeRank) {
    p.metrics.nodeScores.Reset()

    for _, nodeRank := range nodeRanks {

        nodeAddress := hex.AddPrefix(nodeRank.Address.Hex())

        if nodeRank.Score != nil {
            scoreEth := eth.WeiToEth(nodeRank.Score)
            p.metrics.nodeScores.With(prometheus.Labels{"address":nodeAddress, "rank":strconv.Itoa(nodeRank.Rank)}).Set(scoreEth)
        }
    }
}


func (p *nodeMetricsProcess) updateHistogram(nodeRanks []api.NodeRank) {
    p.metrics.nodeScoreHist.Reset()

    if len(nodeRanks) == 0 { return }

    histogram := make(map[float64]int, 100)
    var sumScores float64

    for _, nodeRank := range nodeRanks {

        if nodeRank.Score != nil {
            scoreEth := eth.WeiToEth(nodeRank.Score)

            // find next highest bucket to put in
            bucket := float64(int(scoreEth / BucketInterval)) * BucketInterval
        	if (bucket < scoreEth) {
        	    bucket = bucket + BucketInterval
        	}
            if _, ok := histogram[bucket]; !ok {
                histogram[bucket] = 0
            }
            histogram[bucket]++
            sumScores += scoreEth
        }
    }

    buckets := make([]float64, 0, len(histogram))
    for b := range histogram {
        buckets = append(buckets, b)
    }
    sort.Float64s(buckets)

    accCount := 0
    nextB := buckets[0]
    for _, b := range buckets {

        // fill in the gaps
        for nextB < b {
            p.metrics.nodeScoreHist.With(prometheus.Labels{"le":fmt.Sprintf("%.3f", nextB)}).Set(float64(accCount))
            nextB = nextB + BucketInterval
        }

        accCount += histogram[b]
        p.metrics.nodeScoreHist.With(prometheus.Labels{"le":fmt.Sprintf("%.3f", b)}).Set(float64(accCount))

        nextB = b + BucketInterval
    }

    p.metrics.nodeScoreHistSum.Set(sumScores)
    p.metrics.nodeScoreHistCount.Set(float64(accCount))
}


func (p *nodeMetricsProcess) updateNodeMinipoolCount(nodeRanks []api.NodeRank) {
    p.metrics.nodeMinipoolCounts.Reset()

    for _, nodeRank := range nodeRanks {

        nodeAddress := hex.AddPrefix(nodeRank.Address.Hex())
        minipoolCount := len(nodeRank.Details)
        labels := prometheus.Labels {
            "address":nodeAddress,
            "timezone":nodeRank.TimezoneLocation,
        }
        p.metrics.nodeMinipoolCounts.With(labels).Set(float64(minipoolCount))
    }
}


func (p *nodeMetricsProcess) updateMinipoolCount(nodeRanks []api.NodeRank) {
    p.metrics.minipoolCounts.Reset()

    var totalCount, initializedCount, prelaunchCount, stakingCount, withdrawableCount, dissolvedCount int
    var validatorExistsCount, validatorActiveCount int

    for _, nodeRank := range nodeRanks {
        totalCount += len(nodeRank.Details)
        for _, minipool := range nodeRank.Details {
            switch minipool.Status.Status {
                case types.Initialized:  initializedCount++
                case types.Prelaunch:    prelaunchCount++
                case types.Staking:      stakingCount++
                case types.Withdrawable: withdrawableCount++
                case types.Dissolved:    dissolvedCount++
        	}
            if minipool.Validator.Exists { validatorExistsCount ++ }
            if minipool.Validator.Active { validatorActiveCount ++ }
        }
    }
    p.metrics.minipoolCounts.With(prometheus.Labels{"status":"total"}).Set(float64(totalCount))
    p.metrics.minipoolCounts.With(prometheus.Labels{"status":"initialized"}).Set(float64(initializedCount))
    p.metrics.minipoolCounts.With(prometheus.Labels{"status":"prelaunch"}).Set(float64(prelaunchCount))
    p.metrics.minipoolCounts.With(prometheus.Labels{"status":"staking"}).Set(float64(stakingCount))
    p.metrics.minipoolCounts.With(prometheus.Labels{"status":"withdrawable"}).Set(float64(withdrawableCount))
    p.metrics.minipoolCounts.With(prometheus.Labels{"status":"dissolved"}).Set(float64(dissolvedCount))
    p.metrics.minipoolCounts.With(prometheus.Labels{"status":"validatorExists"}).Set(float64(validatorExistsCount))
    p.metrics.minipoolCounts.With(prometheus.Labels{"status":"validatorActive"}).Set(float64(validatorActiveCount))
}
