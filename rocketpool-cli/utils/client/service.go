package client

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/a8m/envsubst"
	"github.com/alessio/shellescape"
	"github.com/blang/semver/v4"
	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/rocket-pool/smartnode/addons/graffiti_wall_writer"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/docker"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// Install the Rocket Pool service
func (c *Client) InstallService(verbose, noDeps bool, version, path string, dataPath string) error {
	// Get installation script flags
	flags := []string{
		"-v", shellescape.Quote(version),
	}
	if path != "" {
		flags = append(flags, fmt.Sprintf("-p %s", shellescape.Quote(path)))
	}
	if noDeps {
		flags = append(flags, "-d")
	}
	if dataPath != "" {
		flags = append(flags, fmt.Sprintf("-u %s", dataPath))
	}

	// Download the installation script
	resp, err := http.Get(fmt.Sprintf(InstallerURL, version))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http status downloading installation script: %d", resp.StatusCode)
	}

	// Sanity check that the script octet length matches content-length
	script, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if fmt.Sprint(len(script)) != resp.Header.Get("content-length") {
		return fmt.Errorf("downloaded script length %d did not match content-length header %s", len(script), resp.Header.Get("content-length"))
	}

	// Initialize installation command
	cmd := c.newCommand(fmt.Sprintf("sh -s -- %s", strings.Join(flags, " ")))

	// Pass the script to sh via its stdin fd
	cmd.SetStdin(bytes.NewReader(script))

	// Get command output pipes
	cmdOut, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmdErr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Print progress from stdout
	go (func() {
		scanner := bufio.NewScanner(cmdOut)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	})()

	// Read command & error output from stderr; render in verbose mode
	var errMessage string
	go (func() {
		c := color.New(DebugColor)
		scanner := bufio.NewScanner(cmdErr)
		for scanner.Scan() {
			errMessage = scanner.Text()
			if verbose {
				_, _ = c.Println(scanner.Text())
			}
		}
	})()

	// Run command and return error output
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Could not install Rocket Pool service: %s", errMessage)
	}
	return nil
}

// Install the update tracker
func (c *Client) InstallUpdateTracker(verbose bool, version string) error {
	// Get installation script flags
	flags := []string{
		"-v", fmt.Sprintf("%s", shellescape.Quote(version)),
	}

	// Download the installer package
	url := fmt.Sprintf(UpdateTrackerURL, version)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading installer package: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading installer package failed with code %d - [%s]", resp.StatusCode, resp.Status)
	}

	// Create a temp file for it
	path := filepath.Join(os.TempDir(), "install-update-tracker.sh")
	tempFile, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("error creating temporary file for installer script: %w", err)
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing installer package to disk: %w", err)
	}
	tempFile.Close()

	// Initialize installation command
	cmd := c.newCommand(fmt.Sprintf("sh -c \"%s %s\"", path, strings.Join(flags, " ")))
	cmdOut, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmdErr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Print progress from stdout
	go (func() {
		scanner := bufio.NewScanner(cmdOut)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	})()

	// Read command & error output from stderr; render in verbose mode
	var errMessage string
	go (func() {
		c := color.New(DebugColor)
		scanner := bufio.NewScanner(cmdErr)
		for scanner.Scan() {
			errMessage = scanner.Text()
			if verbose {
				_, _ = c.Println(scanner.Text())
			}
		}
	})()

	// Run command and return error output
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Could not install Rocket Pool update tracker: %s", errMessage)
	}
	return nil
}

// Start the Rocket Pool service
func (c *Client) StartService(composeFiles []string) error {
	/*
		// Start the API container first
		cmd, err := c.compose([]string{}, "up -d --quiet-pull")
		if err != nil {
			return fmt.Errorf("error creating compose command for API container: %w", err)
		}
		err = c.printOutput(cmd)
		if err != nil {
			return fmt.Errorf("error starting API container: %w", err)
		}
	*/
	// Start all of the containers
	cmd, err := c.compose(composeFiles, "up -d --remove-orphans --quiet-pull")
	if err != nil {
		return err
	}
	return c.printOutput(cmd)
}

