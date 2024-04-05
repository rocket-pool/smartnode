package watchtower

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/watchtower/collectors"
	"github.com/rocket-pool/smartnode/shared/keys"
)

func runMetricsServer(sp *services.ServiceProvider, logger *log.Logger, scrubCollector *collectors.ScrubCollector, bondReductionCollector *collectors.BondReductionCollector, soloMigrationCollector *collectors.SoloMigrationCollector) *http.Server {
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

	// Set up Prometheus
	registry := prometheus.NewRegistry()
	registry.MustRegister(scrubCollector)
	registry.MustRegister(bondReductionCollector)
	registry.MustRegister(soloMigrationCollector)

	// Start the HTTP server
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	metricsAddress := os.Getenv("WATCHTOWER_METRICS_ADDRESS")
	metricsPort := cfg.Metrics.WatchtowerMetricsPort.Value
	logger.Info("Starting metrics exporter.", slog.String(keys.UrlKey, fmt.Sprintf("%s:%d", metricsAddress, metricsPort)))
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

	// Run the server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", metricsAddress, metricsPort),
		Handler: nil,
	}

	go func() {
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Error running metrics HTTP server", log.Err(err))
		}
	}()

	return server
}
