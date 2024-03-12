package node

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/log"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/node/collectors"
)

func runMetricsServer(sp *services.ServiceProvider, logger log.ColorLogger, stateLocker *collectors.StateLocker) error {
	// Get services
	cfg := sp.GetConfig()

	// Return if metrics are disabled
	if cfg.EnableMetrics.Value == false {
		if strings.ToLower(os.Getenv("ENABLE_METRICS")) == "true" {
			logger.Printlnf("ENABLE_METRICS override set to true, will start Metrics exporter anyway!")
		} else {
			return nil
		}
	}

	// Create the collectors
	demandCollector := collectors.NewDemandCollector(sp, stateLocker)
	performanceCollector := collectors.NewPerformanceCollector(sp, stateLocker)
	supplyCollector := collectors.NewSupplyCollector(sp, stateLocker)
	rplCollector := collectors.NewRplCollector(sp, stateLocker)
	odaoCollector := collectors.NewOdaoCollector(sp, stateLocker)
	nodeCollector := collectors.NewNodeCollector(sp, stateLocker)
	trustedNodeCollector := collectors.NewTrustedNodeCollector(sp, stateLocker)
	beaconCollector := collectors.NewBeaconCollector(sp, stateLocker)
	smoothingPoolCollector := collectors.NewSmoothingPoolCollector(sp, stateLocker)
	snapshotCollector := collectors.NewSnapshotCollector(sp)

	// Set up Prometheus
	registry := prometheus.NewRegistry()
	registry.MustRegister(demandCollector)
	registry.MustRegister(performanceCollector)
	registry.MustRegister(supplyCollector)
	registry.MustRegister(rplCollector)
	registry.MustRegister(odaoCollector)
	registry.MustRegister(nodeCollector)
	registry.MustRegister(trustedNodeCollector)
	registry.MustRegister(beaconCollector)
	registry.MustRegister(smoothingPoolCollector)
	registry.MustRegister(snapshotCollector)

	// Start the HTTP server
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	metricsAddress := os.Getenv("NODE_METRICS_ADDRESS")
	metricsPort := cfg.NodeMetricsPort.Value.(uint64)
	logger.Printlnf("Starting metrics exporter on %s:%d.", metricsAddress, metricsPort)
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
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", metricsAddress, metricsPort), nil)
	if err != nil {
		return fmt.Errorf("Error running HTTP server: %w", err)
	}

	return nil
}
