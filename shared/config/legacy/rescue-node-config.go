package config

type RescueNodeConfig struct {
	Title string `yaml:"-"`

	Enabled Parameter `yaml:"enabled,omitempty"`

	Username Parameter `yaml:"username,omitempty"`
	Password Parameter `yaml:"username,omitempty"`
}

// Creates a new configuration instance
func NewConfig() *RescueNodeConfig {
	return &RescueNodeConfig{
		Title: "Rescue Node Settings",

		Enabled: Parameter{
			ID:                 "enabled",
			Name:               "Enabled",
			Description:        "Enable the Rescue Node\n\nVisit rescuenode.com for more information, or to get a username and password.",
			Type:               ParameterType_Bool,
			Default:            map[Network]interface{}{Network_All: false},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		Username: Parameter{
			ID:                 "username",
			Name:               "Username",
			Description:        "Username from rescuenode.com.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		Password: Parameter{
			ID:                 "password",
			Name:               "Password",
			Description:        "Password from rescuenode.com.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Get the parameters for this config
func (cfg *RescueNodeConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&cfg.Enabled,
		&cfg.Username,
		&cfg.Password,
	}
}

// The the title for the config
func (cfg *RescueNodeConfig) GetConfigTitle() string {
	return cfg.Title
}
