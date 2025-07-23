package config

import (
	"fmt"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool/template"
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const alertmanagerTag string = "prom/alertmanager:v0.26.0"
const AlertmanagerConfigTemplate string = "alerting/alertmanager.tmpl"
const AlertmanagerConfigFile string = "alerting/alertmanager.yml"

// Note: Alerting rules are actually loaded by prometheus, but we control the alerting settings here.
const AlertingRulesConfigTemplate string = "alerting/rules/default.tmpl"
const AlertingRulesConfigFile string = "alerting/rules/default.yml"

// Defaults
const defaultAlertmanagerPort uint16 = 9093
const defaultAlertmanagerHost string = "localhost"
const defaultAlertmanagerOpenPort config.RPCMode = config.RPC_Closed

// Configuration for Alertmanager
type AlertmanagerConfig struct {

	// The parent Rocket Pool Config
	Parent *RocketPoolConfig `yaml:"-"`

	Title string `yaml:"-"`

	// Whether alerting is enabled
	EnableAlerting config.Parameter `yaml:"enableAlerting,omitempty"`

	// Port for alertmanager UI & API
	Port config.Parameter `yaml:"port,omitempty"`

	// Host for alertmanager UI & API. ONLY USED IN NATIVE MODE. In Docker, the host is derived from the container name.
	NativeModeHost config.Parameter `yaml:"nativeModeHost,omitempty"`
	// Port for alertmanager UI & API. ONLY USED IN NATIVE MODE. In Docker, the host is derived from the container.
	NativeModePort config.Parameter `yaml:"nativeModePort,omitempty"`

	// Toggle for forwarding the API port outside of Docker; useful for ability to silence alerts
	OpenPort config.Parameter `yaml:"openPort,omitempty"`

	// The Docker Hub tag for Alertmanager
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// The Discord webhook URL for alert notifications
	DiscordWebhookURL config.Parameter `yaml:"discordWebhookURL,omitempty"`

	// The Pushover Token for alert notifications
	PushoverToken config.Parameter `yaml:"pushoverToken,omitempty"`
	// The Pushover User Key for alert notifications
	PushoverUserKey config.Parameter `yaml:"pushoverUserKey,omitempty"`

	// Alerts configured in prometheus rule configuration file:
	AlertEnabled_ClientSyncStatusBeacon    config.Parameter `yaml:"alertEnabled_ClientSyncStatusBeacon,omitempty"`
	AlertEnabled_ClientSyncStatusExecution config.Parameter `yaml:"alertEnabled_ClientSyncStatusExecution,omitempty"`
	AlertEnabled_UpcomingSyncCommittee     config.Parameter `yaml:"alertEnabled_UpcomingSyncCommittee,omitempty"`
	AlertEnabled_ActiveSyncCommittee       config.Parameter `yaml:"alertEnabled_ActiveSyncCommittee,omitempty"`
	AlertEnabled_UpcomingProposal          config.Parameter `yaml:"alertEnabled_UpcomingProposal,omitempty"`
	AlertEnabled_RecentProposal            config.Parameter `yaml:"alertEnabled_RecentProposal,omitempty"`
	AlertEnabled_LowDiskSpaceWarning       config.Parameter `yaml:"alertEnabled_LowDiskSpaceWarning,omitempty"`
	AlertEnabled_LowDiskSpaceCritical      config.Parameter `yaml:"alertEnabled_LowDiskSpaceCritical,omitempty"`
	AlertEnabled_OSUpdatesAvailable        config.Parameter `yaml:"alertEnabled_OSUpdatesAvailable,omitempty"`
	AlertEnabled_RPUpdatesAvailable        config.Parameter `yaml:"alertEnabled_RPUpdatesAvailable,omitempty"`
	// Alerts manually sent in alerting.go:
	AlertEnabled_FeeRecipientChanged         config.Parameter `yaml:"alertEnabled_FeeRecipientChanged,omitempty"`
	AlertEnabled_MinipoolBondReduced         config.Parameter `yaml:"alertEnabled_MinipoolBondReduced,omitempty"`
	AlertEnabled_MinipoolBalanceDistributed  config.Parameter `yaml:"alertEnabled_MinipoolBalanceDistributed,omitempty"`
	AlertEnabled_MinipoolPromoted            config.Parameter `yaml:"alertEnabled_MinipoolPromoted,omitempty"`
	AlertEnabled_MinipoolStaked              config.Parameter `yaml:"alertEnabled_MinipoolStaked,omitempty"`
	AlertEnabled_ExecutionClientSyncComplete config.Parameter `yaml:"alertEnabled_ExecutionClientSyncComplete,omitempty"`
	AlertEnabled_BeaconClientSyncComplete    config.Parameter `yaml:"alertEnabled_BeaconClientSyncComplete,omitempty"`
}

func NewAlertmanagerConfig(cfg *RocketPoolConfig) *AlertmanagerConfig {

	return &AlertmanagerConfig{
		Parent: cfg,

		Title: "Alertmanager Settings",

		EnableAlerting: config.Parameter{
			ID:                 "enableAlerting",
			Name:               "Enable Alerting",
			Description:        "Enable the Smartnode's alerting system. This will provide you alerts when important events occur with your node.",
			Type:               config.ParameterType_Bool,
			Default:            map[config.Network]interface{}{config.Network_All: true},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node, config.ContainerID_Prometheus, config.ContainerID_Alertmanager},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		Port: config.Parameter{
			ID:                 "port",
			Name:               "Alertmanager Port",
			Description:        "The port Alertmanager will listen on.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultAlertmanagerPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node, config.ContainerID_Alertmanager, config.ContainerID_Prometheus},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		NativeModeHost: config.Parameter{
			ID:                 "nativeModeHost",
			Name:               "Alertmanager Host",
			Description:        "The host that the node should use to communicate with Alertmanager.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: defaultAlertmanagerHost},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node, config.ContainerID_Prometheus},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		NativeModePort: config.Parameter{
			ID:                 "nativeModePort",
			Name:               "Alertmanager Port",
			Description:        "The port that the node should use to communicate with Alertmanager.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultAlertmanagerPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node, config.ContainerID_Prometheus},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		OpenPort: config.Parameter{
			ID:                 "openPort",
			Name:               "Expose Alertmanager Port",
			Description:        "Expose the Alertmanager's port to other processes on your machine, or to your local network so other machines can access it too.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: defaultAlertmanagerOpenPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Alertmanager},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options:            config.PortModes(""),
		},

		ContainerTag: config.Parameter{
			ID:                 "containerTag",
			Name:               "Alertmanager Container Tag",
			Description:        "The tag name of the Alertmanager container you want to use on Docker Hub.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: alertmanagerTag},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Alertmanager},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		DiscordWebhookURL: config.Parameter{
			ID:                 "discordWebhookURL",
			Name:               "Alertmanager Discord Webhook URL",
			Description:        "Discord notifications are sent via the Discord webhook API. See Discord's 'Intro to Webhooks' article to learn how to configure a webhook integration for a channel at https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Alertmanager},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		PushoverToken: config.Parameter{
			ID:                 "pushoverToken",
			Name:               "Alertmanager Pushover Token",
			Description:        "Pushover notifications are sent via the Pushover API. See docs for detailed technical explanation or a tl;dr on how to configure at https://pushover.net/api",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Alertmanager},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		PushoverUserKey: config.Parameter{
			ID:                 "pushoverUserKey",
			Name:               "Alertmanager Pushover User Key",
			Description:        "Pushover notifications are sent via the Pushover API. See docs for detailed technical explanation or a tl;dr on how to configure at https://pushover.net/api",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Alertmanager},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		AlertEnabled_ClientSyncStatusBeacon: createParameterForAlertEnablement(
			"ClientSyncStatusBeacon",
			"beacon client is not synced"),

		AlertEnabled_ClientSyncStatusExecution: createParameterForAlertEnablement(
			"ClientSyncStatusExecution",
			"execution client is not synced"),

		AlertEnabled_UpcomingSyncCommittee: createParameterForAlertEnablement(
			"UpcomingSyncCommittee",
			"about to become part of a sync committee"),

		AlertEnabled_ActiveSyncCommittee: createParameterForAlertEnablement(
			"ActiveSyncCommittee",
			"part of a sync committee"),

		AlertEnabled_UpcomingProposal: createParameterForAlertEnablement(
			"UpcomingProposal",
			"about to propose a block"),

		AlertEnabled_RecentProposal: createParameterForAlertEnablement(
			"RecentProposal",
			"recently proposed a block"),

		AlertEnabled_LowDiskSpaceWarning: createParameterForAlertEnablement(
			"LowDiskSpaceWarning",
			"low disk space"),

		AlertEnabled_LowDiskSpaceCritical: createParameterForAlertEnablement(
			"LowDiskSpaceCritical",
			"critically low disk space"),

		AlertEnabled_OSUpdatesAvailable: createParameterForAlertEnablement(
			"OSUpdatesAvailable",
			"OS updates available"),

		AlertEnabled_RPUpdatesAvailable: createParameterForAlertEnablement(
			"RPUpdatesAvailable",
			"Smartnode Update Available"),

		AlertEnabled_FeeRecipientChanged: createParameterForAlertEnablement(
			"FeeRecipientChanged",
			"Fee Recipient Changed"),

		AlertEnabled_MinipoolBondReduced: createParameterForAlertEnablement(
			"MinipoolBondReduced",
			"Minipool Bond Reduced"),

		AlertEnabled_MinipoolBalanceDistributed: createParameterForAlertEnablement(
			"MinipoolBalanceDistributed",
			"Minipool Balance Distributed"),

		AlertEnabled_MinipoolPromoted: createParameterForAlertEnablement(
			"MinipoolPromoted",
			"Minipool Promoted"),

		AlertEnabled_MinipoolStaked: createParameterForAlertEnablement(
			"MinipoolStaked",
			"Minipool Staked"),

		AlertEnabled_ExecutionClientSyncComplete: createParameterForAlertEnablement(
			"ExecutionClientSyncComplete",
			"execution client is synced"),

		AlertEnabled_BeaconClientSyncComplete: createParameterForAlertEnablement(
			"BeaconClientSyncComplete",
			"beacon client is synced"),
	}
}

func createParameterForAlertEnablement(uniqueName string, label string) config.Parameter {
	titleCaser := cases.Title(language.Und, cases.NoLower)
	return config.Parameter{
		ID:                 fmt.Sprintf("alertEnabled_%s", uniqueName),
		Name:               fmt.Sprintf("Alert for %s", titleCaser.String(label)),
		Description:        fmt.Sprintf("Enable an alert when %s", label),
		Type:               config.ParameterType_Bool,
		Default:            map[config.Network]interface{}{config.Network_All: true},
		AffectsContainers:  []config.ContainerID{config.ContainerID_Prometheus},
		CanBeBlank:         false,
		OverwriteOnUpgrade: false,
	}
}

func (cfg *AlertmanagerConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.EnableAlerting,
		&cfg.Port,
		&cfg.OpenPort,
		&cfg.NativeModeHost,
		&cfg.NativeModePort,
		&cfg.DiscordWebhookURL,
		&cfg.PushoverToken,
		&cfg.PushoverUserKey,
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

func (cfg *AlertmanagerConfig) GetConfigTitle() string {
	return cfg.Title
}

// Used by text/template to format alertmanager.yml
func (cfg *AlertmanagerConfig) GetOpenPorts() string {
	portMode := cfg.OpenPort.Value.(config.RPCMode)
	if !portMode.Open() {
		return ""
	}
	return fmt.Sprintf("\"%s\"", portMode.DockerPortMapping(cfg.Port.Value.(uint16)))
}

// Load the alerting configuration templates, do the template variable substitutions, and save them.
func (cfg *AlertmanagerConfig) UpdateConfigurationFiles(configPath string) error {
	err := cfg.processTemplate(configPath, AlertmanagerConfigTemplate, AlertmanagerConfigFile, "{{", "}}")
	if err != nil {
		return fmt.Errorf("error processing alertmanager config template: %w", err)
	}
	// NOTE: we use unique delimiters here because there are nested go templates in the alert messages
	err = cfg.processTemplate(configPath, AlertingRulesConfigTemplate, AlertingRulesConfigFile, "{{{", "}}}")
	if err != nil {
		return fmt.Errorf("error processing alerting rules template: %w", err)
	}
	return nil
}

func (cfg *AlertmanagerConfig) processTemplate(configPath string, templateFileName string, configFileName string, leftDelim string, rightDelim string) error {
	templatePath, err := homedir.Expand(fmt.Sprintf("%s/%s", configPath, templateFileName))
	if err != nil {
		return fmt.Errorf("error expanding alerting template path for file %s: %w", templateFileName, err)
	}

	configFile, err := homedir.Expand(fmt.Sprintf("%s/%s", configPath, configFileName))
	if err != nil {
		return fmt.Errorf("error expanding alerting file out path for file %s: %w", configFileName, err)
	}

	t := template.Template{
		Src: templatePath,
		Dst: configFile,
	}

	return t.WriteWithDelims(cfg, leftDelim, rightDelim)
}
