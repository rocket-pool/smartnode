package metrics

import (
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Eth2 metrics process
type Eth2MetricsProcess struct {
    p                   *services.Provider
    epochNumber         prometheus.Gauge
    finalizedEpoch      prometheus.Gauge
    justifiedEpoch      prometheus.Gauge
}


// Start eth2 metrics process
func StartEth2MetricsProcess(p *services.Provider) {

    // Initialise process / register metrics
    process := &Eth2MetricsProcess{
        p: p,
        epochNumber: promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:  "smartnode",
            Subsystem:  "eth2",
            Name:       "epoch_number",
            Help:       "The current Eth 2.0 epoch number",
        }),
        finalizedEpoch: promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:  "smartnode",
            Subsystem:  "eth2",
            Name:       "finalized_epoch",
            Help:       "The current highest finalized Eth 2.0 epoch",
        }),
        justifiedEpoch: promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:  "smartnode",
            Subsystem:  "eth2",
            Name:       "justified_epoch",
            Help:       "The current highest justified Eth 2.0 epoch",
        }),
    }

    // Start
    process.start()

}


// Start process
func (p *Eth2MetricsProcess) start() {

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
func (p *Eth2MetricsProcess) update() {

    // Get beacon head
    head, err := p.p.Beacon.GetBeaconHead()
    if err != nil {
        p.p.Log.Println(err)
        return
    }

    // Set metrics
    p.epochNumber.Set(float64(head.Epoch))
    p.finalizedEpoch.Set(float64(head.FinalizedEpoch))
    p.justifiedEpoch.Set(float64(head.JustifiedEpoch))

}

