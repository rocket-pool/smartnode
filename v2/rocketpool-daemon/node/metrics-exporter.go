package node

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/node/collectors"
	wc "github.com/rocket-pool/smartnode/v2/rocketpool-daemon/watchtower/collectors"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

func runMetricsServer(ctx context.Context, sp *services.ServiceProvider, logger *log.Logger, stateLocker *collectors.StateLocker, wg *sync.WaitGroup,
	scrubCollector *wc.ScrubCollector, bondReductionCollector *wc.BondReductionCollector, soloMigrationCollector *wc.SoloMigrationCollector) *http.Server {
	// Get services
	cfg := sp.GetConfig()

	// Return if metrics are disabled
	if !cfg.Metrics.EnableMetrics.Value {
		if strings.ToLower(os.Getenv("ENABLE_METRICS")) == "true" {
			logger.Info("ENABLE_METRICS override set to true, will start Metrics exporter anyway!")
		} else {
			return nil
		}
	}

	// Create the collectors
	demandCollector := collectors.NewDemandCollector(logger, sp, stateLocker)
	performanceCollector := collectors.NewPerformanceCollector(logger, sp, stateLocker)
	supplyCollector := collectors.NewSupplyCollector(logger, sp, stateLocker)
	rplCollector := collectors.NewRplCollector(logger, sp, stateLocker)
	odaoCollector := collectors.NewOdaoCollector(logger, sp, stateLocker)
	nodeCollector := collectors.NewNodeCollector(logger, ctx, sp, stateLocker)
	trustedNodeCollector := collectors.NewTrustedNodeCollector(logger, sp, stateLocker)
	beaconCollector := collectors.NewBeaconCollector(logger, ctx, sp, stateLocker)
	smoothingPoolCollector := collectors.NewSmoothingPoolCollector(logger, sp, stateLocker)
	snapshotCollector := collectors.NewSnapshotCollector(logger, sp)

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

	// Watchtower collectors
	registry.MustRegister(scrubCollector)
	registry.MustRegister(bondReductionCollector)
	registry.MustRegister(soloMigrationCollector)

	// Start the HTTP server
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	metricsAddress := os.Getenv("NODE_METRICS_ADDRESS")
	metricsPort := cfg.Metrics.DaemonMetricsPort.Value
	logger.Info("Starting metrics exporter.", slog.String(keys.UrlKey, fmt.Sprintf("%s:%d", metricsAddress, metricsPort)))
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
			logger.Error("Error running metrics HTTP server", log.Err(err))
		}
	}()

	return server
}
