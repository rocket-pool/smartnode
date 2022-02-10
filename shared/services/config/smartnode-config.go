package config

import "github.com/rocket-pool/smartnode/shared"

// Constants
const smartnodeTag string = "rocketpool/smartnode:" + shared.RocketPoolVersion
const powProxyTag string = "rocketpool/smartnode-pow-proxy:" + shared.RocketPoolVersion
const pruneProvisionerTag string = "rocketpool/eth1-prune-provision:v0.0.1"

// Defaults
const defaultProjectName string = "rocketpool"

// Configuration for the Smartnode
type SmartnodeConfig struct {
	// Docker container prefix
	ProjectName Parameter

	// The path of the data folder where everything is stored
	DataPath Parameter

	// The command for restarting the validator container in native mode
	ValidatorRestartCommand Parameter

	// Which network we're on
	Network Parameter

	// Manual max fee override
	ManualMaxFee Parameter

	// Manual priority fee override
	PriorityFee Parameter

	// Threshold for auto RPL claims
	RplClaimGasThreshold Parameter

	// Threshold for auto minipool stakes
	MinipoolStakeGasThreshold Parameter
}

// Generates a new Smartnode configuration
func NewSmartnodeConfig(config *MasterConfig) *SmartnodeConfig {

	return &SmartnodeConfig{
		ProjectName: Parameter{
			ID:                   "projectName",
			Name:                 "Project Name",
			Description:          "This is the prefix that will be attached to all of the Docker containers managed by the Smartnode.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: defaultProjectName},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth1, ContainerID_Eth2, ContainerID_Validator, ContainerID_Grafana, ContainerID_Prometheus, ContainerID_Exporter},
			EnvironmentVariables: []string{"COMPOSE_PROJECT_NAME"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		DataPath: Parameter{
			ID:                   "passwordPath",
			Name:                 "Password Path",
			Description:          "The absolute path of the `data` folder that contains your node wallet's encrypted file, the password for your node wallet, and all of the validator keys for your minipools. You may use environment variables in this string.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: "$HOME/.rocketpool/data"},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Validator},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ValidatorRestartCommand: Parameter{
			ID:                   "validatorRestartCommand",
			Name:                 "Validator Restart Command",
			Description:          "The absolute path to a custom script that will be invoked when Rocket Pool needs to restart your validator container to load the new key after a minipool is staked. **For Native mode only.**",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: "$HOME/.rocketpool/chains/eth2/restart-validator.sh"},
			AffectsContainers:    []ContainerID{ContainerID_Node},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		Network: Parameter{
			ID:                   "network",
			Name:                 "Network",
			Description:          "The Ethereum network you want to use - select Prater Testnet to practice with fake ETH, or Mainnet to stake on the real network using real ETH.",
			Type:                 ParameterType_Choice,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth1, ContainerID_Eth2, ContainerID_Validator},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []ParameterOption{{
				ID:          "mainnet",
				Name:        "Ethereum Mainnet",
				Description: "This is the real Ethereum main network, using real ETH and real RPL to make real validators.",
				Value:       Network_Mainnet,
			}, {
				ID:          "prater",
				Name:        "Prater Testnet",
				Description: "This is the Prater test network, using free fake ETH and free fake RPL to make fake validators.\nUse this if you want to practice running the Smartnode in a free, safe environment before moving to mainnet.",
				Value:       Network_Prater,
			}},
		},

		ManualMaxFee: Parameter{
			ID:                   "manualMaxFee",
			Name:                 "Manual Max Fee",
			Description:          "Set this if you want all of the Smartnode's transactions to use this specific max fee value (in gwei), which is the most you'd be willing to pay (*including the priority fee*). This will ignore the recommended max fee based on the current network conditions, and explicitly use this value instead. This applies to automated transactions (such as claiming RPL and staking minipools) as well.",
			Type:                 ParameterType_Uint,
			Default:              map[Network]interface{}{Network_All: 0},
			AffectsContainers:    []ContainerID{ContainerID_Node, ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		PriorityFee: Parameter{
			ID:                   "priorityFee",
			Name:                 "Priority Fee",
			Description:          "The default value for the priority fee (in gwei) for all of your transactions. This describes how much you're willing to pay *above the network's current base fee* - the higher this is, the more ETH you give to the miners for including your transaction, which generally means it will be mined faster (as long as your max fee is sufficiently high to cover the current network conditions).",
			Type:                 ParameterType_Uint,
			Default:              map[Network]interface{}{Network_All: 2},
			AffectsContainers:    []ContainerID{ContainerID_Node, ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		RplClaimGasThreshold: Parameter{
			ID:                   "rplClaimGasThreshold",
			Name:                 "RPL Claim Gas Threshold",
			Description:          "Automatic RPL rewards claims will use the `Rapid` suggestion from the gas estimator, based on current network conditions. This threshold is a limit (in gwei) you can put on that suggestion; your node will not try to claim RPL rewards automatically until the suggestion is below this limit.",
			Type:                 ParameterType_Uint,
			Default:              map[Network]interface{}{Network_All: 150},
			AffectsContainers:    []ContainerID{ContainerID_Node, ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		MinipoolStakeGasThreshold: Parameter{
			ID:   "minipoolStakeGasThreshold",
			Name: "Minipool Stake Gas Threshold",
			Description: "Once a newly created minipool passes the scrub check and is ready to perform its second 16 ETH deposit (the `stake` transaction), your node will try to do so automatically using the `Rapid` suggestion from the gas estimator as its max fee. This threshold is a limit (in gwei) you can put on that suggestion; your node will not `stake` the new minipool until the suggestion is below this limit.\n\n" +
				"Note that to ensure your minipool does not get dissolved, the node will ignore this limit and automatically execute the `stake` transaction at whatever the suggested fee happens to be once too much time has passed since its first deposit (currently 7 days).",
			Type:                 ParameterType_Uint,
			Default:              map[Network]interface{}{Network_All: 150},
			AffectsContainers:    []ContainerID{ContainerID_Node},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}

}

// Handle a network change on all of the parameters
func (config *SmartnodeConfig) changeNetwork(oldNetwork Network, newNetwork Network) {
	changeNetworkForParameter(&config.ProjectName, oldNetwork, newNetwork)
	changeNetworkForParameter(&config.DataPath, oldNetwork, newNetwork)
	changeNetworkForParameter(&config.ValidatorRestartCommand, oldNetwork, newNetwork)
	changeNetworkForParameter(&config.Network, oldNetwork, newNetwork)
	changeNetworkForParameter(&config.ManualMaxFee, oldNetwork, newNetwork)
	changeNetworkForParameter(&config.PriorityFee, oldNetwork, newNetwork)
	changeNetworkForParameter(&config.RplClaimGasThreshold, oldNetwork, newNetwork)
	changeNetworkForParameter(&config.MinipoolStakeGasThreshold, oldNetwork, newNetwork)
}
