package validator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Settings
const ValidatorContainerSuffix = "_validator"
const BeaconContainerSuffix = "_eth2"

var validatorRestartTimeout, _ = time.ParseDuration("5s")

// Restart validator process
func RestartValidator(cfg *config.RocketPoolConfig, bc beacon.Client, log *log.ColorLogger, d *client.Client) error {

	// Restart validator container
	if isInsideContainer() {

		// Get validator container name & client type label
		var containerName string
		var clientTypeLabel string
		if cfg.Smartnode.ProjectName.Value == "" {
			return errors.New("Rocket Pool docker project name not set")
		}
		switch clientType := bc.GetClientType(); clientType {
		case beacon.SplitProcess:
			containerName = cfg.Smartnode.ProjectName.Value.(string) + ValidatorContainerSuffix
			clientTypeLabel = "validator"
		case beacon.SingleProcess:
			containerName = cfg.Smartnode.ProjectName.Value.(string) + BeaconContainerSuffix
			clientTypeLabel = "beacon"
		default:
			return fmt.Errorf("Can't restart the validator, unknown client type '%d'", clientType)
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
		if err := d.ContainerRestart(context.Background(), validatorContainerId, &validatorRestartTimeout); err != nil {
			return fmt.Errorf("Could not restart validator container: %w", err)
		}

		// Restart external validator process
	} else {

		// Get validator restart command
		restartCommand := os.ExpandEnv(cfg.Smartnode.ValidatorRestartCommand.Value.(string))

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
	if isInsideContainer() {

		// Get validator container name & client type label
		var containerName string
		var clientTypeLabel string
		if cfg.Smartnode.ProjectName.Value == "" {
			return errors.New("Rocket Pool docker project name not set")
		}
		switch clientType := bc.GetClientType(); clientType {
		case beacon.SplitProcess:
			containerName = cfg.Smartnode.ProjectName.Value.(string) + ValidatorContainerSuffix
			clientTypeLabel = "validator"
		case beacon.SingleProcess:
			containerName = cfg.Smartnode.ProjectName.Value.(string) + BeaconContainerSuffix
			clientTypeLabel = "beacon"
		default:
			return fmt.Errorf("Can't stop the validator, unknown client type '%d'", clientType)
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
			return errors.New("Validator container not found")
		}

		// Stop validator container
		if err := d.ContainerPause(context.Background(), validatorContainerId); err != nil {
			return fmt.Errorf("Could not stop validator container: %w", err)
		}

	} else {
		// Stop external validator process

		// Get validator stop command
		stopCommand := os.ExpandEnv(cfg.Smartnode.ValidatorStopCommand.Value.(string))

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

// Check whether process is running inside a container
func isInsideContainer() bool {
	containerMarkerPaths := []string{
		"/.dockerenv",        // Docker
		"/run/.containerenv", // Podman
	}
	for _, path := range containerMarkerPaths {
		if pathExists(path) {
			return true
		}
	}
	return false
}

// Check if path exists
func pathExists(path string) bool {

	// Check for file info at path
	if _, err := os.Stat(path); err == nil {
		return true
	}

	// Assume that the path does not exist; this may result in false negatives (e.g. due to permissions)
	return false

}
