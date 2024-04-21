package alerting

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-openapi/strfmt"
	"github.com/rocket-pool/node-manager-core/log"
	apiclient "github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/alerting/alertmanager/client"
	apialert "github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/alerting/alertmanager/client/alert"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/alerting/alertmanager/models"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

const (
	DefaultEndsAtDurationForSeverityInfo     = time.Minute * 5
	DefaultEndsAtDurationForSeverityCritical = time.Minute * 60
)

type AlertFetcher struct {
	enabled            bool
	alertmanagerClient *apiclient.Alertmanager
	cfg                *config.SmartNodeConfig
}

func NewAlertFetcher(cfg *config.SmartNodeConfig) *AlertFetcher {
	return &AlertFetcher{
		enabled:            true,
		alertmanagerClient: createClient(cfg),
		cfg:                cfg,
	}
}

// fetches the current alerts directly the alertmanager container/application's API.
// If alerting/metrics are disabled, this function returns an empty array.
func (a *AlertFetcher) FetchAlerts() ([]*models.GettableAlert, error) {
	if !a.enabled {
		// metrics are disabled, so no alerts will be fetched.
		return nil, nil
	}

	// request alerts:
	resp, err := a.alertmanagerClient.Alert.GetAlerts(nil)
	if err != nil {
		return nil, fmt.Errorf("error fetching alerts from alertmanager: %w", err)
	}
	return resp.Payload, nil
}

type Alerter struct {
	enabled            bool
	alertmanagerClient *apiclient.Alertmanager
	log                *slog.Logger
	cfg                *config.SmartNodeConfig
}

func NewAlerter(cfg *config.SmartNodeConfig, l *log.Logger) *Alerter {
	if !cfg.Alertmanager.EnableAlerting.Value {
		return &Alerter{enabled: false}
	}
	return &Alerter{
		enabled:            true,
		alertmanagerClient: createClient(cfg),
		log:                l.With(slog.String(keys.ModuleKey, "alerting")),
		cfg:                cfg,
	}
}

// Sends an alert when the node automatically changed a node's fee recipient or attempted to (success or failure).
// If alerting/metrics are disabled, this function does nothing.
func (a *Alerter) AlertFeeRecipientChanged(newFeeRecipient common.Address, succeeded bool) {
	if !a.enabled {
		return
	}

	if !a.cfg.Alertmanager.AlertEnabled_FeeRecipientChanged.Value {
		return
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
	a.sendAlert(alert)
}

// Sends an alert when the node automatically reduced a minipool's bond or attempted to (success or failure).
// If alerting/metrics are disabled, this function does nothing.
func (a *Alerter) AlertMinipoolBondReduced(minipoolAddress common.Address, succeeded bool) {
	if !a.enabled {
		return
	}

	if !a.cfg.Alertmanager.AlertEnabled_MinipoolBondReduced.Value {
		return
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
	a.sendAlert(alert)

}

// Sends an alert when the node automatically distributes a minipool's balance (success or failure).
// If alerting/metrics are disabled, this function does nothing.
func (a *Alerter) AlertMinipoolBalanceDistributed(minipoolAddress common.Address, succeeded bool) {
	if !a.enabled {
		return
	}

	if !a.cfg.Alertmanager.AlertEnabled_MinipoolBalanceDistributed.Value {
		return
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
	a.sendAlert(alert)
}

// Sends an alert when the node automatically prompted a minipool or attempted to (success or failure).
// If alerting/metrics are disabled, this function does nothing.
func (a *Alerter) AlertMinipoolPromoted(minipoolAddress common.Address, succeeded bool) {
	if a.enabled {
		return
	}

	if !a.cfg.Alertmanager.AlertEnabled_MinipoolPromoted.Value {
		return
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
	a.sendAlert(alert)
}

// Sends an alert when the node automatically staked a minipool or attempted to (success or failure).
// If alerting/metrics are disabled, this function does nothing.
func (a *Alerter) AlertMinipoolStaked(minipoolAddress common.Address, succeeded bool) {
	if a.enabled {
		return
	}

	if !a.cfg.Alertmanager.AlertEnabled_MinipoolStaked.Value {
		return
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
	a.sendAlert(alert)
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

func (a *Alerter) AlertExecutionClientSyncComplete() {
	if !a.cfg.Alertmanager.AlertEnabled_ExecutionClientSyncComplete.Value {
		return
	}
	a.alertClientSyncComplete(ClientKindExecution)
}

func (a *Alerter) AlertBeaconClientSyncComplete() {
	if !a.cfg.Alertmanager.AlertEnabled_BeaconClientSyncComplete.Value {
		return
	}
	a.alertClientSyncComplete(ClientKindBeacon)
}

type ClientKind string

const (
	ClientKindExecution ClientKind = "Execution"
	ClientKindBeacon    ClientKind = "Beacon"
)

func (a *Alerter) alertClientSyncComplete(client ClientKind) {
	alertName := fmt.Sprintf("%sClientSyncComplete", client)
	if !a.enabled {
		return
	}

	alert := createAlert(
		alertName,
		fmt.Sprintf("%s Client Sync Complete", client),
		fmt.Sprintf("The %s client has completed syncing.", client),
		SeverityInfo,
		strfmt.DateTime(time.Now().Add(time.Minute*1)),
		nil,
	)
	a.sendAlert(alert)
}

func (a *Alerter) sendAlert(alert *models.PostableAlert) {
	a.log.Info("sending alert", "label", alert.Labels["alertname"], "summarry", alert.Annotations["summary"])

	params := apialert.NewPostAlertsParams().WithDefaults().WithAlerts(models.PostableAlerts{alert})
	client := a.alertmanagerClient
	_, err := client.Alert.PostAlerts(params)

	if err != nil {
		a.log.Error("error posting alert", log.Err(err))
	}
}

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

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

func createClient(cfg *config.SmartNodeConfig) *apiclient.Alertmanager {
	// use the alertmanager container name for the hostname
	host := fmt.Sprintf("%s:%d", string(config.ContainerID_Alertmanager), cfg.Alertmanager.Port.Value)

	if cfg.IsNativeMode {
		host = fmt.Sprintf("%s:%d", cfg.Alertmanager.NativeModeHost.Value, cfg.Alertmanager.NativeModePort.Value)
	}

	transport := apiclient.DefaultTransportConfig().WithHost(host)
	client := apiclient.NewHTTPClientWithConfig(strfmt.Default, transport)
	return client
}
