package metrics

import (
    "context"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Register metrics
func registerEth1Metrics(p *services.Provider) {

    // Block number
    promauto.NewGaugeFunc(prometheus.GaugeOpts{
        Namespace:  "smartnode",
        Subsystem:  "eth1",
        Name:       "block_number",
        Help:       "The current Eth 1.0 block number",
    }, func() float64 {
        return getEth1BlockNumber(p.Client)
    })

    // Sync status
    promauto.NewGaugeFunc(prometheus.GaugeOpts{
        Namespace:  "smartnode",
        Subsystem:  "eth1",
        Name:       "syncing",
        Help:       "Whether the Eth 1.0 client is currently syncing",
    }, func() float64 {
        return getEth1SyncStatus(p.Client)
    })

}


// Get block number
func getEth1BlockNumber(client *ethclient.Client) float64 {

    // Get latest block header
    header, err := client.HeaderByNumber(context.Background(), nil)
    if err != nil { return 0 }

    // Return block number
    return float64(header.Number.Uint64())

}


// Get sync status
func getEth1SyncStatus(client *ethclient.Client) float64 {

    // Get sync progress
    progress, err := client.SyncProgress(context.Background())
    if err != nil { return 0 }

    // Return status
    if progress != nil {
        return 1
    } else {
        return 0
    }

}

