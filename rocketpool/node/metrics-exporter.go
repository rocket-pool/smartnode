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
    rp, err := services.GetRocketPool(c)
    if err != nil { return err }

    // Return if metrics are disabled
    if !cfg.Smartnode.EnableMetrics {
        return nil;
    }

    // Create the collectors
    demandCollector := collectors.NewDemandCollector(rp)
    performanceCollector := collectors.NewPerformanceCollector(rp)
    supplyCollector := collectors.NewSupplyCollector(rp)
    rplCollector := collectors.NewRplCollector(rp)
    odaoCollector := collectors.NewOdaoCollector(rp)

    // Set up Prometheus
    registry := prometheus.NewRegistry()
    registry.MustRegister(demandCollector)
    registry.MustRegister(performanceCollector)
    registry.MustRegister(supplyCollector)
    registry.MustRegister(rplCollector)
    registry.MustRegister(odaoCollector)
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
        wrappedError := fmt.Errorf("Error running HTTP server: %w", err)
        return wrappedError
    }

    return nil

}