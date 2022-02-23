package config

// Constants
const defaultPocketGatewayMainnet string = "lb/613bb4ae8c124d00353c40a1"
const defaultPocketGatewayPrater string = "lb/6126b4a783e49000343a3a47"
const pocketEventLogInterval int = 25000

// Configuration for Pocket
type PocketConfig struct {
	// Common parameters that Pocket doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"unsupportedCommonParams,omitempty"`

	// Compatible consensus clients
	CompatibleConsensusClients []ConsensusClient `yaml:"compatibleConsensusClients,omitempty"`

	// The max number of events to query in a single event log query
	EventLogInterval int `yaml:"eventLogInterval,omitempty"`

	// The Pocket gateway ID
	GatewayID Parameter `yaml:"gatewayID,omitempty"`
}

// Generates a new Pocket configuration
func NewPocketConfig(config *RocketPoolConfig) *PocketConfig {
	return &PocketConfig{
		UnsupportedCommonParams: []string{ecWsPortID},

		CompatibleConsensusClients: []ConsensusClient{
			ConsensusClient_Lighthouse,
			ConsensusClient_Prysm,
			ConsensusClient_Teku,
		},

		EventLogInterval: pocketEventLogInterval,

		GatewayID: Parameter{
			ID:          "gatewayID",
			Name:        "Gateway ID",
			Description: "If you would like to use a custom gateway for Pocket instead of the default Rocket Pool gateway, enter it here.",
			Type:        ParameterType_String,
			Default: map[Network]interface{}{
				Network_Mainnet: defaultPocketGatewayMainnet,
				Network_Prater:  defaultPocketGatewayPrater,
			},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{"POCKET_GATEWAY_ID"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (config *PocketConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.GatewayID,
	}
}
