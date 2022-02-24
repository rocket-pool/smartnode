package watchtower

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rocket-pool/smartnode/rocketpool/watchtower/collectors"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/urfave/cli"
)

func runMetricsServer(c *cli.Context, logger log.ColorLogger, scrubCollector *collectors.ScrubCollector) error {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return err
	}

	// Return if metrics are disabled
	if cfg.EnableMetrics.Value == false {
		return nil
	}

	// Set up Prometheus
	registry := prometheus.NewRegistry()
	registry.MustRegister(scrubCollector)
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	// Start the HTTP server
	metricsAddress := c.GlobalString("metricsAddress")
	metricsPort := c.GlobalUint("metricsPort")
	logger.Printlnf("Starting metrics exporter on %s:%d.", metricsAddress, metricsPort)
	metricsPath := "/metrics"
	http.Handle(metricsPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
            <head><title>Rocket Pool Watchtower Metrics Exporter</title></head>
            <body>
            <h1>Rocket Pool Watchtower Metrics Exporter</h1>
            <p><a href='` + metricsPath + `'>Metrics</a></p>
            </body>
            </html>`,
		))
	})
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", metricsAddress, metricsPort), nil)
	if err != nil {
		return fmt.Errorf("Error running HTTP server: %w", err)
	}

	return nil

}
