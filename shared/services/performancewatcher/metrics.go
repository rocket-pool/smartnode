package performancewatcher

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	status func() Status
	desc   []*prometheus.Desc
}

func NewCollector(status func() Status) prometheus.Collector {
	names := []string{"enabled", "live_lag_epochs", "complete_epochs", "pending_epochs", "candidates", "rpc_failures_total", "last_epoch_processing_seconds", "storage_bytes"}
	m := &metrics{status: status}
	for _, name := range names {
		m.desc = append(m.desc, prometheus.NewDesc("rocketpool_performance_watcher_"+name, "Performance watcher "+name, nil, nil))
	}
	return m
}
func (m *metrics) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range m.desc {
		ch <- d
	}
}
func (m *metrics) Collect(ch chan<- prometheus.Metric) {
	s := m.status()
	enabled := float64(0)
	if s.Enabled {
		enabled = 1
	}
	lag := uint64(0)
	if s.LatestMeasurableEpoch > s.LatestCollectedEpoch {
		lag = s.LatestMeasurableEpoch - s.LatestCollectedEpoch
	}
	values := []float64{enabled, float64(lag), float64(s.CompleteEpochs), float64(s.PendingEpochs), float64(s.CandidateCount), float64(s.RPCFailures), s.ProcessingSeconds, float64(s.StorageBytes)}
	for i, v := range values {
		kind := prometheus.GaugeValue
		if i == 5 {
			kind = prometheus.CounterValue
		}
		ch <- prometheus.MustNewConstMetric(m.desc[i], kind, v)
	}
}
