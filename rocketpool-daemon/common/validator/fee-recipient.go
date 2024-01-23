package validator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/beacon"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/log"
	"github.com/rocket-pool/smartnode/shared/config"
)

// Settings
const ValidatorContainerSuffix = "_validator"
const BeaconContainerSuffix = "_eth2"

var validatorRestartTimeout, _ = time.ParseDuration("5s")

// Restart validator process
func RestartValidator(cfg *config.RocketPoolConfig, bc beacon.Client, log *log.ColorLogger, d *client.Client) error {

	// Restart validator container
	if !cfg.IsNativeMode {

		// Get validator container name & client type label
		var containerName string
		clientTypeLabel := "validator"
		if cfg.Smartnode.ProjectName.Value == "" {
			return errors.New("Rocket Pool docker project name not set")
		}

		// Log
		if log != nil {
			log.Printlnf("Restarting %s container (%s)...", clientTypeLabel, containerName)
		}

		// Get all containers
		containers, err := d.ContainerList(context.Background(), types.ContainerListOptions{All: true})
		if err != nil {
			return fmt.Errorf("Could not get docker containers: %w", err)
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
			return errors.New("Validator container not found")
		}

		// Restart validator container
		timeout := int(validatorRestartTimeout.Seconds())
		if err := d.ContainerRestart(context.Background(), validatorContainerId, container.StopOptions{Timeout: &timeout}); err != nil {
			return fmt.Errorf("Could not restart validator container: %w", err)
		}

		// Restart external validator process
	} else {

		// Get validator restart command
		restartCommand := os.ExpandEnv(cfg.Native.ValidatorRestartCommand.Value.(string))

		// Log
		if log != nil {
			log.Printlnf("Restarting validator process with command '%s'...", restartCommand)
		}

		// Run validator restart command bound to os stdout/stderr
		cmd := exec.Command(restartCommand)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Could not restart validator process: %w", err)
		}

	}

	// Log & return
	if log != nil {
		log.Println("Successfully restarted validator")
	}
	return nil

}

// Stops the validator process
func StopValidator(cfg *config.RocketPoolConfig, bc beacon.Client, log *log.ColorLogger, d *client.Client) error {

	// Stop validator container
	if !cfg.IsNativeMode {

		// Get validator container name & client type label
		containerName := cfg.Smartnode.ProjectName.Value.(string) + ValidatorContainerSuffix
		clientTypeLabel := "validator"
		if cfg.Smartnode.ProjectName.Value == "" {
			return errors.New("Rocket Pool docker project name not set")
		}

		// Log
		if log != nil {
			log.Printlnf("Stopping %s container (%s)...", clientTypeLabel, containerName)
		}

		// Get all containers
		containers, err := d.ContainerList(context.Background(), types.ContainerListOptions{All: true})
		if err != nil {
			return fmt.Errorf("Could not get docker containers: %w", err)
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
			// TODO: return here if the container doesn't exist? Is erroring out necessary?
			return fmt.Errorf("Validator container %s not found", containerName)
		}

		// Stop validator container
		if err := d.ContainerPause(context.Background(), validatorContainerId); err != nil {
			if strings.Contains(err.Error(), "is not running") {
				// Handle situations where the container is already stopped
				if log != nil {
					log.Printlnf("Validator container %s was not running.", containerName)
				}
				return nil
			}
			return fmt.Errorf("Could not stop validator container %s: %w", containerName, err)
		}

	} else {
		// Stop external validator process

		// Get validator stop command
		stopCommand := os.ExpandEnv(cfg.Native.ValidatorStopCommand.Value.(string))

		// Log
		if log != nil {
			log.Printlnf("Stopping validator process with command '%s'...", stopCommand)
		}

		// Run validator stop command bound to os stdout/stderr
		cmd := exec.Command(stopCommand)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Could not stop validator process: %w", err)
		}

	}

	// Log & return
	if log != nil {
		log.Println("Successfully stopped validator")
	}
	return nil

}