// Pause the Rocket Pool service
func (c *Client) PauseService(composeFiles []string) error {
	cmd, err := c.compose(composeFiles, "stop")
	if err != nil {
		return err
	}
	return c.printOutput(cmd)
}

// Stop the Rocket Pool service
func (c *Client) StopService(composeFiles []string) error {
	cmd, err := c.compose(composeFiles, "down -v")
	if err != nil {
		return err
	}
	return c.printOutput(cmd)
}

// Stop the Rocket Pool service and remove the config folder
func (c *Client) TerminateService(composeFiles []string, configPath string) error {
	// Get the command to run with root privileges
	rootCmd, err := c.getEscalationCommand()
	if err != nil {
		return fmt.Errorf("could not get privilege escalation command: %w", err)
	}

	// Terminate the Docker containers
	cmd, err := c.compose(composeFiles, "down -v")
	if err != nil {
		return fmt.Errorf("error creating Docker artifact removal command: %w", err)
	}
	err = c.printOutput(cmd)
	if err != nil {
		return fmt.Errorf("error removing Docker artifacts: %w", err)
	}

	// Delete the RP directory
	path, err := homedir.Expand(configPath)
	if err != nil {
		return fmt.Errorf("error loading Rocket Pool directory: %w", err)
	}
	fmt.Printf("Deleting Rocket Pool directory (%s)...\n", path)
	cmd = fmt.Sprintf("%s rm -rf %s", rootCmd, path)
	_, err = c.readOutput(cmd)
	if err != nil {
		return fmt.Errorf("error deleting Rocket Pool directory: %w", err)
	}

	fmt.Println("Termination complete.")

	return nil
}

// Print the Rocket Pool service status
func (c *Client) PrintServiceStatus(composeFiles []string) error {
	cmd, err := c.compose(composeFiles, "ps")
	if err != nil {
		return err
	}
	return c.printOutput(cmd)
}

// Print the Rocket Pool service logs
func (c *Client) PrintServiceLogs(composeFiles []string, tail string, serviceNames ...string) error {
	sanitizedStrings := make([]string, len(serviceNames))
	for i, serviceName := range serviceNames {
		sanitizedStrings[i] = fmt.Sprintf("%s", shellescape.Quote(serviceName))
	}
	cmd, err := c.compose(composeFiles, fmt.Sprintf("logs -f --tail %s %s", shellescape.Quote(tail), strings.Join(sanitizedStrings, " ")))
	if err != nil {
		return err
	}
	return c.printOutput(cmd)
}

// Print the Rocket Pool service stats
func (c *Client) PrintServiceStats(composeFiles []string) error {
	// Get service container IDs
	cmd, err := c.compose(composeFiles, "ps -q")
	if err != nil {
		return err
	}
	containers, err := c.readOutput(cmd)
	if err != nil {
		return err
	}
	containerIds := strings.Split(strings.TrimSpace(string(containers)), "\n")

	// Print stats
	return c.printOutput(fmt.Sprintf("docker stats %s", strings.Join(containerIds, " ")))
}

// Print the Rocket Pool service compose config
func (c *Client) PrintServiceCompose(composeFiles []string) error {
	cmd, err := c.compose(composeFiles, "config")
	if err != nil {
		return err
	}
	return c.printOutput(cmd)
}

// Get the Rocket Pool service version
func (c *Client) GetServiceVersion() (string, error) {
	// Get service container version output
	var versionString string
	if c.daemonPath == "" {
		response, err := SendGetRequest[api.ServiceVersionData](c, "service/version", nil)
		if err != nil {
			return "", fmt.Errorf("error requesting Rocket Pool service version: %w", err)
		}
		versionString = response.Data.Version
	} else {
		cmd := fmt.Sprintf("%s --version", shellescape.Quote(c.daemonPath))
		versionBytes, err := c.readOutput(cmd)
		if err != nil {
			return "", fmt.Errorf("error getting Rocket Pool service version: %w", err)
		}
		// Get the version string
		outputString := string(versionBytes)
		elements := strings.Fields(outputString) // Split on whitespace
		if len(elements) < 1 {
			return "", fmt.Errorf("error parsing Rocket Pool service version number from output '%s'", outputString)
		}
		versionString = elements[len(elements)-1]
	}

	// Make sure it's a semantic version
	version, err := semver.Make(versionString)
	if err != nil {
		return "", fmt.Errorf("error parsing Rocket Pool service version number from output '%s': %w", versionString, err)
	}

	// Return the parsed semantic version (extra safety)
	return version.String(), nil
}

