package client

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/mitchellh/go-homedir"
	nmc_config "github.com/rocket-pool/node-manager-core/config"
	gww "github.com/rocket-pool/smartnode/v2/addons/graffiti_wall_writer"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client/template"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/shared/config"
)

const (
	templatesDir           string = "/usr/share/rocketpool/templates"
	addonsSourceDir        string = "/usr/share/rocketpool/addons"
	overrideSourceDir      string = "/usr/share/rocketpool/override"
	nativeScriptsSourceDir string = "/usr/share/rocketpool/scripts/native"
	overrideDir            string = "override"
	runtimeDir             string = "runtime"
	extraScrapeJobsDir     string = "extra-scrape-jobs"
)

// Returns the images used by each service in compose file in "repository:tag"
// format (assuming that is the format specified in the compose files)
func (c *Client) GetComposeImages(composeFiles []string) ([]string, error) {
	cmd, err := c.compose(composeFiles, "config --images")
	if err != nil {
		return nil, err
	}
	output, err := c.readOutput(cmd)
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(output)), nil
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
	err := copyStockFiles(overrideSourceDir, overrideFolder, "override")
	if err != nil {
		return nil, fmt.Errorf("error copying override files: %w", err)
	}

	// Prep the addons folder
	addonsFolder := filepath.Join(smartNodeDir, config.AddonsFolderName)
	err = copyStockFiles(addonsSourceDir, addonsFolder, "addons")
	if err != nil {
		return nil, fmt.Errorf("error copying addons files: %w", err)
	}

	// Prep the native scripts folder
	nativeScriptsFolder := filepath.Join(smartNodeDir, config.NativeScriptsFolderName)
	err = copyStockFiles(nativeScriptsSourceDir, nativeScriptsFolder, "native scripts")
	if err != nil {
		return nil, fmt.Errorf("error copying native scripts files: %w", err)
	}

	// Remove the obsolete Docker Compose version from the overrides
	err = removeComposeVersion(overrideFolder)
	if err != nil {
		return nil, fmt.Errorf("error removing obsolete Docker Compose version from overrides: %w", err)
	}

	// Clear out the runtime folder and remake it
	runtimeFolder := filepath.Join(smartNodeDir, runtimeDir)
	err = os.RemoveAll(runtimeFolder)
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

// Copy stock installation files from a directory to the user's directory for customization
func copyStockFiles(sourceDir string, targetDir string, filetype string) error {
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating %s folder [%s]: %w", filetype, targetDir, err)
	}

	files, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("error enumerating %s source folder [%s]: %w", filetype, sourceDir, err)
	}

	// Copy any override files that don't exist in the local user directory
	for _, file := range files {
		filename := file.Name()
		targetPath := filepath.Join(targetDir, filename)
		if file.IsDir() {
			// Recurse
			srcPath := filepath.Join(sourceDir, file.Name())
			err = copyStockFiles(srcPath, targetPath, filetype)
			if err != nil {
				return err
			}
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
			return fmt.Errorf("error reading %s file [%s]: %w", filetype, srcPath, err)
		}

		// Write a copy to the user dir
		err = os.WriteFile(targetPath, contents, 0644)
		if err != nil {
			return fmt.Errorf("error writing local %s file [%s]: %w", filetype, targetPath, err)
		}
	}
	return nil
}

// Remove the obsolete Docker Compose version from each compose file in the target directory
func removeComposeVersion(targetDir string) error {
	files, err := os.ReadDir(targetDir)
	if err != nil {
		return fmt.Errorf("error enumerating folder [%s]: %w", targetDir, err)
	}

	// Copy any override files that don't exist in the local user directory
	for _, file := range files {
		filename := file.Name()
		targetPath := filepath.Join(targetDir, filename)
		if file.IsDir() {
			// Recurse
			subdir := filepath.Join(targetDir, file.Name())
			err = removeComposeVersion(subdir)
			if err != nil {
				return err
			}
		}

		// Ignore it if it's not a YAML file
		if filepath.Ext(filename) != ".yml" {
			continue
		}

		// Read the source
		contents, err := os.ReadFile(targetPath)
		if err != nil {
			return fmt.Errorf("error reading file [%s]: %w", targetPath, err)
		}

		// Remove the version field, accounting for both Windows and Unix line endings
		newContents := bytes.ReplaceAll(contents, []byte("\r\nversion: \"3.7\""), []byte("\r\n"))
		newContents = bytes.ReplaceAll(newContents, []byte("\nversion: \"3.7\""), []byte("\n"))

		// Write the updated contents if they differ
		if len(newContents) != len(contents) {
			err = os.WriteFile(targetPath, newContents, 0644)
			if err != nil {
				return fmt.Errorf("error updating file [%s]: %w", targetPath, err)
			}
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
