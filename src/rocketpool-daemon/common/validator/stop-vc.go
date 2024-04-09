package validator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

const (
	validatorStopTimeout time.Duration = time.Second * 5
)

// Stop or restart the validator client
func StopValidator(ctx context.Context, cfg *config.SmartNodeConfig, bc beacon.IBeaconClient, d *client.Client, restart bool) error {
	// Get the loggerger
	logger, exists := log.FromContext(ctx)
	if !exists {
		panic("context didn't have a loggerger!")
	}

	// Restart validator container
	if !cfg.IsNativeMode {
		// Get validator container name & client type label
		containerName := cfg.GetDockerArtifactName(config.ValidatorClientSuffix)

		// logger
		if logger != nil {
			if restart {
				logger.Info("Restarting validator client...", slog.String(keys.ContainerKey, containerName))
			} else {
				logger.Info("Stopping validator client...", slog.String(keys.ContainerKey, containerName))
			}
		}

		// Get all containers
		containers, err := d.ContainerList(context.Background(), container.ListOptions{All: true})
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
			if logger != nil {
				logger.Info("Restarting validator client process...", slog.String(keys.CommandKey, command))
			}
		} else {
			command = os.ExpandEnv(cfg.ValidatorClient.NativeValidatorStopCommand.Value)
			if logger != nil {
				logger.Info("Stopping validator client process...", slog.String(keys.CommandKey, command))
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

	// logger & return
	if logger != nil {
		if restart {
			logger.Info("Successfully restarted validator client.")
		} else {
			logger.Info("Successfully stopped validator client.")
		}
	}
	return nil
}