// Deletes the node wallet and all validator keys, and restarts the Docker containers
func (c *Client) PurgeAllKeys(composeFiles []string) error {
	// Get the command to run with root privileges
	rootCmd, err := c.getEscalationCommand()
	if err != nil {
		return fmt.Errorf("could not get privilege escalation command: %w", err)
	}

	// Get the config
	cfg, _, err := c.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading user settings: %w", err)
	}

	// Check for Native mode
	if cfg.IsNativeMode {
		return fmt.Errorf("this function is not supported in Native Mode; you will have to shut down your client and daemon services and remove the keys manually")
	}

	// Shut down the containers
	fmt.Println("Stopping containers...")
	err = c.PauseService(composeFiles)
	if err != nil {
		return fmt.Errorf("error stopping Docker containers: %w", err)
	}

	// Delete the wallet
	walletPath, err := homedir.Expand(cfg.Smartnode.GetWalletPathInCLI())
	if err != nil {
		return fmt.Errorf("error loading wallet path: %w", err)
	}
	fmt.Println("Deleting wallet...")
	cmd := fmt.Sprintf("%s rm -f %s", rootCmd, walletPath)
	_, err = c.readOutput(cmd)
	if err != nil {
		return fmt.Errorf("error deleting wallet: %w", err)
	}

	// Delete the password
	passwordPath, err := homedir.Expand(cfg.Smartnode.GetPasswordPathInCLI())
	if err != nil {
		return fmt.Errorf("error loading password path: %w", err)
	}
	fmt.Println("Deleting password...")
	cmd = fmt.Sprintf("%s rm -f %s", rootCmd, passwordPath)
	_, err = c.readOutput(cmd)
	if err != nil {
		return fmt.Errorf("error deleting password: %w", err)
	}

	// Delete the validators dir
	validatorsPath, err := homedir.Expand(cfg.Smartnode.GetValidatorKeychainPathInCLI())
	if err != nil {
		return fmt.Errorf("error loading validators folder path: %w", err)
	}
	fmt.Println("Deleting validator keys...")
	cmd = fmt.Sprintf("%s rm -rf %s/*", rootCmd, validatorsPath)
	_, err = c.readOutput(cmd)
	if err != nil {
		return fmt.Errorf("error deleting validator keys: %w", err)
	}
	cmd = fmt.Sprintf("%s rm -rf %s/.[a-zA-Z0-9]*", rootCmd, validatorsPath)
	_, err = c.readOutput(cmd)
	if err != nil {
		return fmt.Errorf("error deleting hidden files in validator folder: %w", err)
	}

	// Start the containers
	fmt.Println("Starting containers...")
	err = c.StartService(composeFiles)
	if err != nil {
		return fmt.Errorf("error starting Docker containers: %w", err)
	}

	fmt.Println("Purge complete.")

	return nil
}

// Runs the prune provisioner
func (c *Client) RunPruneProvisioner(container string, volume string, image string) error {

	// Run the prune provisioner
	cmd := fmt.Sprintf("docker run --rm --name %s -v %s:/ethclient %s", container, volume, image)
	output, err := c.readOutput(cmd)
	if err != nil {
		return err
	}

	outputString := strings.TrimSpace(string(output))
	if outputString != "" {
		return fmt.Errorf("Unexpected output running the prune provisioner: %s", outputString)
	}

	return nil

}

// Runs the prune provisioner
func (c *Client) RunNethermindPruneStarter(container string) error {
	cmd := fmt.Sprintf("docker exec %s %s %s", container, nethermindPruneStarterCommand, nethermindAdminUrl)
	err := c.printOutput(cmd)
	if err != nil {
		return err
	}
	return nil
}

