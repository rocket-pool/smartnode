package collectors

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// Represents the collector for the solo migration check metrics
type SoloMigrationCollector struct {

	// The total number of vacant minipools
	totalMinipoolsDesc *prometheus.Desc

	// The number of solo migration cancellations due to the validator not existing
	doesntExistDesc *prometheus.Desc

	// The number of solo migration cancellations due an invalid validator state
	invalidStateDesc *prometheus.Desc

	// The number of solo migration cancellations due a time out
	timedOutDesc *prometheus.Desc

	// The number of solo migration cancellations due invalid withdrawal credentials
	invalidCredentialsDesc *prometheus.Desc

	// The number of solo migration cancellations due balance being too low
	balanceTooLowDesc *prometheus.Desc

	// The time of the latest block that the check was run against
	latestBlockTimeDesc *prometheus.Desc

	// Counters
	TotalMinipools     float64
	DoesntExist        float64
	InvalidState       float64
	TimedOut           float64
	InvalidCredentials float64
	BalanceTooLow      float64
	LatestBlockTime    float64

	// Mutex
	UpdateLock *sync.Mutex
}

// Create a new ScrubCollector instance
func NewSoloMigrationCollector() *SoloMigrationCollector {
	subsystem := "solo_migration"
	return &SoloMigrationCollector{
		totalMinipoolsDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "total_minipools"),
			"The total number of vacant minipools",
			nil, nil,
		),
		doesntExistDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "doesnt_exist"),
			"The number of solo migration cancellations due to the validator not existing",
			nil, nil,
		),
		invalidStateDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "invalid_state"),
			"The number of solo migration cancellations due an invalid validator state",
			nil, nil,
		),
		timedOutDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "timed_out"),
			"The number of solo migration cancellations due a time out",
			nil, nil,
		),
		invalidCredentialsDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "invalid_credentials"),
			"The number of solo migration cancellations due invalid withdrawal credentials",
			nil, nil,
		),
		balanceTooLowDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "balance_too_low"),
			"The number of solo migration cancellations due balance being too low",
			nil, nil,
		),
		latestBlockTimeDesc: prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, "latest_block_time"),
			"The time of the latest block that the check was run against",
			nil, nil,
		),
		UpdateLock: &sync.Mutex{},
	}
}

// Write metric descriptions to the Prometheus channel
func (c *SoloMigrationCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- c.totalMinipoolsDesc
	channel <- c.doesntExistDesc
	channel <- c.invalidStateDesc
	channel <- c.timedOutDesc
	channel <- c.invalidCredentialsDesc
	channel <- c.balanceTooLowDesc
	channel <- c.latestBlockTimeDesc
}

// Collect the latest metric values and pass them to Prometheus
func (c *SoloMigrationCollector) Collect(channel chan<- prometheus.Metric) {
	// Sync
	c.UpdateLock.Lock()
	defer c.UpdateLock.Unlock()

	// Update all of the metrics
	channel <- prometheus.MustNewConstMetric(
		c.totalMinipoolsDesc, prometheus.GaugeValue, c.TotalMinipools)
	channel <- prometheus.MustNewConstMetric(
		c.doesntExistDesc, prometheus.GaugeValue, c.DoesntExist)
	channel <- prometheus.MustNewConstMetric(
		c.invalidStateDesc, prometheus.GaugeValue, c.InvalidState)
	channel <- prometheus.MustNewConstMetric(
		c.timedOutDesc, prometheus.GaugeValue, c.TimedOut)
	channel <- prometheus.MustNewConstMetric(
		c.invalidCredentialsDesc, prometheus.GaugeValue, c.InvalidCredentials)
	channel <- prometheus.MustNewConstMetric(
		c.balanceTooLowDesc, prometheus.GaugeValue, c.BalanceTooLow)
	channel <- prometheus.MustNewConstMetric(
		c.latestBlockTimeDesc, prometheus.GaugeValue, c.LatestBlockTime)
}
