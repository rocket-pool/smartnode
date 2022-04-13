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
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
)

// Settings
const ValidatorContainerSuffix = "_validator"
const BeaconContainerSuffix = "_eth2"

var validatorRestartTimeout, _ = time.ParseDuration("5s")

// Get the distributor address for the provided node wallet
func GetDistributorAddress(rp *rocketpool.RocketPool, nodeAddress common.Address) (*common.Address, error) {
	// Return nil if the merge contract update hasn't been deployed yet
	isMergeUpdateDeployed, err := rputils.IsMergeUpdateDeployed(rp)
	if err != nil {
		return nil, err
	}
	if !isMergeUpdateDeployed {
		return nil, nil
	}

	// Get the distributor address for the node
	distributor, err := node.GetDistributorAddress(rp, nodeAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting distributor address for node [%s]: %w", nodeAddress.Hex(), err)
	}

	return &distributor, nil
}

// Restart validator process
func RestartValidator(cfg *config.RocketPoolConfig, bc beacon.Client, log *log.ColorLogger, d *client.Client) error {

	// Restart validator container
	if !cfg.IsNativeMode {

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
