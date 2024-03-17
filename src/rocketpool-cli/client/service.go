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

	"github.com/alessio/shellescape"
	"github.com/blang/semver/v4"
	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	nmc_config "github.com/rocket-pool/node-manager-core/config"
	gww "github.com/rocket-pool/smartnode/addons/graffiti_wall_writer"
	"github.com/rocket-pool/smartnode/rocketpool-cli/client/template"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/shared/config"
)

const (
	debugColor color.Attribute = color.FgYellow

	nethermindPruneStarterCommand string = "dotnet /setup/NethermindPruneStarter/NethermindPruneStarter.dll"
	nethermindAdminUrl            string = "http://127.0.0.1:7434"

	templatesDir       string = "/usr/share/rocketpool/templates"
	overrideSourceDir  string = "/usr/share/rocketpool/override"
	overrideDir        string = "override"
	runtimeDir         string = "runtime"
	extraScrapeJobsDir string = "extra-scrape-jobs"
)

// Install the Rocket Pool service
func (c *Client) InstallService(verbose bool, noDeps bool, version string, path string, useLocalInstaller bool) error {
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

	var script []byte
	if useLocalInstaller {
		// Make sure it exists
		_, err := os.Stat(InstallerName)
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("local install script [%s] does not exist", InstallerName)
		}
		if err != nil {
			return fmt.Errorf("error checking install script [%s]: %w", InstallerName, err)
		}

		// Read it
		script, err = os.ReadFile(InstallerName)
		if err != nil {
			return fmt.Errorf("error reading local install script [%s]: %w", InstallerName, err)
		}

		// Set the "local mode" flag
		flags = append(flags, "-l")
	} else {
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
		script, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if fmt.Sprint(len(script)) != resp.Header.Get("content-length") {
			return fmt.Errorf("downloaded script length %d did not match content-length header %s", len(script), resp.Header.Get("content-length"))
		}
	}

	// Get the escalation command
	escalationCmd, err := c.getEscalationCommand()
	if err != nil {
		return fmt.Errorf("error getting escalation command: %w", err)
	}

	// Initialize installation command
	cmd := c.newCommand(fmt.Sprintf("%s sh -s -- %s", escalationCmd, strings.Join(flags, " ")))

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
		c := color.New(debugColor)
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
		return fmt.Errorf("could not install Smart Node service: %s", errMessage)
	}
	return nil
}

// Install the update tracker
func (c *Client) InstallUpdateTracker(verbose bool, version string, useLocalInstaller bool) error {
	// Get installation script flags
	flags := []string{
		"-v", fmt.Sprintf("%s", shellescape.Quote(version)),
	}

	var script []byte
	if useLocalInstaller {
		// Make sure it exists
		_, err := os.Stat(UpdateTrackerInstallerName)
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("local update tracker install script [%s] does not exist", UpdateTrackerInstallerName)
		}
		if err != nil {
			return fmt.Errorf("error checking update tracker install script [%s]: %w", UpdateTrackerInstallerName, err)
		}

		// Read it
		script, err = os.ReadFile(UpdateTrackerInstallerName)
		if err != nil {
			return fmt.Errorf("error reading local update tracker install script [%s]: %w", UpdateTrackerInstallerName, err)
		}

		// Set the "local mode" flag
		flags = append(flags, "-l")
	} else {
		// Download the update tracker script
		resp, err := http.Get(fmt.Sprintf(UpdateTrackerURL, version))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected http status downloading update tracker installer script: %d", resp.StatusCode)
		}

		// Sanity check that the script octet length matches content-length
		script, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if fmt.Sprint(len(script)) != resp.Header.Get("content-length") {
			return fmt.Errorf("downloaded script length %d did not match content-length header %s", len(script), resp.Header.Get("content-length"))
		}
	}

	// Get the escalation command
	escalationCmd, err := c.getEscalationCommand()
	if err != nil {
		return fmt.Errorf("error getting escalation command: %w", err)
	}

	// Initialize installation command
	cmd := c.newCommand(fmt.Sprintf("%s sh -s -- %s", escalationCmd, strings.Join(flags, " ")))

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
		c := color.New(debugColor)
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
		return fmt.Errorf("could not install Rocket Pool update tracker: %s", errMessage)
	}
	return nil
}

