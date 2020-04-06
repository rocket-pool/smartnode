package metrics

import (
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"

    "github.com/rocket-pool/smartnode/shared/services"
)


// RP metrics process
type RocketPoolMetricsProcess struct {
    p   *services.Provider
}


// Start RP metrics process
func StartRocketPoolMetricsProcess(p *services.Provider) {

    // Initialise process / register metrics
    process := &RocketPoolMetricsProcess{
        p: p,
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

    // Set metrics

}