// Runs the EC migrator
func (c *Client) RunEcMigrator(container string, volume string, targetDir string, mode string, image string) error {
	cmd := fmt.Sprintf("docker run --rm --name %s -v %s:/ethclient -v %s:/mnt/external -e EC_MIGRATE_MODE='%s' %s", container, volume, targetDir, mode, image)
	err := c.printOutput(cmd)
	if err != nil {
		return err
	}

	return nil
}

// Gets the size of the target directory via the EC migrator for importing, which should have the same permissions as exporting
func (c *Client) GetDirSizeViaEcMigrator(container string, targetDir string, image string) (uint64, error) {
	cmd := fmt.Sprintf("docker run --rm --name %s -v %s:/mnt/external -e OPERATION='size' %s", container, targetDir, image)
	output, err := c.readOutput(cmd)
	if err != nil {
		return 0, fmt.Errorf("Error getting source directory size: %w", err)
	}

	trimmedOutput := strings.TrimRight(string(output), "\n")
	dirSize, err := strconv.ParseUint(trimmedOutput, 0, 64)
	if err != nil {
		return 0, fmt.Errorf("Error parsing directory size output [%s]: %w", trimmedOutput, err)
	}

	return dirSize, nil
}

// Build a docker compose command
func (c *Client) compose(composeFiles []string, args string) (string, error) {

	// Cancel if running in non-docker mode
	if c.daemonPath != "" {
		return "", errors.New("command unavailable in Native Mode (with '--daemon-path' option specified)")
	}

	// Get the expanded config path
	expandedConfigPath, err := homedir.Expand(c.configPath)
	if err != nil {
		return "", err
	}

	// Load config
	cfg, isNew, err := c.LoadConfig()
	if err != nil {
		return "", err
	}

	if isNew {
		return "", fmt.Errorf("Settings file not found. Please run `rocketpool service config` to set up your Smartnode before starting it.")
	}

	// Check config
	if cfg.ExecutionClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Unknown {
		return "", fmt.Errorf("You haven't selected local or external mode for your Execution (ETH1) client.\nPlease run 'rocketpool service config' before running this command.")
	} else if cfg.ExecutionClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local && cfg.ExecutionClient.Value.(cfgtypes.ExecutionClient) == cfgtypes.ExecutionClient_Unknown {
		return "", errors.New("No Execution (ETH1) client selected. Please run 'rocketpool service config' before running this command.")
	}
	if cfg.ConsensusClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Unknown {
		return "", fmt.Errorf("You haven't selected local or external mode for your Consensus (ETH2) client.\nPlease run 'rocketpool service config' before running this command.")
	} else if cfg.ConsensusClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local && cfg.ConsensusClient.Value.(cfgtypes.ConsensusClient) == cfgtypes.ConsensusClient_Unknown {
		return "", errors.New("No Consensus (ETH2) client selected. Please run 'rocketpool service config' before running this command.")
	}

	// Get the external IP address
	var externalIP string
	ip, err := getExternalIP()
	if err != nil {
		fmt.Println("Warning: couldn't get external IP address; if you're using Nimbus or Besu, it may have trouble finding peers:")
		fmt.Println(err.Error())
	} else {
		if ip.To4() == nil {
			fmt.Println("Warning: external IP address is v6; if you're using Nimbus or Besu, it may have trouble finding peers:")
		}
		externalIP = ip.String()
	}

	// Set up environment variables and deploy the template config files
	settings := cfg.GenerateEnvironmentVariables()
	if externalIP != "" {
		settings["EXTERNAL_IP"] = shellescape.Quote(externalIP)
	}

	// Deploy the templates and run environment variable substitution on them
	deployedContainers, err := c.deployTemplates(cfg, expandedConfigPath, settings)
	if err != nil {
		return "", fmt.Errorf("error deploying Docker templates: %w", err)
	}

	// Set up all of the environment variables to pass to the run command
	env := []string{}
	for key, value := range settings {
		env = append(env, fmt.Sprintf("%s=%s", key, shellescape.Quote(value)))
	}

	// Include all of the relevant docker compose definition files
	composeFileFlags := []string{}
	for _, container := range deployedContainers {
		composeFileFlags = append(composeFileFlags, fmt.Sprintf("-f %s", shellescape.Quote(container)))
	}
	for _, container := range composeFiles {
		composeFileFlags = append(composeFileFlags, fmt.Sprintf("-f %s", shellescape.Quote(container)))
	}

	// Return command
	return fmt.Sprintf("%s docker compose --project-directory %s %s %s", strings.Join(env, " "), shellescape.Quote(expandedConfigPath), strings.Join(composeFileFlags, " "), args), nil

}

