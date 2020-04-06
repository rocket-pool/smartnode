package metrics

import (
    "context"
    "errors"
    "time"

    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Eth1 metrics process
type Eth1MetricsProcess struct {
    p                   *services.Provider
    blockNumber         prometheus.Gauge
    syncing             prometheus.Gauge
    syncStartingBlock   prometheus.Gauge
    syncCurrentBlock    prometheus.Gauge
    syncHighestBlock    prometheus.Gauge
}


// Start eth1 metrics process
func StartEth1MetricsProcess(p *services.Provider) {

    // Initialise process / register metrics
    process := &Eth1MetricsProcess{
        p: p,
        blockNumber: promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:  "smartnode",
            Subsystem:  "eth1",
            Name:       "block_number",
            Help:       "The current Eth 1.0 block number",
        }),
        syncing: promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:  "smartnode",
            Subsystem:  "eth1",
            Name:       "syncing",
            Help:       "Whether the Eth 1.0 node is currently syncing",
        }),
        syncStartingBlock: promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:  "smartnode",
            Subsystem:  "eth1",
            Name:       "sync_starting_block",
            Help:       "The block that the Eth 1.0 node started syncing at",
        }),
        syncCurrentBlock: promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:  "smartnode",
            Subsystem:  "eth1",
            Name:       "sync_current_block",
            Help:       "The block that the Eth 1.0 node is currently synced to",
        }),
        syncHighestBlock: promauto.NewGauge(prometheus.GaugeOpts{
            Namespace:  "smartnode",
            Subsystem:  "eth1",
            Name:       "sync_highest_block",
            Help:       "The highest block that the Eth 1.0 node is syncing to",
        }),
    }

    // Start
    process.start()

}


// Start process
func (p *Eth1MetricsProcess) start() {

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
func (p *Eth1MetricsProcess) update() {

    // Data channels
    headerChannel := make(chan *types.Header)
    syncProgressChannel := make(chan *ethereum.SyncProgress)
    errorChannel := make(chan error)

    // Get latest block header
    go (func() {
        if header, err := p.p.Client.HeaderByNumber(context.Background(), nil); err != nil {
            errorChannel <- errors.New("Error retrieving Eth 1.0 block header: " + err.Error())
        } else {
            headerChannel <- header
        }
    })()

    // Get sync progress
    go (func() {
        if progress, err := p.p.Client.SyncProgress(context.Background()); err != nil {
            errorChannel <- errors.New("Error retrieving Eth 1.0 client sync progress: " + err.Error())
        } else {
            syncProgressChannel <- progress
        }
    })()

    // Receive data
    var header *types.Header
    var syncProgress *ethereum.SyncProgress
    for received := 0; received < 2; {
        select {
            case header = <-headerChannel:
                received++
            case syncProgress = <-syncProgressChannel:
                received++
            case err := <-errorChannel:
                p.p.Log.Println(err)
                return
        }
    }

    // Set metrics
    p.blockNumber.Set(float64(header.Number.Uint64()))
    if syncProgress != nil {
        p.syncing.Set(1)
        p.syncStartingBlock.Set(float64(syncProgress.StartingBlock))
        p.syncCurrentBlock.Set(float64(syncProgress.CurrentBlock))
        p.syncHighestBlock.Set(float64(syncProgress.HighestBlock))
    } else {
        p.syncing.Set(0)
        p.syncStartingBlock.Set(0)
        p.syncCurrentBlock.Set(0)
        p.syncHighestBlock.Set(0)
    }

}

