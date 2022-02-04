package config

// Configuration for Pocket
type PocketConfig struct {
	// Common parameters shared across clients
	CommonParams *ExecutionCommonParams

	// Common parameters that Pocket doesn't support and should be hidden
	UnsupportedCommonParams []string

	// The Pocket gateway ID
	GatewayID *Parameter
}

// Generates a new Pocket configuration
func NewPocketConfig(commonParams *ExecutionCommonParams) *PocketConfig {
	return &PocketConfig{
		CommonParams: commonParams,

		UnsupportedCommonParams: []string{ecWsPortID},

		GatewayID: &Parameter{
			ID:                   "gatewayID",
			Name:                 "Gateway ID",
			Description:          "If you would like to use a custom gateway for Pocket instead of the default Rocket Pool gateway, enter it here.",
			Type:                 ParameterType_String,
			Default:              "", // TODO: change based on which network is selected
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{"POCKET_GATEWAY_ID"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}
}
