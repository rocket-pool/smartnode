package rescue_node

import (
	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/smartnode/v2/addons/rescue_node/ids"
)

type RescueNodeConfig struct {
	Enabled  config.Parameter[bool]
	Username config.Parameter[string]
	Password config.Parameter[string]
}

// Creates a new configuration instance
func NewConfig() *RescueNodeConfig {
	return &RescueNodeConfig{
		Enabled: config.Parameter[bool]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.RescueNodeEnabledID,
				Name:               "Enabled",
				Description:        "Enable the Rescue Node\n\nVisit rescuenode.com for more information, or to get a username and password.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]bool{
				config.Network_All: false,
			},
		},

		Username: config.Parameter[string]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.RescueNodeUsernameID,
				Name:               "Username",
				Description:        "Username from rescuenode.com.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_ValidatorClient},
				CanBeBlank:         true,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]string{
				config.Network_All: "",
			},
		},

		Password: config.Parameter[string]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.RescueNodePasswordID,
				Name:               "Password",
				Description:        "Password from rescuenode.com.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_ValidatorClient},
				CanBeBlank:         true,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]string{
				config.Network_All: "",
			},
		},
	}
}

// Get the title for the config
func (cfg *RescueNodeConfig) GetTitle() string {
	return "Rescue Node"
}

// Get the parameters for this config
func (cfg *RescueNodeConfig) GetParameters() []config.IParameter {
	return []config.IParameter{
		&cfg.Enabled,
		&cfg.Username,
		&cfg.Password,
	}
}

// Get the sections underneath this one
func (cfg *RescueNodeConfig) GetSubconfigs() map[string]config.IConfigSection {
	return map[string]config.IConfigSection{}
}
