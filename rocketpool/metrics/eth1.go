package metrics

import (
    "context"

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
        return getEth1BlockNumber(p)
    })

}


// Get block number
func getEth1BlockNumber(p *services.Provider) float64 {

    // Get latest block header
    header, err := p.Client.HeaderByNumber(context.Background(), nil)
    if err != nil { return 0 }

    // Return block number
    return float64(header.Number.Uint64())

}

