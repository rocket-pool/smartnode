package alerting

import (
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/rocket-pool/smartnode/shared/services/alerting/alertmanager/client"
	apialert "github.com/rocket-pool/smartnode/shared/services/alerting/alertmanager/client/alert"
	"github.com/rocket-pool/smartnode/shared/services/alerting/alertmanager/models"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

const (
	DefaultEndsAtDurationForSeverityInfo     = time.Minute * 5
	DefaultEndsAtDurationForSeverityCritical = time.Minute * 60
)

// fetches the current alerts directly the alertmanager container/application's API.
func FetchAlerts(cfg *config.RocketPoolConfig) ([]*models.GettableAlert, error) {
	// NOTE: don't log to stdout here since this method is on the "api" path and all stdout is parsed as a json "api" response.
	if !isAlertingEnabled(cfg) {
		// metrics are disabled, so no alerts will be fetched.
		return make([]*models.GettableAlert, 0), nil
	}

	//logMessage("Fetching alerts from alertmanager...")
	client := createClient(cfg)
	// request alerts:
	resp, err := client.Alert.GetAlerts(nil)
	if err != nil {
		//logMessage("ERROR fetching alerts from alertmanager.")
		return nil, fmt.Errorf("error fetching alerts from alertmanager: %w", err)
	}
	//logMessage("fetching alerts from alertmanager succeeded (%d).", len(resp.Payload))
	return resp.Payload, nil
}

// Sends an alert when the node automatically prompted a minipool or attempted to (success or failure).
// If alerting/metrics are disabled, this function does nothing.
func AlertMinipoolPromoted(cfg *config.RocketPoolConfig, minipoolAddress common.Address, succeeded bool) error {
	if !isAlertingEnabled(cfg) {
		logMessage("alerting is disabled, not sending AlertMinipoolPromoted.")
		return nil
	}

	// prepare the alert information:
	endsAt := strfmt.DateTime(time.Now().Add(DefaultEndsAtDurationForSeverityInfo))
	severity := SeverityInfo
	succeededOrFailed := "succeeded"
	if !succeeded {
		succeededOrFailed = "failed"
		severity = SeverityCritical
		endsAt = strfmt.DateTime(time.Now().Add(DefaultEndsAtDurationForSeverityCritical))
	}

	alert := createAlert(
		fmt.Sprintf("MinipoolPromoted-%s-%s", succeededOrFailed, minipoolAddress.Hex()),
		fmt.Sprintf("Minipool %s %s", minipoolAddress.Hex(), succeededOrFailed),
		fmt.Sprintf("The vacant minipool with address %s promoted with status %s.", minipoolAddress.Hex(), succeededOrFailed),
		severity,
		endsAt,
		map[string]string{
			"minipool": minipoolAddress.Hex(),
		},
	)
	return sendAlert(alert, cfg)
}

// Sends an alert when the node automatically staked a minipool or attempted to (success or failure).
// If alerting/metrics are disabled, this function does nothing.
func AlertMinipoolStaked(cfg *config.RocketPoolConfig, minipoolAddress common.Address, succeeded bool) error {
	if !isAlertingEnabled(cfg) {
		logMessage("alerting is disabled, not sending AlertMinipoolStaked.")
		return nil
	}

	// prepare the alert information:
	endsAt := strfmt.DateTime(time.Now().Add(DefaultEndsAtDurationForSeverityInfo))
	severity := SeverityInfo
	succeededOrFailed := "succeeded"
	if !succeeded {
		succeededOrFailed = "failed"
		severity = SeverityCritical
		endsAt = strfmt.DateTime(time.Now().Add(DefaultEndsAtDurationForSeverityCritical))
	}

	alert := createAlert(
		fmt.Sprintf("MinipoolStaked-%s-%s", succeededOrFailed, minipoolAddress.Hex()),
		fmt.Sprintf("Minipool %s %s", minipoolAddress.Hex(), succeededOrFailed),
		fmt.Sprintf("The minipool with address %s staked with status %s.", minipoolAddress.Hex(), succeededOrFailed),
		severity,
		endsAt,
		map[string]string{
			"minipool": minipoolAddress.Hex(),
		},
	)
	return sendAlert(alert, cfg)
}

func AlertExecutionClientSyncComplete(cfg *config.RocketPoolConfig) error {
	return alertClientSyncComplete(cfg, ClientKindExecution)
}

func AlertBeaconClientSyncComplete(cfg *config.RocketPoolConfig) error {
	return alertClientSyncComplete(cfg, ClientKindBeacon)
}

type ClientKind string

const (
	ClientKindExecution ClientKind = "Execution"
	ClientKindBeacon    ClientKind = "Beacon"
)

func alertClientSyncComplete(cfg *config.RocketPoolConfig, client ClientKind) error {
	alertName := fmt.Sprintf("%sClientSyncComplete", client)
	if !isAlertingEnabled(cfg) {
		logMessage(fmt.Sprintf("alerting is disabled, not sending %s.", alertName))
		return nil
	}

	alert := createAlert(
		alertName,
		fmt.Sprintf("%s Client Sync Complete", client),
		fmt.Sprintf("The %s client has completed syncing.", client),
		SeverityInfo,
		strfmt.DateTime(time.Now().Add(time.Minute*1)),
		nil,
	)
	return sendAlert(alert, cfg)
}

func sendAlert(alert *models.PostableAlert, cfg *config.RocketPoolConfig) error {
	logMessage("sending alert for %s: %s", alert.Labels["alertname"], alert.Annotations["summary"])

	params := apialert.NewPostAlertsParams().WithDefaults().WithAlerts(models.PostableAlerts{alert})
	client := createClient(cfg)
	_, err := client.Alert.PostAlerts(params)
	if err != nil {
		return fmt.Errorf("error posting alert: %s", err.Error())
	}
	return nil
}

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

func isAlertingEnabled(cfg *config.RocketPoolConfig) bool {
	return cfg.EnableMetrics.Value == true
}

// Creates a uniform alert with the basic labels and annotations we expect.
func createAlert(uniqueName string, summary string, description string, severity Severity, endsAt strfmt.DateTime, extraLabels map[string]string) *models.PostableAlert {
	alert := &models.PostableAlert{
		Annotations: map[string]string{
			"description": description,
			"summary":     summary,
		},
		Alert: models.Alert{
			Labels: map[string]string{
				"alertname": uniqueName,
				"severity":  string(severity),
			},
		},
		EndsAt: endsAt,
	}

	for k, v := range extraLabels {
		alert.Labels[k] = v
	}
	return alert
}

func createClient(cfg *config.RocketPoolConfig) *apiclient.Alertmanager {
	host := fmt.Sprintf("%s:%d", config.AlertmanagerContainerName, cfg.Alertmanager.Port.Value)
	transport := apiclient.DefaultTransportConfig().WithHost(host)
	client := apiclient.NewHTTPClientWithConfig(strfmt.Default, transport)
	return client
}

func logMessage(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("[alerting] %s\n", msg)
}