// Deploys all of the appropriate docker compose template files and provisions them based on the provided configuration
func (c *Client) deployTemplates(cfg *config.RocketPoolConfig, rocketpoolDir string, settings map[string]string) ([]string, error) {

	// Check for the folders
	runtimeFolder := filepath.Join(rocketpoolDir, runtimeDir)
	templatesFolder := filepath.Join(rocketpoolDir, templatesDir)
	_, err := os.Stat(templatesFolder)
	if os.IsNotExist(err) {
		return []string{}, fmt.Errorf("templates folder [%s] does not exist", templatesFolder)
	}
	overrideFolder := filepath.Join(rocketpoolDir, overrideDir)
	_, err = os.Stat(overrideFolder)
	if os.IsNotExist(err) {
		return []string{}, fmt.Errorf("override folder [%s] does not exist", overrideFolder)
	}

	// Clear out the runtime folder and remake it
	err = os.RemoveAll(runtimeFolder)
	if err != nil {
		return []string{}, fmt.Errorf("error deleting runtime folder [%s]: %w", runtimeFolder, err)
	}
	err = os.Mkdir(runtimeFolder, 0775)
	if err != nil {
		return []string{}, fmt.Errorf("error creating runtime folder [%s]: %w", runtimeFolder, err)
	}

	// Set the environment variables for substitution
	oldValues := map[string]string{}
	for varName, varValue := range settings {
		oldValues[varName] = os.Getenv(varName)
		os.Setenv(varName, varValue)
	}
	defer func() {
		// Unset the env vars
		for name, value := range oldValues {
			os.Setenv(name, value)
		}
	}()

	// Read and substitute the templates
	deployedContainers := []string{}

	// Node
	contents, err := envsubst.ReadFile(filepath.Join(templatesFolder, string(docker.ContainerName_Node)+templateSuffix))
	if err != nil {
		return []string{}, fmt.Errorf("error reading and substituting node container template: %w", err)
	}
	nodeComposePath := filepath.Join(runtimeFolder, string(docker.ContainerName_Node)+composeFileSuffix)
	err = os.WriteFile(nodeComposePath, contents, 0664)
	if err != nil {
		return []string{}, fmt.Errorf("could not write node container file to %s: %w", nodeComposePath, err)
	}
	deployedContainers = append(deployedContainers, nodeComposePath)
	deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, string(docker.ContainerName_Node)+composeFileSuffix))

	// Watchtower
	contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, string(docker.ContainerName_Watchtower)+templateSuffix))
	if err != nil {
		return []string{}, fmt.Errorf("error reading and substituting watchtower container template: %w", err)
	}
	watchtowerComposePath := filepath.Join(runtimeFolder, string(docker.ContainerName_Watchtower)+composeFileSuffix)
	err = os.WriteFile(watchtowerComposePath, contents, 0664)
	if err != nil {
		return []string{}, fmt.Errorf("could not write watchtower container file to %s: %w", watchtowerComposePath, err)
	}
	deployedContainers = append(deployedContainers, watchtowerComposePath)
	deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, string(docker.ContainerName_Watchtower)+composeFileSuffix))

	// Validator
	contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, string(docker.ContainerName_ValidatorClient)+templateSuffix))
	if err != nil {
		return []string{}, fmt.Errorf("error reading and substituting validator container template: %w", err)
	}
	validatorComposePath := filepath.Join(runtimeFolder, string(docker.ContainerName_ValidatorClient)+composeFileSuffix)
	err = os.WriteFile(validatorComposePath, contents, 0664)
	if err != nil {
		return []string{}, fmt.Errorf("could not write validator container file to %s: %w", validatorComposePath, err)
	}
	deployedContainers = append(deployedContainers, validatorComposePath)
	deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, string(docker.ContainerName_ValidatorClient)+composeFileSuffix))

	// Check the EC mode to see if it needs to be deployed
	if cfg.ExecutionClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
		contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, string(docker.ContainerName_ExecutionClient)+templateSuffix))
		if err != nil {
			return []string{}, fmt.Errorf("error reading and substituting execution client container template: %w", err)
		}
		eth1ComposePath := filepath.Join(runtimeFolder, string(docker.ContainerName_ExecutionClient)+composeFileSuffix)
		err = os.WriteFile(eth1ComposePath, contents, 0664)
		if err != nil {
			return []string{}, fmt.Errorf("could not write execution client container file to %s: %w", eth1ComposePath, err)
		}
		deployedContainers = append(deployedContainers, eth1ComposePath)
		deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, string(docker.ContainerName_ExecutionClient)+composeFileSuffix))
	}

	// Check the Consensus mode
	if cfg.ConsensusClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
		contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, string(docker.ContainerName_BeaconNode)+templateSuffix))
		if err != nil {
			return []string{}, fmt.Errorf("error reading and substituting consensus client container template: %w", err)
		}
		eth2ComposePath := filepath.Join(runtimeFolder, string(docker.ContainerName_BeaconNode)+composeFileSuffix)
		err = os.WriteFile(eth2ComposePath, contents, 0664)
		if err != nil {
			return []string{}, fmt.Errorf("could not write consensus client container file to %s: %w", eth2ComposePath, err)
		}
		deployedContainers = append(deployedContainers, eth2ComposePath)
		deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, string(docker.ContainerName_BeaconNode)+composeFileSuffix))
	}

	// Check the metrics containers
	if cfg.EnableMetrics.Value == true {
		// Grafana
		contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, string(docker.ContainerName_Grafana)+templateSuffix))
		if err != nil {
			return []string{}, fmt.Errorf("error reading and substituting Grafana container template: %w", err)
		}
		grafanaComposePath := filepath.Join(runtimeFolder, string(docker.ContainerName_Grafana)+composeFileSuffix)
		err = os.WriteFile(grafanaComposePath, contents, 0664)
		if err != nil {
			return []string{}, fmt.Errorf("could not write Grafana container file to %s: %w", grafanaComposePath, err)
		}
		deployedContainers = append(deployedContainers, grafanaComposePath)
		deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, string(docker.ContainerName_Grafana)+composeFileSuffix))

		// Node exporter
		contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, string(docker.ContainerName_Exporter)+templateSuffix))
		if err != nil {
			return []string{}, fmt.Errorf("error reading and substituting Node Exporter container template: %w", err)
		}
		exporterComposePath := filepath.Join(runtimeFolder, string(docker.ContainerName_Exporter)+composeFileSuffix)
		err = os.WriteFile(exporterComposePath, contents, 0664)
		if err != nil {
			return []string{}, fmt.Errorf("could not write Node Exporter container file to %s: %w", exporterComposePath, err)
		}
		deployedContainers = append(deployedContainers, exporterComposePath)
		deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, string(docker.ContainerName_Exporter)+composeFileSuffix))

		// Prometheus
		contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, string(docker.ContainerName_Prometheus)+templateSuffix))
		if err != nil {
			return []string{}, fmt.Errorf("error reading and substituting Prometheus container template: %w", err)
		}
		prometheusComposePath := filepath.Join(runtimeFolder, string(docker.ContainerName_Prometheus)+composeFileSuffix)
		err = os.WriteFile(prometheusComposePath, contents, 0664)
		if err != nil {
			return []string{}, fmt.Errorf("could not write Prometheus container file to %s: %w", prometheusComposePath, err)
		}
		deployedContainers = append(deployedContainers, prometheusComposePath)
		deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, string(docker.ContainerName_Prometheus)+composeFileSuffix))
	}

	// Check MEV-Boost
	if cfg.EnableMevBoost.Value == true && cfg.MevBoost.Mode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
		contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, string(docker.ContainerName_MevBoost)+templateSuffix))
		if err != nil {
			return []string{}, fmt.Errorf("error reading and substituting MEV-Boost container template: %w", err)
		}
		mevBoostComposePath := filepath.Join(runtimeFolder, string(docker.ContainerName_MevBoost)+composeFileSuffix)
		err = os.WriteFile(mevBoostComposePath, contents, 0664)
		if err != nil {
			return []string{}, fmt.Errorf("could not write MEV-Boost container file to %s: %w", mevBoostComposePath, err)
		}
		deployedContainers = append(deployedContainers, mevBoostComposePath)
		deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, string(docker.ContainerName_MevBoost)+composeFileSuffix))
	}

	// Create the custom keys dir
	customKeyDir, err := homedir.Expand(filepath.Join(cfg.Smartnode.DataPath.Value.(string), "custom-keys"))
	if err != nil {
		fmt.Printf("%sWARNING: Couldn't expand the custom validator key directory (%s). You will not be able to recover any minipool keys you created outside of the Smartnode until you create the folder manually.%s\n", terminal.ColorYellow, err.Error(), terminal.ColorReset)
		return deployedContainers, nil
	}
	err = os.MkdirAll(customKeyDir, 0775)
	if err != nil {
		fmt.Printf("%sWARNING: Couldn't create the custom validator key directory (%s). You will not be able to recover any minipool keys you created outside of the Smartnode until you create the folder [%s] manually.%s\n", terminal.ColorYellow, err.Error(), customKeyDir, terminal.ColorReset)
	}

	// Create the rewards file dir
	rewardsFilePath, err := homedir.Expand(cfg.Smartnode.GetRewardsTreePath(0, false))
	if err != nil {
		fmt.Printf("%sWARNING: Couldn't expand the rewards tree file directory (%s). You will not be able to view or claim your rewards until you create the folder manually.%s\n", terminal.ColorYellow, err.Error(), terminal.ColorReset)
		return deployedContainers, nil
	}
	rewardsFileDir := filepath.Dir(rewardsFilePath)
	err = os.MkdirAll(rewardsFileDir, 0775)
	if err != nil {
		fmt.Printf("%sWARNING: Couldn't create the rewards tree file directory (%s). You will not be able to view or claim your rewards until you create the folder [%s] manually.%s\n", terminal.ColorYellow, err.Error(), rewardsFileDir, terminal.ColorReset)
	}

	return c.composeAddons(cfg, rocketpoolDir, settings, deployedContainers)

}

