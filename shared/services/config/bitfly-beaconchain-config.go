package config

import (
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Defaults
const (
	defaultBitflyNodeMetricsSecret      string = ""
	defaultBitflyNodeMetricsEndpoint    string = "https://beaconcha.in/api/v1/client/metrics"
	defaultBitflyNodeMetricsMachineName string = "Smartnode"
)

// Configuration for Bitfly Node Metrics
type BitflyNodeMetricsConfig struct {
	Title string `yaml:"-"`

	Secret config.Parameter `yaml:"secret,omitempty"`

	Endpoint config.Parameter `yaml:"endpoint,omitempty"`

	MachineName config.Parameter `yaml:"machineName,omitempty"`
}

// Generates a new Bitfly Node Metrics config
func NewBitflyNodeMetricsConfig(cfg *RocketPoolConfig) *BitflyNodeMetricsConfig {
	return &BitflyNodeMetricsConfig{
		Title: "Bitfly Node Metrics Settings",

		Secret: config.Parameter{
			ID:                "bitflySecret",
			Name:              "Beaconcha.in API Key",
			Description:       "The API key used to authenticate your Beaconcha.in node metrics integration. Can be found in your Beaconcha.in account settings.\n\nPlease visit https://beaconcha.in/user/settings#api to access your account information.",
			Type:              config.ParameterType_String,
			Default:           map[config.Network]interface{}{config.Network_All: defaultBitflyNodeMetricsSecret},
			AffectsContainers: []config.ContainerID{config.ContainerID_Validator, config.ContainerID_Eth2},
			// ensures the string is 28 characters of Base64
			Regex:              "^[A-Za-z0-9+/]{28}$",
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		Endpoint: config.Parameter{
			ID:                 "bitflyEndpoint",
			Name:               "Node Metrics Endpoint",
			Description:        "The endpoint to send your Beaconcha.in Node Metrics data to. Should be left as the default.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: defaultBitflyNodeMetricsEndpoint},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Validator, config.ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		MachineName: config.Parameter{
			ID:                 "bitflyMachineName",
			Name:               "Node Metrics Machine Name",
			Description:        "The name of the machine you are running on. This is used to identify your machine in the mobile app.\nChange this if you are running multiple Smartnodes with the same Secret.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: defaultBitflyNodeMetricsMachineName},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Validator, config.ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},
	}
}

// Get the parameters for this config
func (cfg *BitflyNodeMetricsConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.Secret,
		&cfg.Endpoint,
		&cfg.MachineName,
	}
}

// The the title for the config
func (cfg *BitflyNodeMetricsConfig) GetConfigTitle() string {
	return cfg.Title
}
