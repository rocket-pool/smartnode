package metrics

import (
    "fmt"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/settings"
)


// RP metrics process
type RocketPoolMetricsProcess struct {
    p                       *services.Provider
    stakingDurationEnabled  map[string]prometheus.Gauge

}


// Start RP metrics process
func StartRocketPoolMetricsProcess(p *services.Provider) {

    // Initialise process / register metrics
    process := &RocketPoolMetricsProcess{
        p: p,
        stakingDurationEnabled: make(map[string]prometheus.Gauge),
    }

    // Start
    process.start()

}


// Start process
func (p *RocketPoolMetricsProcess) start() {

    // Update metrics on interval
    go (func() {
        p.update()
        updateMetricsTimer := time.NewTicker(updateMetricsInterval)
        for _ = range updateMetricsTimer.C {
            p.update()
        }
    })()

}


// Update metrics
func (p *RocketPoolMetricsProcess) update() {

    // Get minipool staking durations
    stakingDurations, err := settings.GetMinipoolStakingDurations(p.p.CM)
    if err != nil {
        p.p.Log.Println(err)
        return
    }

    // Create metrics
    for _, duration := range stakingDurations {
        if _, ok := p.stakingDurationEnabled[duration.Id]; !ok {
            p.stakingDurationEnabled[duration.Id] = promauto.NewGauge(prometheus.GaugeOpts{
                Namespace:  "smartnode",
                Subsystem:  "rocketpool",
                Name:       fmt.Sprintf("staking_enabled_%s", duration.Id),
                Help:       fmt.Sprintf("Whether the '%s' staking duration is enabled", duration.Id),
            })
        }
    }

    // Set metrics
    for _, duration := range stakingDurations {
        if duration.Enabled {
            p.stakingDurationEnabled[duration.Id].Set(1)
        } else {
            p.stakingDurationEnabled[duration.Id].Set(0)
        }
    }

}

