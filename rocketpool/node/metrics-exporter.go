package node

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/rocketpool/node/collectors"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

func runMetricsServer(ctx context.Context, c *cli.Command, logger log.ColorLogger, stateLocker *collectors.StateLocker) error {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return err
	}
	w, err := services.GetHdWallet(c)
	if err != nil {
		return err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return err
	}
	reg, err := services.GetRocketSignerRegistry(c)
	if err != nil {
		return err
	}

	// Return if metrics are disabled
	if cfg.EnableMetrics.Value == false {
		if strings.ToLower(os.Getenv("ENABLE_METRICS")) == "true" {
			logger.Printlnf("ENABLE_METRICS override set to true, will start Metrics exporter anyway!")
		} else {
			return nil
		}
	}

	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return fmt.Errorf("Error getting node account: %w", err)
	}

	// Create the collectors
	demandCollector := collectors.NewDemandCollector(rp, stateLocker)
	performanceCollector := collectors.NewPerformanceCollector(rp, stateLocker)
	supplyCollector := collectors.NewSupplyCollector(rp, stateLocker)
	rplCollector := collectors.NewRplCollector(rp, cfg, stateLocker)
	odaoCollector := collectors.NewOdaoCollector(rp, stateLocker)
	nodeCollector := collectors.NewNodeCollector(rp, bc, ec, nodeAccount.Address, cfg, stateLocker)
	trustedNodeCollector := collectors.NewTrustedNodeCollector(rp, bc, nodeAccount.Address, cfg, stateLocker)
	beaconCollector := collectors.NewBeaconCollector(rp, bc, ec, nodeAccount.Address, stateLocker)
	smoothingPoolCollector := collectors.NewSmoothingPoolCollector(rp, ec, stateLocker)
	governanceCollector := collectors.NewGovernanceCollector(rp)
	versionUpdateCollector := collectors.NewVersionUpdateCollector(logger.Printlnf)

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
	registry.MustRegister(governanceCollector)
	registry.MustRegister(versionUpdateCollector)

	// Set up snapshot checking if enabled
	if cfg.Smartnode.GetRocketSignerRegistryAddress() != "" {
		signallingAddress, err := reg.NodeToSigner(&bind.CallOpts{}, nodeAccount.Address)
		if err != nil {
			logger.Printlnf("Error getting the signalling address: %w", err)
			// Set signallingAddress to blank address instead of erroring out of the task loop.
			signallingAddress = common.Address{}
		}
		snapshotCollector := collectors.NewSnapshotCollector(rp, cfg, ec, bc, reg, nodeAccount.Address, signallingAddress)
		registry.MustRegister(snapshotCollector)

	}

	// Start the HTTP server
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	metricsAddress := c.Root().String("metricsAddress")
	metricsPort := c.Root().Uint("metricsPort")
	if metricsPort == 0 {
		metricsPort = uint(cfg.NodeMetricsPort.Value.(uint16))
	}
	logger.Printlnf("Starting metrics exporter on %s:%d.", metricsAddress, metricsPort)
	metricsPath := "/metrics"
	mux := http.NewServeMux()
	mux.Handle(metricsPath, handler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
            <head><title>Rocket Pool Metrics Exporter</title></head>
            <body>
            <h1>Rocket Pool Metrics Exporter</h1>
            <p><a href='` + metricsPath + `'>Metrics</a></p>
            </body>
            </html>`,
		))
		if err != nil {
			logger.Printlnf("Error writing metrics exporter HTML: %v", err)
		}
	})
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", metricsAddress, metricsPort),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("Error running HTTP server: %w", err)
	}

	return nil

}
