package node

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/node/collectors"
)

func runMetricsServer(ctx context.Context, sp *services.ServiceProvider, logger *log.Logger, stateLocker *collectors.StateLocker, wg *sync.WaitGroup) *http.Server {
	// Get services
	cfg := sp.GetConfig()

	// Return if metrics are disabled
	if !cfg.Metrics.EnableMetrics.Value {
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
	nodeCollector := collectors.NewNodeCollector(ctx, sp, stateLocker)
	trustedNodeCollector := collectors.NewTrustedNodeCollector(sp, stateLocker)
	beaconCollector := collectors.NewBeaconCollector(ctx, sp, stateLocker)
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
	metricsPort := cfg.Metrics.DaemonMetricsPort.Value
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

	// Run the server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", metricsAddress, metricsPort),
		Handler: nil,
	}
	go func() {
		defer wg.Done()

		wg.Add(1)
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Printlnf("error running metrics HTTP server: %s", err.Error())
		}
	}()

	return server
}
