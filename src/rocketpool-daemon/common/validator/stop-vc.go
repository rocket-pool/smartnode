package validator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/utils/log"
	"github.com/rocket-pool/smartnode/shared/config"
)

const (
	validatorStopTimeout time.Duration = time.Second * 5
)

// Stop or restart the validator client
func StopValidator(cfg *config.SmartNodeConfig, bc beacon.IBeaconClient, log *log.ColorLogger, d *client.Client, restart bool) error {
	// Restart validator container
	if !cfg.IsNativeMode {
		// Get validator container name & client type label
		containerName := cfg.GetDockerArtifactName(config.ValidatorClientSuffix)

		// Log
		if log != nil {
			if restart {
				log.Printlnf("Restarting validator client (%s)...", containerName)
			} else {
				log.Printlnf("Stopping validator client (%s)...", containerName)
			}
		}

		// Get all containers
		containers, err := d.ContainerList(context.Background(), types.ContainerListOptions{All: true})
		if err != nil {
			return fmt.Errorf("error getting docker containers: %w", err)
		}

		// Get validator container ID
		var validatorContainerId string
		for _, container := range containers {
			if container.Names[0] == "/"+containerName {
				validatorContainerId = container.ID
				break
			}
		}
		if validatorContainerId == "" {
			return errors.New("validator client container not found")
		}

		// Stop / restart validator container
		timeout := int(validatorStopTimeout.Seconds())
		if restart {
			err = d.ContainerRestart(context.Background(), validatorContainerId, container.StopOptions{Timeout: &timeout})
		} else {
			err = d.ContainerStop(context.Background(), validatorContainerId, container.StopOptions{Timeout: &timeout})
		}
		if err != nil {
			if restart {
				return fmt.Errorf("error restarting validator client container: %w", err)
			} else {
				return fmt.Errorf("error stopping validator client container: %w", err)
			}
		}
	} else {
		// Get validator control command
		var command string
		if restart {
			command = os.ExpandEnv(cfg.ValidatorClient.NativeValidatorRestartCommand.Value)
			if log != nil {
				log.Printlnf("Restarting validator client process with command '%s'...", command)
			}
		} else {
			command = os.ExpandEnv(cfg.ValidatorClient.NativeValidatorStopCommand.Value)
			if log != nil {
				log.Printlnf("Stopping validator client process with command '%s'...", command)
			}
		}

		// Run validator control command bound to os stdout/stderr
		cmd := exec.Command(command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			if restart {
				return fmt.Errorf("error restarting validator client process: %w", err)
			} else {
				return fmt.Errorf("error stopping validator client process: %w", err)
			}
		}
	}

	// Log & return
	if log != nil {
		if restart {
			log.Println("Successfully restarted validator client.")
		} else {
			log.Println("Successfully stopped validator client.")
		}
	}
	return nil
}
