package node

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rocket-pool/smartnode/rocketpool/node/collectors"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/urfave/cli"
)

func runMetricsServer(c *cli.Context, logger log.ColorLogger) error {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return err
	}
	w, err := services.GetWallet(c)
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
	s, err := services.GetSnapshotDelegation(c)
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
	votingId := cfg.Smartnode.GetVotingSnapshotID()
	votingDelegate, err := s.Delegation(nil, nodeAccount.Address, votingId)
	if err != nil {
		return fmt.Errorf("Error getting node delegate: %w", err)
	}
	// Create the collectors
	demandCollector := collectors.NewDemandCollector(rp)
	performanceCollector := collectors.NewPerformanceCollector(rp)
	supplyCollector := collectors.NewSupplyCollector(rp)
	rplCollector := collectors.NewRplCollector(rp)
	odaoCollector := collectors.NewOdaoCollector(rp)
	nodeCollector := collectors.NewNodeCollector(rp, bc, nodeAccount.Address, cfg)
	trustedNodeCollector := collectors.NewTrustedNodeCollector(rp, bc, nodeAccount.Address, cfg)
	beaconCollector := collectors.NewBeaconCollector(rp, bc, ec, nodeAccount.Address)
	snapshotCollector := collectors.NewSnapshotCollector(rp, cfg, nodeAccount.Address, votingDelegate)
	smoothingPoolCollector := collectors.NewSmoothingPoolCollector(rp, ec)

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
	registry.MustRegister(snapshotCollector)
	registry.MustRegister(smoothingPoolCollector)
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	// Start the HTTP server
	metricsAddress := c.GlobalString("metricsAddress")
	metricsPort := c.GlobalUint("metricsPort")
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
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", metricsAddress, metricsPort), nil)
	if err != nil {
		return fmt.Errorf("Error running HTTP server: %w", err)
	}

	return nil

}
