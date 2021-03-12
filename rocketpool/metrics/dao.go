package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/rocketpool-go/dao"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// minipool metrics process
type daoGauges struct {
    memberCount         prometheus.Gauge
    proposalCount       prometheus.Gauge
    proposalStateCount  *prometheus.GaugeVec
}


type daoMetricsProcess struct {
    rp *rocketpool.RocketPool
    metrics daoGauges
    logger log.ColorLogger
}


// Start minipool metrics process
func startDaoMetricsProcess(c *cli.Context, interval time.Duration, logger log.ColorLogger) {

    logger.Printlnf("Enter startDaoMetricsProcess")
    timer := time.NewTicker(interval)
    var p *daoMetricsProcess
    var err error
    // put create process in a loop because it may fail initially
    for ; true; <- timer.C {
        p, err = newDaoMetricsProcss(c, logger)
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
    logger.Printlnf("Exit startDaoMetricsProcess")
}


// Create new minipoolMetricsProcss object
func newDaoMetricsProcss(c *cli.Context, logger log.ColorLogger) (*daoMetricsProcess, error) {

    // Get services
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    if err := services.RequireBeaconClientSynced(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Initialise metrics
    metrics := daoGauges {
        memberCount:            promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:          "rocketpool",
            Subsystem:          "dao",
            Name:               "member_count",
            Help:               "number of members in Rocket Pool dao",
        }),
        proposalCount:          promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:          "rocketpool",
            Subsystem:          "dao",
            Name:               "proposal_count",
            Help:               "number of proposals in Rocket Pool dao",
        }),
        proposalStateCount:    promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace:      "rocketpool",
                Subsystem:      "dao",
                Name:           "proposal_state_count",
                Help:           "the count of various states of Rocket Pool dao proposal",
            },
            []string{"state"},
        ),
    }

    p := &daoMetricsProcess {
        rp: rp,
        metrics: metrics,
        logger: logger,
    }

    return p, nil
}


// Update minipool metrics
func (p *daoMetricsProcess) updateMetrics() error {
    p.logger.Println("Enter dao updateMetrics")

    var memberCount, proposalCount uint64
    var proposals []dao.ProposalDetails

    // Sync
    var wg errgroup.Group

    // Get data
    wg.Go(func() error {
        var err error
        memberCount, err = trustednode.GetMemberCount(p.rp, nil)
        return err
    })

    wg.Go(func() error {
        var err error
        proposalCount, err = dao.GetProposalCount(p.rp, nil)
        return err
    })

    wg.Go(func() error {
        var err error
        proposals, err = dao.GetProposals(p.rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return err
    }

    p.metrics.memberCount.Set(float64(memberCount))
    p.metrics.proposalCount.Set(float64(proposalCount))

    // Tally up proposal states
    stateCounts := make(map[types.ProposalState]uint32, len(types.ProposalStates))
    for _, proposal := range proposals {

        if _, ok := stateCounts[proposal.State]; !ok {
            stateCounts[proposal.State] = 0
        }
        stateCounts[proposal.State]++
    }

    for state, count := range stateCounts {
        p.metrics.proposalStateCount.With(prometheus.Labels{"state":types.ProposalStates[state]}).Set(float64(count))
    }

    return nil
}

