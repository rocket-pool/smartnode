package config

import (
	"fmt"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/rocket-pool/node-manager-core/config"
	nmc_ids "github.com/rocket-pool/node-manager-core/config/ids"
	"github.com/rocket-pool/smartnode/v2/shared/config/ids"
)

// Constants
const (
	alertmanagerTag string = "prom/alertmanager:v0.26.0"

	// Defaults
	defaultAlertmanagerPort     uint16             = 9093
	defaultAlertmanagerHost     string             = "localhost"
	defaultAlertmanagerOpenPort config.RpcPortMode = config.RpcPortMode_Closed
)

// Configuration for Alertmanager
type AlertmanagerConfig struct {
	// Whether alerting is enabled
	EnableAlerting config.Parameter[bool]

	// Port for alertmanager UI & API
	Port config.Parameter[uint16]

	// Host for alertmanager UI & API. ONLY USED IN NATIVE MODE. In Docker, the host is derived from the container name.
	NativeModeHost config.Parameter[string]

	// Port for alertmanager UI & API. ONLY USED IN NATIVE MODE. In Docker, the host is derived from the container.
	NativeModePort config.Parameter[uint16]

	// Toggle for forwarding the API port outside of Docker; useful for ability to silence alerts
	OpenPort config.Parameter[config.RpcPortMode]

	// The Docker Hub tag for Alertmanager
	ContainerTag config.Parameter[string]

	// The Discord webhook URL for alert notifications
	DiscordWebhookUrl config.Parameter[string]

	// Alerts configured in prometheus rule configuration file:
	AlertEnabled_ClientSyncStatusBeacon    config.Parameter[bool]
	AlertEnabled_ClientSyncStatusExecution config.Parameter[bool]
	AlertEnabled_UpcomingSyncCommittee     config.Parameter[bool]
	AlertEnabled_ActiveSyncCommittee       config.Parameter[bool]
	AlertEnabled_UpcomingProposal          config.Parameter[bool]
	AlertEnabled_RecentProposal            config.Parameter[bool]
	AlertEnabled_LowDiskSpaceWarning       config.Parameter[bool]
	AlertEnabled_LowDiskSpaceCritical      config.Parameter[bool]
	AlertEnabled_OSUpdatesAvailable        config.Parameter[bool]
	AlertEnabled_RPUpdatesAvailable        config.Parameter[bool]
	// Alerts manually sent in alerting.go:
	AlertEnabled_FeeRecipientChanged         config.Parameter[bool]
	AlertEnabled_MinipoolBondReduced         config.Parameter[bool]
	AlertEnabled_MinipoolBalanceDistributed  config.Parameter[bool]
	AlertEnabled_MinipoolPromoted            config.Parameter[bool]
	AlertEnabled_MinipoolStaked              config.Parameter[bool]
	AlertEnabled_ExecutionClientSyncComplete config.Parameter[bool]
	AlertEnabled_BeaconClientSyncComplete    config.Parameter[bool]
}

