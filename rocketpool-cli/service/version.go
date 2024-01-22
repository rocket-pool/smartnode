package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
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
	err = utils.PrintNetwork(cfg.GetNetwork(), isNew)
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
		fmt.Printf("Rocket Pool client version: %s\n", c.App.Version)
		fmt.Printf("Rocket Pool service version: %s\n", serviceVersion)
		fmt.Println("Configured for Native Mode")
		return nil
	}

	// Get the execution client string
	var eth1ClientString string
	eth1ClientMode := cfg.ExecutionClientMode.Value.(cfgtypes.Mode)
	switch eth1ClientMode {
	case cfgtypes.Mode_Local:
		eth1Client := cfg.ExecutionClient.Value.(cfgtypes.ExecutionClient)
		format := "%s (Locally managed)\n\tImage: %s"
		switch eth1Client {
		case cfgtypes.ExecutionClient_Geth:
			eth1ClientString = fmt.Sprintf(format, "Geth", cfg.Geth.ContainerTag.Value.(string))
		case cfgtypes.ExecutionClient_Nethermind:
			eth1ClientString = fmt.Sprintf(format, "Nethermind", cfg.Nethermind.ContainerTag.Value.(string))
		case cfgtypes.ExecutionClient_Besu:
			eth1ClientString = fmt.Sprintf(format, "Besu", cfg.Besu.ContainerTag.Value.(string))
		default:
			return fmt.Errorf("unknown local execution client [%v]", eth1Client)
		}

	case cfgtypes.Mode_External:
		eth1ClientString = "Externally managed"

	default:
		return fmt.Errorf("unknown execution client mode [%v]", eth1ClientMode)
	}

	// Get the consensus client string
	var eth2ClientString string
	eth2ClientMode := cfg.ConsensusClientMode.Value.(cfgtypes.Mode)
	switch eth2ClientMode {
	case cfgtypes.Mode_Local:
		eth2Client := cfg.ConsensusClient.Value.(cfgtypes.ConsensusClient)
		format := "%s (Locally managed)\n\tImage: %s"
		switch eth2Client {
		case cfgtypes.ConsensusClient_Lighthouse:
			eth2ClientString = fmt.Sprintf(format, "Lighthouse", cfg.Lighthouse.ContainerTag.Value.(string))
		case cfgtypes.ConsensusClient_Lodestar:
			eth2ClientString = fmt.Sprintf(format, "Lodestar", cfg.Lodestar.ContainerTag.Value.(string))
		case cfgtypes.ConsensusClient_Nimbus:
			eth2ClientString = fmt.Sprintf(format+"\n\tVC image: %s", "Nimbus", cfg.Nimbus.BnContainerTag.Value.(string), cfg.Nimbus.VcContainerTag.Value.(string))
		case cfgtypes.ConsensusClient_Prysm:
			eth2ClientString = fmt.Sprintf(format+"\n\tVC image: %s", "Prysm", cfg.Prysm.BnContainerTag.Value.(string), cfg.Prysm.VcContainerTag.Value.(string))
		case cfgtypes.ConsensusClient_Teku:
			eth2ClientString = fmt.Sprintf(format, "Teku", cfg.Teku.ContainerTag.Value.(string))
		default:
			return fmt.Errorf("unknown local consensus client [%v]", eth2Client)
		}

	case cfgtypes.Mode_External:
		eth2Client := cfg.ExternalConsensusClient.Value.(cfgtypes.ConsensusClient)
		format := "%s (Externally managed)\n\tVC Image: %s"
		switch eth2Client {
		case cfgtypes.ConsensusClient_Lighthouse:
			eth2ClientString = fmt.Sprintf(format, "Lighthouse", cfg.ExternalLighthouse.ContainerTag.Value.(string))
		case cfgtypes.ConsensusClient_Lodestar:
			eth2ClientString = fmt.Sprintf(format, "Lodestar", cfg.ExternalLodestar.ContainerTag.Value.(string))
		case cfgtypes.ConsensusClient_Nimbus:
			eth2ClientString = fmt.Sprintf(format, "Nimbus", cfg.ExternalNimbus.ContainerTag.Value.(string))
		case cfgtypes.ConsensusClient_Prysm:
			eth2ClientString = fmt.Sprintf(format, "Prysm", cfg.ExternalPrysm.ContainerTag.Value.(string))
		case cfgtypes.ConsensusClient_Teku:
			eth2ClientString = fmt.Sprintf(format, "Teku", cfg.ExternalTeku.ContainerTag.Value.(string))
		default:
			return fmt.Errorf("unknown external consensus client [%v]", eth2Client)
		}

	default:
		return fmt.Errorf("unknown consensus client mode [%v]", eth2ClientMode)
	}

	// Print version info
	fmt.Printf("Rocket Pool client version: %s\n", c.App.Version)
	fmt.Printf("Rocket Pool service version: %s\n", serviceVersion)
	fmt.Printf("Selected Execution client: %s\n", eth1ClientString)
	fmt.Printf("Selected Consensus client: %s\n", eth2ClientString)
	return nil
}
