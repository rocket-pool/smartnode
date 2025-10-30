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
	"github.com/rocket-pool/smartnode/shared/types/api"
)

const (
	DefaultEndsAtDurationForSeverityInfo     = time.Minute * 5
	DefaultEndsAtDurationForSeverityCritical = time.Minute * 60
)

// fetches the current alerts directly the alertmanager container/application's API.
// If alerting/metrics are disabled, this function returns an empty array.
func FetchAlerts(cfg *config.RocketPoolConfig) ([]*models.GettableAlert, error) {
	// NOTE: don't log to stdout here since this method is on the "api" path and all stdout is parsed as a json "api" response.
	if !isAlertingEnabled(cfg) {
		// metrics are disabled, so no alerts will be fetched.
		return nil, nil
	}

	//logMessage("Fetching alerts from alertmanager...")
	client := createClient(cfg)
	// request alerts:
	resp, err := client.Alert.GetAlerts(nil)
	if err != nil {
		return nil, fmt.Errorf("error fetching alerts from alertmanager: %w", err)
	}
	return resp.Payload, nil
}

func FetchNodeAlerts(cfg *config.RocketPoolConfig) ([]api.NodeAlert, error) {
	alerts, err := FetchAlerts(cfg)
	if err != nil {
		return nil, err
	}

	response := make([]api.NodeAlert, len(alerts)+1)

	for i, a := range alerts {
		response[i] = api.NodeAlert{
			State:       *a.Status.State,
			Labels:      a.Labels,
			Annotations: a.Annotations,
		}
	}

	labels := map[string]string{
		"alertname": "NodeAlert",
		"severity":  "critical",
	}
	annotations := map[string]string{
		"summary":     "Node Alert",
		"description": "Node Alert",
	}
	response[len(alerts)] = api.NodeAlert{
		State:       "active",
		Labels:      labels,
		Annotations: annotations,
	}

	return response, nil
}

// Sends an alert when the node automatically changed a node's fee recipient or attempted to (success or failure).
// If alerting/metrics are disabled, this function does nothing.
func AlertFeeRecipientChanged(cfg *config.RocketPoolConfig, newFeeRecipient common.Address, succeeded bool) error {
	if !isAlertingEnabled(cfg) {
		logMessage("alerting is disabled, not sending AlertFeeRecipientChanged.")
		return nil
	}

	if cfg.Alertmanager.AlertEnabled_FeeRecipientChanged.Value != true {
		logMessage("alert for FeeRecipientChanged is disabled, not sending.")
		return nil
	}

	// prepare the alert information:
	endsAt, severity, succeededOrFailedText := getAlertSettingsForEvent(succeeded)
	alert := createAlert(
		fmt.Sprintf("FeeRecipientChanged-%s-%s", succeededOrFailedText, newFeeRecipient.Hex()),
		fmt.Sprintf("Fee Recipient Change %s", succeededOrFailedText),
		fmt.Sprintf("The fee recipient was changed to %s with status %s.", newFeeRecipient.Hex(), succeededOrFailedText),
		severity,
		endsAt,
		map[string]string{},
	)
	return sendAlert(alert, cfg)
}

// Sends an alert when the node automatically reduced a minipool's bond or attempted to (success or failure).
// If alerting/metrics are disabled, this function does nothing.
func AlertMinipoolBondReduced(cfg *config.RocketPoolConfig, minipoolAddress common.Address, succeeded bool) error {
	if !isAlertingEnabled(cfg) {
		logMessage("alerting is disabled, not sending AlertMinipoolBondReduced.")
		return nil
	}

	if cfg.Alertmanager.AlertEnabled_MinipoolBondReduced.Value != true {
		logMessage("alert for MinipoolBondReduced is disabled, not sending.")
		return nil
	}

	// prepare the alert information:
	endsAt, severity, succeededOrFailedText := getAlertSettingsForEvent(succeeded)
	alert := createAlert(
		fmt.Sprintf("MinipoolBondReduced-%s-%s", succeededOrFailedText, minipoolAddress.Hex()),
		fmt.Sprintf("Minipool %s reduce bond %s", minipoolAddress.Hex(), succeededOrFailedText),
		fmt.Sprintf("The minipool with address %s reduced bond with status %s.", minipoolAddress.Hex(), succeededOrFailedText),
		severity,
		endsAt,
		map[string]string{
			"minipool": minipoolAddress.Hex(),
		},
	)
	return sendAlert(alert, cfg)

}