// Start the Rocket Pool service
func (c *Client) StartService(composeFiles []string) error {
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
	response, err := c.Api.Service.Version()
	if err != nil {
		return "", fmt.Errorf("error requesting Rocket Pool service version: %w", err)
	}
	versionString := response.Data.Version

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

	// Delete the address
	nodeAddressPath, err := homedir.Expand(cfg.GetNodeAddressPath())
	if err != nil {
		return fmt.Errorf("error loading node address file path: %w", err)
	}
	fmt.Println("Deleting node address file...")
	cmd := fmt.Sprintf("%s rm -f %s", rootCmd, nodeAddressPath)
	_, err = c.readOutput(cmd)
	if err != nil {
		return fmt.Errorf("error deleting node address file: %w", err)
	}

	// Delete the wallet
	walletPath, err := homedir.Expand(cfg.GetWalletPath())
	if err != nil {
		return fmt.Errorf("error loading wallet path: %w", err)
	}
	fmt.Println("Deleting wallet...")
	cmd = fmt.Sprintf("%s rm -f %s", rootCmd, walletPath)
	_, err = c.readOutput(cmd)
	if err != nil {
		return fmt.Errorf("error deleting wallet: %w", err)
	}

	// Delete the next account file
	nextAccountPath, err := homedir.Expand(cfg.GetNextAccountFilePath())
	if err != nil {
		return fmt.Errorf("error loading next account file path: %w", err)
	}
	fmt.Println("Deleting next account file...")
	cmd = fmt.Sprintf("%s rm -f %s", rootCmd, nextAccountPath)
	_, err = c.readOutput(cmd)
	if err != nil {
		return fmt.Errorf("error deleting next account file: %w", err)
	}

	// Delete the password
	passwordPath, err := homedir.Expand(cfg.GetPasswordPath())
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
	validatorsPath, err := homedir.Expand(cfg.GetValidatorsFolderPath())
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
func (c *Client) RunPruneProvisioner(container string, volume string) error {
	// Run the prune provisioner
	cmd := fmt.Sprintf("docker run --rm --name %s -v %s:/ethclient %s", container, volume, config.PruneProvisionerTag)
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
func (c *Client) RunEcMigrator(container string, volume string, targetDir string, mode string) error {
	cmd := fmt.Sprintf("docker run --rm --name %s -v %s:/ethclient -v %s:/mnt/external -e EC_MIGRATE_MODE='%s' %s", container, volume, targetDir, mode, config.EcMigratorTag)
	err := c.printOutput(cmd)
	if err != nil {
		return err
	}

	return nil
}

// Gets the size of the target directory via the EC migrator for importing, which should have the same permissions as exporting
func (c *Client) GetDirSizeViaEcMigrator(container string, targetDir string) (uint64, error) {
	cmd := fmt.Sprintf("docker run --rm --name %s -v %s:/mnt/external -e OPERATION='size' %s", container, targetDir, config.EcMigratorTag)
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
	// Cancel if running in native mode
	if c.Context.NativeMode {
		return "", errors.New("command unavailable in Native Mode")
	}

	// Get the expanded config path
	expandedConfigPath, err := homedir.Expand(c.Context.ConfigPath)
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
	if cfg.ClientMode.Value == nmc_config.ClientMode_Unknown {
		return "", fmt.Errorf("You haven't selected local or external mode for your clients yet.\nPlease run 'rocketpool service config' before running this command.")
	} else if cfg.IsLocalMode() && cfg.LocalExecutionClient.ExecutionClient.Value == nmc_config.ExecutionClient_Unknown {
		return "", errors.New("no Execution Client selected. Please run 'rocketpool service config' before running this command")
	}
	if cfg.IsLocalMode() && cfg.LocalBeaconClient.BeaconNode.Value == nmc_config.BeaconNode_Unknown {
		return "", errors.New("no Beacon Node selected. Please run 'rocketpool service config' before running this command")
	}

	// Deploy the templates and run environment variable substitution on them
	deployedContainers, err := c.deployTemplates(cfg, expandedConfigPath)
	if err != nil {
		return "", fmt.Errorf("error deploying Docker templates: %w", err)
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
	return fmt.Sprintf("COMPOSE_PROJECT_NAME=%s docker compose --project-directory %s %s %s", cfg.ProjectName.Value, shellescape.Quote(expandedConfigPath), strings.Join(composeFileFlags, " "), args), nil
}

// Deploys all of the appropriate docker compose template files and provisions them based on the provided configuration
func (c *Client) deployTemplates(cfg *config.SmartNodeConfig, smartNodeDir string) ([]string, error) {
	// Prep the override folder
	overrideFolder := filepath.Join(smartNodeDir, overrideDir)
	copyOverrideFiles(overrideSourceDir, overrideFolder)

	// Clear out the runtime folder and remake it
	runtimeFolder := filepath.Join(smartNodeDir, runtimeDir)
	err := os.RemoveAll(runtimeFolder)
	if err != nil {
		return []string{}, fmt.Errorf("error deleting runtime folder [%s]: %w", runtimeFolder, err)
	}
	err = os.Mkdir(runtimeFolder, 0775)
	if err != nil {
		return []string{}, fmt.Errorf("error creating runtime folder [%s]: %w", runtimeFolder, err)
	}

	// Make the extra scrape jobs folder
	extraScrapeJobsFolder := filepath.Join(smartNodeDir, extraScrapeJobsDir)
	err = os.MkdirAll(extraScrapeJobsFolder, 0755)
	if err != nil {
		return []string{}, fmt.Errorf("error creating extra-scrape-jobs folder: %w", err)
	}

	composePaths := template.ComposePaths{
		RuntimePath:  runtimeFolder,
		TemplatePath: templatesDir,
		OverridePath: overrideFolder,
	}

	// Read and substitute the templates
	deployedContainers := []string{}

	// These containers always run
	toDeploy := []nmc_config.ContainerID{
		nmc_config.ContainerID_Daemon,
		config.ContainerID_Watchtower,
		nmc_config.ContainerID_ValidatorClient,
	}

	// Check if we are running the Execution Layer locally
	if cfg.IsLocalMode() {
		toDeploy = append(toDeploy, nmc_config.ContainerID_ExecutionClient)
		toDeploy = append(toDeploy, nmc_config.ContainerID_BeaconNode)
	}

	// Check the metrics containers
	if cfg.Metrics.EnableMetrics.Value {
		toDeploy = append(toDeploy,
			nmc_config.ContainerID_Grafana,
			nmc_config.ContainerID_Exporter,
			nmc_config.ContainerID_Prometheus,
		)
	}

	// Check if we are running the MEV-Boost container locally
	if cfg.MevBoost.Enable.Value && cfg.MevBoost.Mode.Value == nmc_config.ClientMode_Local {
		toDeploy = append(toDeploy, nmc_config.ContainerID_MevBoost)
	}

	// Deploy main containers
	for _, containerID := range toDeploy {
		containerName := config.GetContainerName(containerID)
		containers, err := composePaths.File(containerName).Write(cfg)
		if err != nil {
			return []string{}, fmt.Errorf("could not create %s container definition: %w", containerName, err)
		}
		deployedContainers = append(deployedContainers, containers...)
	}

	// Create the custom keys dir
	customKeyDir, err := homedir.Expand(filepath.Join(cfg.UserDataPath.Value, config.CustomKeysFolderName))
	if err != nil {
		fmt.Printf("%sWARNING: Couldn't expand the custom validator key directory (%s). You will not be able to recover any minipool keys you created outside of the Smart Node until you create the folder manually.%s\n", terminal.ColorYellow, err.Error(), terminal.ColorReset)
		return deployedContainers, nil
	}
	err = os.MkdirAll(customKeyDir, 0775)
	if err != nil {
		fmt.Printf("%sWARNING: Couldn't create the custom validator key directory (%s). You will not be able to recover any minipool keys you created outside of the Smart Node until you create the folder [%s] manually.%s\n", terminal.ColorYellow, err.Error(), customKeyDir, terminal.ColorReset)
	}

	// Create the rewards file dir
	rewardsFilePath, err := homedir.Expand(cfg.GetRewardsTreePath(0))
	if err != nil {
		fmt.Printf("%sWARNING: Couldn't expand the rewards tree file directory (%s). You will not be able to view or claim your rewards until you create the folder manually.%s\n", terminal.ColorYellow, err.Error(), terminal.ColorReset)
		return deployedContainers, nil
	}
	rewardsFileDir := filepath.Dir(rewardsFilePath)
	err = os.MkdirAll(rewardsFileDir, 0775)
	if err != nil {
		fmt.Printf("%sWARNING: Couldn't create the rewards tree file directory (%s). You will not be able to view or claim your rewards until you create the folder [%s] manually.%s\n", terminal.ColorYellow, err.Error(), rewardsFileDir, terminal.ColorReset)
	}

	return c.composeAddons(cfg, smartNodeDir, deployedContainers)
}

// Make sure the override files have all been copied to the local user dir
func copyOverrideFiles(sourceDir string, targetDir string) error {
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating override folder: %w", err)
	}

	files, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("error enumerating override source folder: %w", err)
	}

	// Copy any override files that don't exist in the local user directory
	for _, file := range files {
		filename := file.Name()
		targetPath := filepath.Join(targetDir, filename)
		if file.IsDir() {
			// Recurse
			srcPath := filepath.Join(sourceDir, file.Name())
			copyOverrideFiles(srcPath, targetPath)
		}

		_, err := os.Stat(targetPath)
		if !os.IsNotExist(err) {
			// Ignore files that already exist
			continue
		}

		// Read the source
		srcPath := filepath.Join(sourceDir, filename)
		contents, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("error reading override file [%s]: %w", srcPath, err)
		}

		// Write a copy to the user dir
		err = os.WriteFile(targetPath, contents, 0644)
		if err != nil {
			return fmt.Errorf("error writing local override file [%s]: %w", targetPath, err)
		}
	}
	return nil
}

// Handle composing for addons
func (c *Client) composeAddons(cfg *config.SmartNodeConfig, rocketpoolDir string, deployedContainers []string) ([]string, error) {
	// GWW
	if cfg.Addons.GraffitiWallWriter.Enabled.Value {
		composePaths := template.ComposePaths{
			RuntimePath:  filepath.Join(rocketpoolDir, runtimeDir, config.AddonsFolderName, gww.FolderName),
			TemplatePath: filepath.Join(rocketpoolDir, templatesDir, config.AddonsFolderName, gww.FolderName),
			OverridePath: filepath.Join(rocketpoolDir, overrideDir, config.AddonsFolderName, gww.FolderName),
		}

		// Make the addon folder
		err := os.MkdirAll(composePaths.RuntimePath, 0775)
		if err != nil {
			return []string{}, fmt.Errorf("error creating addon runtime folder (%s): %w", composePaths.RuntimePath, err)
		}

		containers, err := composePaths.File(string(gww.ContainerID_GraffitiWallWriter)).Write(cfg)
		if err != nil {
			return []string{}, fmt.Errorf("could not create gww container definition: %w", err)
		}
		deployedContainers = append(deployedContainers, containers...)
	}

	return deployedContainers, nil
}
