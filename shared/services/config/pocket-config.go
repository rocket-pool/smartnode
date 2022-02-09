package config

// Constants
const defaultPocketGatewayMainnet string = "lb/613bb4ae8c124d00353c40a1"
const defaultPocketGatewayPrater string = "lb/6126b4a783e49000343a3a47"

// Configuration for Pocket
type PocketConfig struct {
	// The master configuration this belongs to
	MasterConfig *MasterConfig

	// Common parameters that Pocket doesn't support and should be hidden
	UnsupportedCommonParams []string

	// The Pocket gateway ID
	GatewayID *Parameter
}

// Generates a new Pocket configuration
func NewPocketConfig(config *MasterConfig) *PocketConfig {
	return &PocketConfig{
		MasterConfig: config,

		UnsupportedCommonParams: []string{ecWsPortID},

		GatewayID: &Parameter{
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
