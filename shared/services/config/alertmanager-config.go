package config

import (
	"fmt"

	"github.com/mitchellh/go-homedir"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool/template"
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const alertmanagerTag string = "prom/alertmanager:v0.26.0"

const AlertmanagerConfigTemplate string = "alertmanager.tmpl"
const AlertmanagerConfigFile string = "alertmanager.yml"

// Defaults
const defaultAlertmanagerPort uint16 = 9093
const defaultAlertmanagerOpenPort config.RPCMode = config.RPC_Closed

// Configuration for Alertmanager
type AlertmanagerConfig struct {
	Title string `yaml:"-"`

	// Port for alertmanager UI & API
	Port config.Parameter `yaml:"port,omitempty"`

	// Toggle for forwarding the API port outside of Docker; useful for ability to silence alerts
	OpenPort config.Parameter `yaml:"openPort,omitempty"`

	// The Docker Hub tag for Alertmanager
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// The Discord webhook URL for alert notifications
	DiscordWebhookURL config.Parameter `yaml:"discordWebhookURL,omitempty"`
}

func NewAlertmanagerConfig(cfg *RocketPoolConfig) *AlertmanagerConfig {

	return &AlertmanagerConfig{
		Title: "Alertmanager Settings",

		Port: config.Parameter{
			ID:                 "port",
			Name:               "Alertmanager Port",
			Description:        "The port Alertmanager will listen on.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultAlertmanagerPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Alertmanager, config.ContainerID_Prometheus},
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
	}
}

func (cfg *AlertmanagerConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.Port,
		&cfg.OpenPort,
		&cfg.ContainerTag,
		&cfg.DiscordWebhookURL,
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

// Load the Alertmanager template, do an template variable substitution, and save it
func (cfg *AlertmanagerConfig) UpdateConfigurationFile(configPath string) error {
	templatePath, err := homedir.Expand(fmt.Sprintf("%s/alerting/%s", configPath, AlertmanagerConfigTemplate))
	if err != nil {
		return fmt.Errorf("error expanding Alertmanager template path: %w", err)
	}

	configFile, err := homedir.Expand(fmt.Sprintf("%s/alerting/%s", configPath, AlertmanagerConfigFile))
	if err != nil {
		return fmt.Errorf("error expanding Alertmanager config file path: %w", err)
	}

	t := template.Template{
		Src: templatePath,
		Dst: configFile,
	}

	return t.Write(cfg)
}
