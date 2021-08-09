package node

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rocket-pool/smartnode/rocketpool/node/collectors"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/urfave/cli"
)


func runMetricsServer(c *cli.Context, logger log.ColorLogger) (error) {

    // Get services
    cfg, err := services.GetConfig(c)
    if err != nil { return err }
    w, err := services.GetWallet(c)
    if err != nil { return err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return err }

    // Return if metrics are disabled
    if !cfg.Smartnode.EnableMetrics {
        return nil;
    }
    
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return fmt.Errorf("Error getting node account: %w", err)
    }

    // Create the collectors
    demandCollector := collectors.NewDemandCollector(rp)
    performanceCollector := collectors.NewPerformanceCollector(rp)
    supplyCollector := collectors.NewSupplyCollector(rp)
    rplCollector := collectors.NewRplCollector(rp)
    odaoCollector := collectors.NewOdaoCollector(rp)
    nodeCollector := collectors.NewNodeCollector(rp, nodeAccount.Address)

    // Set up Prometheus
    registry := prometheus.NewRegistry()
    registry.MustRegister(demandCollector)
    registry.MustRegister(performanceCollector)
    registry.MustRegister(supplyCollector)
    registry.MustRegister(rplCollector)
    registry.MustRegister(odaoCollector)
    registry.MustRegister(nodeCollector)
    handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

    // Start the HTTP server
    logger.Printlnf("Starting metrics exporter on %s:%s.", cfg.Smartnode.NodeAddress, cfg.Smartnode.NodePort)
    metricsPath := "/metrics"
    http.Handle(metricsPath, handler)
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`<html>
            <head><title>Rocket Pool Metrics Exporter</title></head>
            <body>
            <h1>Rocket Pool Metrics Exporter</h1>
            <p><a href='` + metricsPath + `'>Metrics</a></p>
            </body>
            </html>`,
        ))
    })
    err = http.ListenAndServe(fmt.Sprintf("%s:%s", cfg.Smartnode.NodeAddress, cfg.Smartnode.NodePort), nil)
    if err != nil {
        return fmt.Errorf("Error running HTTP server: %w", err)
    }

    return nil

}