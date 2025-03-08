package rescue_node

import (
	"github.com/rocket-pool/smartnode/shared/types/config"
)

type RescueNodeConfig struct {
	Title string `yaml:"-"`

	Enabled config.Parameter `yaml:"enabled,omitempty"`

	Username config.Parameter `yaml:"username,omitempty"`
	Password config.Parameter `yaml:"username,omitempty"`
}

// Creates a new configuration instance
func NewConfig() *RescueNodeConfig {
	return &RescueNodeConfig{
		Title: "Rescue Node Settings",

		Enabled: config.Parameter{
			ID:                 "enabled",
			Name:               "Enabled",
			Description:        "Enable the Rescue Node\n\nVisit rescuenode.com for more information, or to get a username and password.",
			Type:               config.ParameterType_Bool,
			Default:            map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		Username: config.Parameter{
			ID:                 "username",
			Name:               "Username",
			Description:        "Username from rescuenode.com.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		Password: config.Parameter{
			ID:                 "password",
			Name:               "Password",
			Description:        "Password from rescuenode.com.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Get the parameters for this config
func (cfg *RescueNodeConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.Enabled,
		&cfg.Username,
		&cfg.Password,
	}
}

// The title for the config
func (cfg *RescueNodeConfig) GetConfigTitle() string {
	return cfg.Title
}