func NewAlertmanagerConfig() *AlertmanagerConfig {

	return &AlertmanagerConfig{
		EnableAlerting: config.Parameter[bool]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.AlertmanagerEnableAlertingID,
				Name:               "Enable Alerting",
				Description:        "Enable the Smart Node's alerting system. This will provide you alerts when important events occur with your node.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_Daemon, config.ContainerID_Prometheus, ContainerID_Alertmanager},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]bool{
				config.Network_All: true,
			},
		},

		Port: config.Parameter[uint16]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 nmc_ids.PortID,
				Name:               "Alertmanager Port",
				Description:        "The port Alertmanager will listen on.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_Daemon, ContainerID_Alertmanager, config.ContainerID_Prometheus},
				CanBeBlank:         true,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]uint16{
				config.Network_All: defaultAlertmanagerPort,
			},
		},

		NativeModeHost: config.Parameter[string]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.AlertmanagerNativeModeHostID,
				Name:               "Alertmanager Host",
				Description:        "The host that the node should use to communicate with Alertmanager.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_Daemon, config.ContainerID_Prometheus},
				CanBeBlank:         true,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]string{
				config.Network_All: defaultAlertmanagerHost,
			},
		},

		NativeModePort: config.Parameter[uint16]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.AlertmanagerNativeModePortID,
				Name:               "Alertmanager Port",
				Description:        "The port that the node should use to communicate with Alertmanager.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_Daemon, config.ContainerID_Prometheus},
				CanBeBlank:         true,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]uint16{
				config.Network_All: defaultAlertmanagerPort,
			},
		},

		OpenPort: config.Parameter[config.RpcPortMode]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 nmc_ids.OpenPortID,
				Name:               "Expose Alertmanager Port",
				Description:        "Expose the Alertmanager's port to other processes on your machine, or to your local network so other machines can access it too.",
				AffectsContainers:  []config.ContainerID{ContainerID_Alertmanager},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Options: config.GetPortModes(""),
			Default: map[config.Network]config.RpcPortMode{
				config.Network_All: defaultAlertmanagerOpenPort,
			},
		},

		ContainerTag: config.Parameter[string]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 nmc_ids.ContainerTagID,
				Name:               "Alertmanager Container Tag",
				Description:        "The tag name of the Alertmanager container you want to use on Docker Hub.",
				AffectsContainers:  []config.ContainerID{ContainerID_Alertmanager},
				CanBeBlank:         false,
				OverwriteOnUpgrade: true,
			},
			Default: map[config.Network]string{
				config.Network_All: alertmanagerTag,
			},
		},

		DiscordWebhookUrl: config.Parameter[string]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.AlertmanagerDiscordWebhookUrlID,
				Name:               "Alertmanager Discord Webhook URL",
				Description:        "Discord notifications are sent via the Discord webhook API. See Discord's 'Intro to Webhooks' article to learn how to configure a webhook integration for a channel at https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks",
				AffectsContainers:  []config.ContainerID{ContainerID_Alertmanager},
				CanBeBlank:         true,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]string{
				config.Network_All: "",
			},
		},

		AlertEnabled_ClientSyncStatusBeacon: createParameterForAlertEnablement(
			ids.AlertmanagerClientSyncStatusBeaconID,
			"beacon client is not synced"),

		AlertEnabled_ClientSyncStatusExecution: createParameterForAlertEnablement(
			ids.AlertmanagerClientSyncStatusExecutionID,
			"execution client is not synced"),

		AlertEnabled_UpcomingSyncCommittee: createParameterForAlertEnablement(
			ids.AlertmanagerUpcomingSyncCommitteeID,
			"about to become part of a sync committee"),

		AlertEnabled_ActiveSyncCommittee: createParameterForAlertEnablement(
			ids.AlertmanagerActiveSyncCommitteeID,
			"part of a sync committee"),

		AlertEnabled_UpcomingProposal: createParameterForAlertEnablement(
			ids.AlertmanagerUpcomingProposalID,
			"about to propose a block"),

		AlertEnabled_RecentProposal: createParameterForAlertEnablement(
			ids.AlertmanagerRecentProposalID,
			"recently proposed a block"),

		AlertEnabled_LowDiskSpaceWarning: createParameterForAlertEnablement(
			ids.AlertmanagerLowDiskSpaceWarningID,
			"low disk space"),

		AlertEnabled_LowDiskSpaceCritical: createParameterForAlertEnablement(
			ids.AlertmanagerLowDiskSpaceCriticalID,
			"critically low disk space"),

		AlertEnabled_OSUpdatesAvailable: createParameterForAlertEnablement(
			ids.AlertmanagerOSUpdatesAvailableID,
			"OS updates available"),

		AlertEnabled_RPUpdatesAvailable: createParameterForAlertEnablement(
			ids.AlertmanagerRPUpdatesAvailableID,
			"Smartnode Update Available"),

		AlertEnabled_FeeRecipientChanged: createParameterForAlertEnablement(
			ids.AlertmanagerFeeRecipientChangedID,
			"Fee Recipient Changed"),

		AlertEnabled_MinipoolBondReduced: createParameterForAlertEnablement(
			ids.AlertmanagerMinipoolBondReducedID,
			"Minipool Bond Reduced"),

		AlertEnabled_MinipoolBalanceDistributed: createParameterForAlertEnablement(
			ids.AlertmanagerMinipoolBalanceDistributedID,
			"Minipool Balance Distributed"),

		AlertEnabled_MinipoolPromoted: createParameterForAlertEnablement(
			ids.AlertmanagerMinipoolPromotedID,
			"Minipool Promoted"),

		AlertEnabled_MinipoolStaked: createParameterForAlertEnablement(
			ids.AlertmanagerMinipoolStakedID,
			"Minipool Staked"),

		AlertEnabled_ExecutionClientSyncComplete: createParameterForAlertEnablement(
			ids.AlertmanagerExecutionClientSyncCompleteID,
			"execution client is synced"),

		AlertEnabled_BeaconClientSyncComplete: createParameterForAlertEnablement(
			ids.AlertmanagerBeaconClientSyncCompleteID,
			"beacon client is synced"),
	}
}