// Handle composing for addons
func (c *Client) composeAddons(cfg *config.RocketPoolConfig, rocketpoolDir string, settings map[string]string, deployedContainers []string) ([]string, error) {

	// GWW
	if cfg.GraffitiWallWriter.GetEnabledParameter().Value == true {
		runtimeFolder := filepath.Join(rocketpoolDir, runtimeDir, "addons", "gww")
		templatesFolder := filepath.Join(rocketpoolDir, templatesDir, "addons", "gww")
		overrideFolder := filepath.Join(rocketpoolDir, overrideDir, "addons", "gww")

		// Make the addon folder
		err := os.MkdirAll(runtimeFolder, 0775)
		if err != nil {
			return []string{}, fmt.Errorf("error creating addon runtime folder (%s): %w", runtimeFolder, err)
		}

		contents, err := envsubst.ReadFile(filepath.Join(templatesFolder, graffiti_wall_writer.GraffitiWallWriterContainerName+templateSuffix))
		if err != nil {
			return []string{}, fmt.Errorf("error reading and substituting GWW addon container template: %w", err)
		}
		composePath := filepath.Join(runtimeFolder, graffiti_wall_writer.GraffitiWallWriterContainerName+composeFileSuffix)
		err = os.WriteFile(composePath, contents, 0664)
		if err != nil {
			return []string{}, fmt.Errorf("could not write GWW addon container file to %s: %w", composePath, err)
		}
		deployedContainers = append(deployedContainers, composePath)
		deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, graffiti_wall_writer.GraffitiWallWriterContainerName+composeFileSuffix))
	}

	return deployedContainers, nil

}
