package collectors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	semver "github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/rocket-pool/smartnode/shared"
)

const (
	githubLatestReleaseURL = "https://api.github.com/repos/rocket-pool/smartnode/releases/latest"
	versionCheckInterval   = time.Hour
	versionCheckTimeout    = 15 * time.Second
)

// VersionUpdateCollector exposes whether a newer Smart Node release is available.
type VersionUpdateCollector struct {
	versionUpdate     *prometheus.Desc
	versionUpdateInfo *prometheus.Desc
	current           string
	latestURL         string
	client            *http.Client
	logf              func(string, ...interface{})

	mu              sync.Mutex
	updateAvailable float64
	latestVersion   string
	lastChecked     time.Time
}

type githubReleaseResponse struct {
	TagName string `json:"tag_name"`
}

// NewVersionUpdateCollector creates a collector backed by an hourly GitHub release check.
func NewVersionUpdateCollector(logf func(string, ...interface{})) *VersionUpdateCollector {
	return &VersionUpdateCollector{
		versionUpdate: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "version_update"),
			"New Rocket Pool version available",
			nil, nil,
		),
		versionUpdateInfo: prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "version_update_info"),
			"The latest available Rocket Pool version",
			[]string{"version"}, nil,
		),
		current:   shared.RocketPoolVersion(),
		latestURL: githubLatestReleaseURL,
		client: &http.Client{
			Timeout: versionCheckTimeout,
		},
		logf: logf,
	}
}

// Describe writes metric descriptions to the Prometheus channel.
func (collector *VersionUpdateCollector) Describe(channel chan<- *prometheus.Desc) {
	channel <- collector.versionUpdate
	channel <- collector.versionUpdateInfo
}

// Collect emits the latest cached version update status.
func (collector *VersionUpdateCollector) Collect(channel chan<- prometheus.Metric) {
	collector.checkIfDue(context.Background())

	collector.mu.Lock()
	updateAvailable := collector.updateAvailable
	latestVersion := collector.latestVersion
	collector.mu.Unlock()

	channel <- prometheus.MustNewConstMetric(
		collector.versionUpdate, prometheus.GaugeValue, updateAvailable)
	if latestVersion != "" {
		channel <- prometheus.MustNewConstMetric(
			collector.versionUpdateInfo, prometheus.GaugeValue, 1, latestVersion)
	}
}

func (collector *VersionUpdateCollector) checkIfDue(ctx context.Context) {
	collector.mu.Lock()
	defer collector.mu.Unlock()

	if time.Since(collector.lastChecked) < versionCheckInterval {
		return
	}
	collector.lastChecked = time.Now()

	updateAvailable, latestVersion, err := collector.checkForUpdate(ctx)
	if err != nil {
		if collector.logf != nil {
			collector.logf("Error checking latest Rocket Pool release: %v", err)
		}
		return
	}

	if updateAvailable {
		collector.updateAvailable = 1
	} else {
		collector.updateAvailable = 0
	}
	collector.latestVersion = latestVersion
}

func (collector *VersionUpdateCollector) checkForUpdate(ctx context.Context) (bool, string, error) {
	latest, err := collector.getLatestVersion(ctx)
	if err != nil {
		return false, "", err
	}

	updateAvailable, err := isNewerVersion(collector.current, latest)
	if err != nil {
		return false, "", err
	}

	return updateAvailable, latest, nil
}

func (collector *VersionUpdateCollector) getLatestVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, collector.latestURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating GitHub release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "rocketpool-smartnode")

	resp, err := collector.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error fetching latest GitHub release: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			collector.logf("Error closing GitHub release response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub release request returned status %s", resp.Status)
	}

	var release githubReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("error decoding latest GitHub release: %w", err)
	}
	if strings.TrimSpace(release.TagName) == "" {
		return "", fmt.Errorf("latest GitHub release did not include a tag_name")
	}

	return release.TagName, nil
}

func isNewerVersion(currentVersion string, latestVersion string) (bool, error) {
	current, err := semver.ParseTolerant(strings.TrimSpace(currentVersion))
	if err != nil {
		return false, fmt.Errorf("error parsing current version %q: %w", currentVersion, err)
	}

	latest, err := semver.ParseTolerant(strings.TrimSpace(latestVersion))
	if err != nil {
		return false, fmt.Errorf("error parsing latest version %q: %w", latestVersion, err)
	}

	return latest.Compare(current) > 0, nil
}