// The title for the config
func (cfg *AlertmanagerConfig) GetTitle() string {
	return "Alertmanager Settings"
}

// Get the parameters for this config
func (cfg *AlertmanagerConfig) GetParameters() []config.IParameter {
	return []config.IParameter{
		&cfg.EnableAlerting,
		&cfg.Port,
		&cfg.OpenPort,
		&cfg.NativeModeHost,
		&cfg.NativeModePort,
		&cfg.DiscordWebhookUrl,
		&cfg.ContainerTag,
		&cfg.AlertEnabled_ClientSyncStatusBeacon,
		&cfg.AlertEnabled_ClientSyncStatusExecution,
		&cfg.AlertEnabled_UpcomingSyncCommittee,
		&cfg.AlertEnabled_ActiveSyncCommittee,
		&cfg.AlertEnabled_UpcomingProposal,
		&cfg.AlertEnabled_RecentProposal,
		&cfg.AlertEnabled_LowDiskSpaceWarning,
		&cfg.AlertEnabled_LowDiskSpaceCritical,
		&cfg.AlertEnabled_OSUpdatesAvailable,
		&cfg.AlertEnabled_RPUpdatesAvailable,
		&cfg.AlertEnabled_FeeRecipientChanged,
		&cfg.AlertEnabled_MinipoolBondReduced,
		&cfg.AlertEnabled_MinipoolBalanceDistributed,
		&cfg.AlertEnabled_MinipoolPromoted,
		&cfg.AlertEnabled_MinipoolStaked,
		&cfg.AlertEnabled_ExecutionClientSyncComplete,
		&cfg.AlertEnabled_BeaconClientSyncComplete,
	}
}

// Get the sections underneath this one
func (cfg *AlertmanagerConfig) GetSubconfigs() map[string]config.IConfigSection {
	return map[string]config.IConfigSection{}
}

// Used by text/template to format alertmanager.yml
func (cfg *AlertmanagerConfig) GetOpenPorts() string {
	portMode := cfg.OpenPort.Value
	if !portMode.IsOpen() {
		return ""
	}
	return fmt.Sprintf("\"%s\"", portMode.DockerPortMapping(cfg.Port.Value))
}

func createParameterForAlertEnablement(id string, label string) config.Parameter[bool] {
	titleCaser := cases.Title(language.Und, cases.NoLower)
	return config.Parameter[bool]{
		ParameterCommon: &config.ParameterCommon{
			ID:                 id,
			Name:               fmt.Sprintf("Alert for %s", titleCaser.String(label)),
			Description:        fmt.Sprintf("Enable an alert when %s", label),
			AffectsContainers:  []config.ContainerID{config.ContainerID_Prometheus},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},
		Default: map[config.Network]bool{
			config.Network_All: true,
		},
	}
}
