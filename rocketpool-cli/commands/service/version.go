package service

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/urfave/cli/v2"
)

// View the Rocket Pool service version information
func serviceVersion(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get the config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading configuration: %w", err)
	}

	// Print what network we're on
	err = utils.PrintNetwork(cfg.Network.Value, isNew)
	if err != nil {
		return err
	}

	// Get RP service version
	serviceVersion, err := rp.GetServiceVersion()
	if err != nil {
		return err
	}

	// Handle native mode
	if cfg.IsNativeMode {
		fmt.Printf("Smart Node client version: %s\n", c.App.Version)
		fmt.Printf("Smart Node service version: %s\n", serviceVersion)
		fmt.Println("Configured for Native Mode")
		return nil
	}

	// Get the execution client string
	var executionClientString string
	var beaconNodeString string
	clientMode := cfg.ClientMode.Value
	switch clientMode {
	case config.ClientMode_Local:
		// Execution client
		ec := cfg.LocalExecutionClient.ExecutionClient.Value
		ecFormat := "%s (Locally managed)\n\tImage: %s"
		switch ec {
		case config.ExecutionClient_Geth:
			executionClientString = fmt.Sprintf(ecFormat, "Geth", cfg.LocalExecutionClient.Geth.ContainerTag.Value)
		case config.ExecutionClient_Nethermind:
			executionClientString = fmt.Sprintf(ecFormat, "Nethermind", cfg.LocalExecutionClient.Nethermind.ContainerTag.Value)
		case config.ExecutionClient_Besu:
			executionClientString = fmt.Sprintf(ecFormat, "Besu", cfg.LocalExecutionClient.Besu.ContainerTag.Value)
		case config.ExecutionClient_Reth:
			executionClientString = fmt.Sprintf(ecFormat, "Reth", cfg.LocalExecutionClient.Reth.ContainerTag.Value)
		default:
			return fmt.Errorf("unknown local execution client [%v]", ec)
		}

		// Beacon node
		bn := cfg.LocalBeaconClient.BeaconNode.Value
		bnFormat := "%s (Locally managed)\n\tBN Image: %s\n\tVC image: %s"
		switch bn {
		case config.BeaconNode_Lighthouse:
			beaconNodeString = fmt.Sprintf(bnFormat, "Lighthouse", cfg.LocalBeaconClient.Lighthouse.ContainerTag.Value, cfg.ValidatorClient.Lighthouse.ContainerTag.Value)
		case config.BeaconNode_Lodestar:
			beaconNodeString = fmt.Sprintf(bnFormat, "Lodestar", cfg.LocalBeaconClient.Lodestar.ContainerTag.Value, cfg.ValidatorClient.Lodestar.ContainerTag.Value)
		case config.BeaconNode_Nimbus:
			beaconNodeString = fmt.Sprintf(bnFormat, "Nimbus", cfg.LocalBeaconClient.Nimbus.ContainerTag.Value, cfg.ValidatorClient.Nimbus.ContainerTag.Value)
		case config.BeaconNode_Prysm:
			beaconNodeString = fmt.Sprintf(bnFormat, "Prysm", cfg.LocalBeaconClient.Prysm.ContainerTag.Value, cfg.ValidatorClient.Prysm.ContainerTag.Value)
		case config.BeaconNode_Teku:
			beaconNodeString = fmt.Sprintf(bnFormat, "Teku", cfg.LocalBeaconClient.Teku.ContainerTag.Value, cfg.ValidatorClient.Teku.ContainerTag.Value)
		default:
			return fmt.Errorf("unknown local Beacon Node [%v]", bn)
		}

	case config.ClientMode_External:
		// Execution client
		ec := cfg.ExternalExecutionClient.ExecutionClient.Value
		ecFormat := "%s (Externally managed)"
		switch ec {
		case config.ExecutionClient_Geth:
			executionClientString = fmt.Sprintf(ecFormat, "Geth")
		case config.ExecutionClient_Nethermind:
			executionClientString = fmt.Sprintf(ecFormat, "Nethermind")
		case config.ExecutionClient_Besu:
			executionClientString = fmt.Sprintf(ecFormat, "Besu")
		case config.ExecutionClient_Reth:
			executionClientString = fmt.Sprintf(ecFormat, "Reth")
		default:
			return fmt.Errorf("unknown external Execution Client [%v]", ec)
		}

		// Beacon node
		bn := cfg.ExternalBeaconClient.BeaconNode.Value
		bnFormat := "%s (Externally managed)\n\tVC Image: %s"
		switch bn {
		case config.BeaconNode_Lighthouse:
			beaconNodeString = fmt.Sprintf(bnFormat, "Lighthouse", cfg.ValidatorClient.Lighthouse.ContainerTag.Value)
		case config.BeaconNode_Lodestar:
			beaconNodeString = fmt.Sprintf(bnFormat, "Lodestar", cfg.ValidatorClient.Lodestar.ContainerTag.Value)
		case config.BeaconNode_Nimbus:
			beaconNodeString = fmt.Sprintf(bnFormat, "Nimbus", cfg.ValidatorClient.Nimbus.ContainerTag.Value)
		case config.BeaconNode_Prysm:
			beaconNodeString = fmt.Sprintf(bnFormat, "Prysm", cfg.ValidatorClient.Prysm.ContainerTag.Value)
		case config.BeaconNode_Teku:
			beaconNodeString = fmt.Sprintf(bnFormat, "Teku", cfg.ValidatorClient.Teku.ContainerTag.Value)
		default:
			return fmt.Errorf("unknown external Beacon Node [%v]", bn)
		}

	default:
		return fmt.Errorf("unknown client mode [%v]", clientMode)
	}

	var mevBoostString string
	if cfg.MevBoost.Enable.Value {
		if cfg.MevBoost.Mode.Value == config.ClientMode_Local {
			mevBoostString = fmt.Sprintf("Enabled (Local Mode)\n\tImage: %s", cfg.MevBoost.ContainerTag.Value)
		} else {
			mevBoostString = "Enabled (External Mode)"
		}
	} else {
		mevBoostString = "Disabled"
	}

	// Print version info
	fmt.Printf("Smart Node client version: %s\n", c.App.Version)
	fmt.Printf("Smart Node service version: %s\n", serviceVersion)
	fmt.Printf("Selected Execution Client: %s\n", executionClientString)
	fmt.Printf("Selected Beacon Node: %s\n", beaconNodeString)
	fmt.Printf("MEV-Boost client: %s\n", mevBoostString)
	return nil
}
