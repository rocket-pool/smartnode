package apcupsd

import (
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const (
	containerTag string = "rocketpool/graffiti-wall-addon:v1.0.1"
)

// Configuration for the Graffiti Wall Writer
type ApcupsdConfig struct {
	Title string `yaml:"-"`

	Enabled config.Parameter `yaml:"enabled,omitempty"`
}

// Creates a new configuration instance
func NewConfig() *ApcupsdConfig {
	return &ApcupsdConfig{
		Title: "APCUPSD Settings",

		Enabled: config.Parameter{
			ID:                   "enabled",
			Name:                 "Enabled",
			Description:          "Enable APCUPSD monitoring",
			Type:                 config.ParameterType_Bool,
			Default:              map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:    []config.ContainerID{ContainerID_Apcupsd, config.ContainerID_Validator},
			EnvironmentVariables: []string{"ADDON_APCUPSD_ENABLED"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (cfg *ApcupsdConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.Enabled,
	}
}

// The the title for the config
func (cfg *ApcupsdConfig) GetConfigTitle() string {
	return cfg.Title
}