// Sends an alert when the node automatically distributes a minipool's balance (success or failure).
// If alerting/metrics are disabled, this function does nothing.
func AlertMinipoolBalanceDistributed(cfg *config.RocketPoolConfig, minipoolAddress common.Address, succeeded bool) error {
	if !isAlertingEnabled(cfg) {
		logMessage("alerting is disabled, not sending AlertMinipoolBalanceDistributed.")
		return nil
	}

	if cfg.Alertmanager.AlertEnabled_MinipoolBalanceDistributed.Value != true {
		logMessage("alert for MinipoolBalanceDistributed is disabled, not sending.")
		return nil
	}

	// prepare the alert information:
	endsAt, severity, succeededOrFailedText := getAlertSettingsForEvent(succeeded)
	alert := createAlert(
		fmt.Sprintf("MinipoolBalanceDistributed-%s-%s", succeededOrFailedText, minipoolAddress.Hex()),
		fmt.Sprintf("Minipool %s balance distributed %s", minipoolAddress.Hex(), succeededOrFailedText),
		fmt.Sprintf("The minipool with address %s had its balance distributed with status %s.", minipoolAddress.Hex(), succeededOrFailedText),
		severity,
		endsAt,
		map[string]string{
			"minipool": minipoolAddress.Hex(),
		},
	)
	return sendAlert(alert, cfg)
}

// Sends an alert when the node automatically prompted a minipool or attempted to (success or failure).
// If alerting/metrics are disabled, this function does nothing.
func AlertMinipoolPromoted(cfg *config.RocketPoolConfig, minipoolAddress common.Address, succeeded bool) error {
	if !isAlertingEnabled(cfg) {
		logMessage("alerting is disabled, not sending AlertMinipoolPromoted.")
		return nil
	}

	if cfg.Alertmanager.AlertEnabled_MinipoolPromoted.Value != true {
		logMessage("alert for MinipoolPromoted is disabled, not sending.")
		return nil
	}

	// prepare the alert information:
	endsAt, severity, succeededOrFailedText := getAlertSettingsForEvent(succeeded)
	alert := createAlert(
		fmt.Sprintf("MinipoolPromoted-%s-%s", succeededOrFailedText, minipoolAddress.Hex()),
		fmt.Sprintf("Minipool %s promote %s", minipoolAddress.Hex(), succeededOrFailedText),
		fmt.Sprintf("The vacant minipool with address %s promoted with status %s.", minipoolAddress.Hex(), succeededOrFailedText),
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

	if cfg.Alertmanager.AlertEnabled_MinipoolStaked.Value != true {
		logMessage("alert for MinipoolStaked is disabled, not sending.")
		return nil
	}

	// prepare the alert information:
	endsAt, severity, succeededOrFailedText := getAlertSettingsForEvent(succeeded)

	alert := createAlert(
		fmt.Sprintf("MinipoolStaked-%s-%s", succeededOrFailedText, minipoolAddress.Hex()),
		fmt.Sprintf("Minipool %s stake %s", minipoolAddress.Hex(), succeededOrFailedText),
		fmt.Sprintf("The minipool with address %s staked with status %s.", minipoolAddress.Hex(), succeededOrFailedText),
		severity,
		endsAt,
		map[string]string{
			"minipool": minipoolAddress.Hex(),
		},
	)
	return sendAlert(alert, cfg)
}

// Gets various settings for an alert based on whether a process succeeded or failed.
func getAlertSettingsForEvent(succeeded bool) (strfmt.DateTime, Severity, string) {
	endsAt := strfmt.DateTime(time.Now().Add(DefaultEndsAtDurationForSeverityInfo))
	severity := SeverityInfo
	if !succeeded {
		severity = SeverityCritical
		endsAt = strfmt.DateTime(time.Now().Add(DefaultEndsAtDurationForSeverityCritical))
	}
	succeededOrFailedText := "failed"
	if succeeded {
		succeededOrFailedText = "succeeded"
	}
	return endsAt, severity, succeededOrFailedText
}

func AlertExecutionClientSyncComplete(cfg *config.RocketPoolConfig) error {
	if cfg.Alertmanager.AlertEnabled_ExecutionClientSyncComplete.Value != true {
		logMessage("alert for ExecutionClientSyncComplete is disabled, not sending.")
		return nil
	}
	return alertClientSyncComplete(cfg, ClientKindExecution)
}

func AlertBeaconClientSyncComplete(cfg *config.RocketPoolConfig) error {
	if cfg.Alertmanager.AlertEnabled_BeaconClientSyncComplete.Value != true {
		logMessage("alert for BeaconClientSyncComplete is disabled, not sending.")
		return nil
	}
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
		logMessage("alerting is disabled, not sending %s.", alertName)
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
	return cfg.Alertmanager.EnableAlerting.Value == true
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
	// use the alertmanager container name for the hostname
	host := fmt.Sprintf("%s:%d", config.AlertmanagerContainerName, cfg.Alertmanager.Port.Value)

	if cfg.IsNativeMode {
		host = fmt.Sprintf("%s:%d", cfg.Alertmanager.NativeModeHost.Value, cfg.Alertmanager.NativeModePort.Value)
	}

	transport := apiclient.DefaultTransportConfig().WithHost(host)
	client := apiclient.NewHTTPClientWithConfig(strfmt.Default, transport)
	return client
}

func logMessage(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("[alerting] %s\n", msg)
}
