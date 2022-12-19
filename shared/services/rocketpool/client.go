package rocketpool

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/a8m/envsubst"
	"github.com/fatih/color"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"

	"github.com/alessio/shellescape"
	"github.com/blang/semver/v4"
	externalip "github.com/glendc/go-external-ip"
	"github.com/mitchellh/go-homedir"
	"github.com/rocket-pool/smartnode/addons/graffiti_wall_writer"
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
)

// Config
const (
	InstallerURL     string = "https://github.com/rocket-pool/smartnode-install/releases/download/%s/install.sh"
	UpdateTrackerURL string = "https://github.com/rocket-pool/smartnode-install/releases/download/%s/install-update-tracker.sh"

	LegacyBackupFolder       string = "old_config_backup"
	SettingsFile             string = "user-settings.yml"
	BackupSettingsFile       string = "user-settings-backup.yml"
	LegacyConfigFile         string = "config.yml"
	LegacySettingsFile       string = "settings.yml"
	PrometheusConfigTemplate string = "prometheus.tmpl"
	PrometheusFile           string = "prometheus.yml"

	APIContainerSuffix string = "_api"
	APIBinPath         string = "/go/bin/rocketpool"

	templatesDir                  string = "templates"
	overrideDir                   string = "override"
	runtimeDir                    string = "runtime"
	defaultFeeRecipientFile       string = "fr-default.tmpl"
	defaultNativeFeeRecipientFile string = "fr-default-env.tmpl"

	templateSuffix    string = ".tmpl"
	composeFileSuffix string = ".yml"

	nethermindPruneStarterCommand string = "dotnet /setup/NethermindPruneStarter/NethermindPruneStarter.dll"
	nethermindAdminUrl            string = "http://127.0.0.1:7434"

	DebugColor = color.FgYellow
)

// Get the external IP address. Try finding an IPv4 address first to:
// * Improve peer discovery and node performance
// * Avoid unnecessary container restarts caused by switching between IPv4 and IPv6
func getExternalIP() (net.IP, error) {
	// Try IPv4 first
	ip4Consensus := externalip.DefaultConsensus(nil, nil)
	ip4Consensus.UseIPProtocol(4)
	if ip, err := ip4Consensus.ExternalIP(); err == nil {
		return ip, nil
	}

	// Try IPv6 as fallback
	ip6Consensus := externalip.DefaultConsensus(nil, nil)
	ip6Consensus.UseIPProtocol(6)
	return ip6Consensus.ExternalIP()
}

// Rocket Pool client
type Client struct {
	configPath         string
	daemonPath         string
	maxFee             float64
	maxPrioFee         float64
	gasLimit           uint64
	customNonce        *big.Int
	client             *ssh.Client
	originalMaxFee     float64
	originalMaxPrioFee float64
	originalGasLimit   uint64
	debugPrint         bool
	ignoreSyncCheck    bool
	forceFallbacks     bool
}

// Create new Rocket Pool client from CLI context
func NewClientFromCtx(c *cli.Context) (*Client, error) {
	return NewClient(c.GlobalString("config-path"),
		c.GlobalString("daemon-path"),
		c.GlobalFloat64("maxFee"),
		c.GlobalFloat64("maxPrioFee"),
		c.GlobalUint64("gasLimit"),
		c.GlobalString("nonce"),
		c.GlobalBool("debug"))
}

// Create new Rocket Pool client
func NewClient(configPath string, daemonPath string, maxFee float64, maxPrioFee float64, gasLimit uint64, customNonce string, debug bool) (*Client, error) {

	// Initialize SSH client if configured for SSH
	var sshClient *ssh.Client
	var customNonceBigInt *big.Int = nil
	var success bool
	if customNonce != "" {
		customNonceBigInt, success = big.NewInt(0).SetString(customNonce, 0)
		if !success {
			return nil, fmt.Errorf("Invalid nonce: %s", customNonce)
		}
	}

	// Return client
	client := &Client{
		configPath:         os.ExpandEnv(configPath),
		daemonPath:         os.ExpandEnv(daemonPath),
		maxFee:             maxFee,
		maxPrioFee:         maxPrioFee,
		gasLimit:           gasLimit,
		originalMaxFee:     maxFee,
		originalMaxPrioFee: maxPrioFee,
		originalGasLimit:   gasLimit,
		customNonce:        customNonceBigInt,
		client:             sshClient,
		debugPrint:         debug,
		forceFallbacks:     false,
		ignoreSyncCheck:    false,
	}

	return client, nil

}

// Close client remote connection
func (c *Client) Close() {
	if c.client != nil {
		_ = c.client.Close()
	}
}

// Load the config
func (c *Client) LoadConfig() (*config.RocketPoolConfig, bool, error) {
	settingsFilePath := filepath.Join(c.configPath, SettingsFile)
	expandedPath, err := homedir.Expand(settingsFilePath)
	if err != nil {
		return nil, false, fmt.Errorf("error expanding settings file path: %w", err)
	}

	cfg, err := rp.LoadConfigFromFile(expandedPath)
	if err != nil {
		return nil, false, err
	}

	isNew := false
	if cfg == nil {
		cfg = config.NewRocketPoolConfig(c.configPath, c.daemonPath != "")
		isNew = true
	}
	return cfg, isNew, nil
}

// Load the backup config
func (c *Client) LoadBackupConfig() (*config.RocketPoolConfig, error) {
	settingsFilePath := filepath.Join(c.configPath, BackupSettingsFile)
	expandedPath, err := homedir.Expand(settingsFilePath)
	if err != nil {
		return nil, fmt.Errorf("error expanding backup settings file path: %w", err)
	}

	return rp.LoadConfigFromFile(expandedPath)
}

// Save the config
func (c *Client) SaveConfig(cfg *config.RocketPoolConfig) error {
	settingsFilePath := filepath.Join(c.configPath, SettingsFile)
	expandedPath, err := homedir.Expand(settingsFilePath)
	if err != nil {
		return err
	}
	return rp.SaveConfig(cfg, expandedPath)
}

// Remove the upgrade flag file
func (c *Client) RemoveUpgradeFlagFile() error {
	expandedPath, err := homedir.Expand(c.configPath)
	if err != nil {
		return err
	}
	return rp.RemoveUpgradeFlagFile(expandedPath)
}

// Returns whether or not this is the first run of the configurator since a previous installation
func (c *Client) IsFirstRun() (bool, error) {
	expandedPath, err := homedir.Expand(c.configPath)
	if err != nil {
		return false, fmt.Errorf("error expanding settings file path: %w", err)
	}
	return rp.IsFirstRun(expandedPath), nil
}

// Load the legacy config if one exists
func (c *Client) LoadLegacyConfigFromBackup() (*config.RocketPoolConfig, error) {
	// Check if the backup config file exists
	configPath, err := homedir.Expand(filepath.Join(c.configPath, LegacyBackupFolder, LegacyConfigFile))
	if err != nil {
		return nil, fmt.Errorf("Error expanding legacy config file path: %w", err)
	}
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		return nil, nil
	}

	// The backup config file exists, try loading the settings file
	settingsPath, err := homedir.Expand(filepath.Join(c.configPath, LegacyBackupFolder, LegacySettingsFile))
	if err != nil {
		return nil, fmt.Errorf("Error expanding legacy settings file path: %w", err)
	}
	_, err = os.Stat(settingsPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("Found a legacy config.yml file in the backup directory but not a legacy settings.yml file.")
	}

	// Migrate the old config to a new one
	newCfg, err := c.MigrateLegacyConfig(configPath, settingsPath)
	if err != nil {
		return nil, fmt.Errorf("Error migrating legacy config from a previous installation: %w", err)
	}
	return newCfg, nil
}

// Load the Prometheus template, do an environment variable substitution, and save it
func (c *Client) UpdatePrometheusConfiguration(settings map[string]string) error {
	prometheusTemplatePath, err := homedir.Expand(fmt.Sprintf("%s/%s", c.configPath, PrometheusConfigTemplate))
	if err != nil {
		return fmt.Errorf("Error expanding Prometheus template path: %w", err)
	}

	prometheusConfigPath, err := homedir.Expand(fmt.Sprintf("%s/%s", c.configPath, PrometheusFile))
	if err != nil {
		return fmt.Errorf("Error expanding Prometheus config file path: %w", err)
	}

	// Set the environment variables defined in the user settings for metrics
	oldValues := map[string]string{}
	for varName, varValue := range settings {
		oldValues[varName] = os.Getenv(varName)
		os.Setenv(varName, varValue)
	}

	// Read and substitute the template
	contents, err := envsubst.ReadFile(prometheusTemplatePath)
	if err != nil {
		return fmt.Errorf("Error reading and substituting Prometheus configuration template: %w", err)
	}

	// Unset the env vars
	for name, value := range oldValues {
		os.Setenv(name, value)
	}

	// Write the actual Prometheus config file
	err = ioutil.WriteFile(prometheusConfigPath, contents, 0664)
	if err != nil {
		return fmt.Errorf("Could not write Prometheus config file to %s: %w", shellescape.Quote(prometheusConfigPath), err)
	}
	err = os.Chmod(prometheusConfigPath, 0664)
	if err != nil {
		return fmt.Errorf("Could not set Prometheus config file permissions: %w", shellescape.Quote(prometheusConfigPath), err)
	}

	return nil
}

// Migrate a legacy configuration (pre-v1.3) to a modern post-v1.3 one
func (c *Client) MigrateLegacyConfig(legacyConfigFilePath string, legacySettingsFilePath string) (*config.RocketPoolConfig, error) {

	// Check if the files exist
	_, err := os.Stat(legacyConfigFilePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("Legacy configuration file [%s] does not exist or is not accessible.", legacyConfigFilePath)
	}
	_, err = os.Stat(legacySettingsFilePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("Legacy settings file [%s] does not exist or is not accessible.", legacySettingsFilePath)
	}

	// Load the legacy config
	isNative := (c.daemonPath != "")
	legacyCfg, err := c.LoadMergedConfig_Legacy(legacyConfigFilePath, legacySettingsFilePath)
	if err != nil {
		return nil, fmt.Errorf("error loading legacy configuration: %w", err)
	}
	cfg := config.NewRocketPoolConfig(c.configPath, isNative)

	// Do the conversion

	// Network
	chainID := legacyCfg.Chains.Eth1.ChainID
	var network cfgtypes.Network
	switch chainID {
	case "1":
		network = cfgtypes.Network_Mainnet
	case "5":
		network = cfgtypes.Network_Prater
	default:
		return nil, fmt.Errorf("legacy config had an unknown chain ID [%s]", chainID)
	}
	cfg.Smartnode.Network.Value = network

	// Migrate the EC
	err = c.migrateProviderInfo(legacyCfg.Chains.Eth1.Provider, legacyCfg.Chains.Eth1.WsProvider, "eth1", &cfg.ExecutionClientMode, &cfg.ExecutionCommon.HttpPort, &cfg.ExecutionCommon.WsPort, &cfg.ExternalExecution.HttpUrl, &cfg.ExternalExecution.WsUrl)
	if err != nil {
		return nil, fmt.Errorf("error migrating eth1 provider info: %w", err)
	}

	err = c.migrateEcSelection(legacyCfg.Chains.Eth1.Client.Selected, &cfg.ExecutionClient, &cfg.ExecutionClientMode)
	if err != nil {
		return nil, fmt.Errorf("error migrating eth1 client selection: %w", err)
	}

	err = c.migrateEth1Params(legacyCfg.Chains.Eth1.Client.Selected, network, legacyCfg.Chains.Eth1.Client.Params, cfg.ExecutionCommon, cfg.Geth, cfg.ExternalExecution)
	if err != nil {
		return nil, fmt.Errorf("error migrating eth1 params: %w", err)
	}

	// Disable fallback migration which didn't exist in the same sense with v1.2.x
	cfg.UseFallbackClients.Value = false

	// Migrate the CC
	ccProvider := legacyCfg.Chains.Eth2.Provider
	ccMode, ccPort, err := c.getLegacyProviderInfo(ccProvider, "eth2")
	if err != nil {
		return nil, fmt.Errorf("error migrating eth2 provider info: %w", err)
	}
	cfg.ConsensusClientMode.Value = ccMode

	selectedCC := legacyCfg.Chains.Eth2.Client.Selected
	if ccMode == cfgtypes.Mode_Local {
		err = c.migrateCcSelection(selectedCC, &cfg.ConsensusClient)
		if err != nil {
			return nil, fmt.Errorf("error migrating local eth2 client selection: %w", err)
		}
		cfg.ConsensusCommon.ApiPort.Value = ccPort
	} else {
		err = c.migrateCcSelection(selectedCC, &cfg.ExternalConsensusClient)
		if err != nil {
			return nil, fmt.Errorf("error migrating external eth2 client selection: %w", err)
		}
		cfg.ExternalLighthouse.HttpUrl.Value = ccProvider
		cfg.ExternalPrysm.HttpUrl.Value = ccProvider
		cfg.ExternalTeku.HttpUrl.Value = ccProvider
	}

	for _, param := range legacyCfg.Chains.Eth2.Client.Params {
		switch param.Env {
		case config.CustomGraffitiEnvVar:
			cfg.ConsensusCommon.Graffiti.Value = param.Value
			cfg.ExternalLighthouse.Graffiti.Value = param.Value
			cfg.ExternalPrysm.Graffiti.Value = param.Value
			cfg.ExternalTeku.Graffiti.Value = param.Value
		case "ETH2_MAX_PEERS":
			switch cfg.ConsensusClient.Value.(cfgtypes.ConsensusClient) {
			case cfgtypes.ConsensusClient_Lighthouse:
				convertUintParam(param, &cfg.Lighthouse.MaxPeers, network, 16)
			case cfgtypes.ConsensusClient_Nimbus:
				convertUintParam(param, &cfg.Nimbus.MaxPeers, network, 16)
			case cfgtypes.ConsensusClient_Prysm:
				convertUintParam(param, &cfg.Prysm.MaxPeers, network, 16)
			case cfgtypes.ConsensusClient_Teku:
				convertUintParam(param, &cfg.Teku.MaxPeers, network, 16)
			}
		case "ETH2_P2P_PORT":
			convertUintParam(param, &cfg.ConsensusCommon.P2pPort, network, 16)
		case "ETH2_CHECKPOINT_SYNC_URL":
			cfg.ConsensusCommon.CheckpointSyncProvider.Value = param.Value
		case "ETH2_DOPPELGANGER_DETECTION":
			if param.Value == "y" {
				cfg.ConsensusCommon.DoppelgangerDetection.Value = true
				cfg.ExternalLighthouse.DoppelgangerDetection.Value = true
				cfg.ExternalPrysm.DoppelgangerDetection.Value = true
			} else {
				cfg.ConsensusCommon.DoppelgangerDetection.Value = false
				cfg.ExternalLighthouse.DoppelgangerDetection.Value = false
				cfg.ExternalPrysm.DoppelgangerDetection.Value = false
			}
		case "ETH2_RPC_PORT":
			convertUintParam(param, &cfg.Prysm.RpcPort, network, 16)
			port := cfg.Prysm.RpcPort.Value.(uint16)
			cfg.Prysm.RpcPort.Value = uint16(port)
			externalPrysmUrl := strings.Replace(ccProvider, fmt.Sprintf(":%d", ccPort), fmt.Sprintf(":%d", port), 1)
			cfg.ExternalPrysm.JsonRpcUrl.Value = externalPrysmUrl
		}
	}

	// Migrate metrics
	cfg.EnableMetrics.Value = legacyCfg.Metrics.Enabled
	for _, param := range legacyCfg.Metrics.Settings {
		switch param.Env {
		case "ETH2_METRICS_PORT":
			convertUintParam(param, &cfg.BnMetricsPort, network, 16)
		case "VALIDATOR_METRICS_PORT":
			convertUintParam(param, &cfg.VcMetricsPort, network, 16)
		case "NODE_METRICS_PORT":
			convertUintParam(param, &cfg.NodeMetricsPort, network, 16)
		case "EXPORTER_METRICS_PORT":
			convertUintParam(param, &cfg.ExporterMetricsPort, network, 16)
		case "WATCHTOWER_METRICS_PORT":
			convertUintParam(param, &cfg.WatchtowerMetricsPort, network, 16)
		case "PROMETHEUS_PORT":
			convertUintParam(param, &cfg.Prometheus.Port, network, 16)
		case "GRAFANA_PORT":
			convertUintParam(param, &cfg.Grafana.Port, network, 16)
		}
	}

	// Top-level parameters
	cfg.ReconnectDelay.Value = legacyCfg.Chains.Eth1.ReconnectDelay
	if cfg.ReconnectDelay.Value == "" {
		cfg.ReconnectDelay.Value = cfg.ReconnectDelay.Default[cfgtypes.Network_All]
	}

	// Smartnode settings
	cfg.Smartnode.ProjectName.Value = legacyCfg.Smartnode.ProjectName
	cfg.Smartnode.ManualMaxFee.Value = legacyCfg.Smartnode.MaxFee
	cfg.Smartnode.PriorityFee.Value = legacyCfg.Smartnode.MaxPriorityFee
	cfg.Smartnode.MinipoolStakeGasThreshold.Value = legacyCfg.Smartnode.MinipoolStakeGasThreshold

	// Docker images
	for _, option := range legacyCfg.Chains.Eth1.Client.Options {
		if option.ID == "geth" {
			cfg.Geth.ContainerTag.Value = option.Image
		}
	}
	for _, option := range legacyCfg.Chains.Eth2.Client.Options {
		switch option.ID {
		case "lighthouse":
			cfg.Lighthouse.ContainerTag.Value = option.Image
			cfg.ExternalLighthouse.ContainerTag.Value = option.Image
		case "nimbus":
			cfg.Nimbus.ContainerTag.Value = option.Image
		case "prysm":
			cfg.Prysm.BnContainerTag.Value = option.BeaconImage
			cfg.Prysm.VcContainerTag.Value = option.ValidatorImage
			cfg.ExternalPrysm.ContainerTag.Value = option.ValidatorImage
		case "teku":
			cfg.Teku.ContainerTag.Value = option.Image
			cfg.ExternalTeku.ContainerTag.Value = option.Image
		}
	}

	// Handle native mode
	cfg.Native.EcHttpUrl.Value = legacyCfg.Chains.Eth1.Provider
	cfg.Native.CcHttpUrl.Value = legacyCfg.Chains.Eth2.Provider
	c.migrateCcSelection(legacyCfg.Chains.Eth2.Client.Selected, &cfg.Native.ConsensusClient)
	cfg.Native.ValidatorRestartCommand.Value = legacyCfg.Smartnode.ValidatorRestartCommand
	cfg.Smartnode.DataPath.Value = filepath.Join(c.configPath, "data")

	return cfg, nil

}

// Install the Rocket Pool service
func (c *Client) InstallService(verbose, noDeps bool, network, version, path string, dataPath string) error {

	// Get installation script downloader type
	downloader, err := c.getDownloader()
	if err != nil {
		return err
	}

	// Get installation script flags
	flags := []string{
		"-n", fmt.Sprintf("%s", shellescape.Quote(network)),
		"-v", fmt.Sprintf("%s", shellescape.Quote(version)),
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

	// Initialize installation command
	cmd, err := c.newCommand(fmt.Sprintf("%s %s | sh -s -- %s", downloader, fmt.Sprintf(InstallerURL, version), strings.Join(flags, " ")))
	if err != nil {
		return err
	}
	defer func() {
		_ = cmd.Close()
	}()

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

	// Get installation script downloader type
	downloader, err := c.getDownloader()
	if err != nil {
		return err
	}

	// Get installation script flags
	flags := []string{
		"-v", fmt.Sprintf("%s", shellescape.Quote(version)),
	}

	// Initialize installation command
	cmd, err := c.newCommand(fmt.Sprintf("%s %s | sh -s -- %s", downloader, fmt.Sprintf(UpdateTrackerURL, version), strings.Join(flags, " ")))
	if err != nil {
		return err
	}
	defer func() {
		_ = cmd.Close()
	}()

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
		return fmt.Errorf("Could not install Rocket Pool update tracker: %s", errMessage)
	}
	return nil

}

// Start the Rocket Pool service
func (c *Client) StartService(composeFiles []string) error {

	// Start the API container first
	cmd, err := c.compose([]string{}, "up -d")
	if err != nil {
		return fmt.Errorf("error creating compose command for API container: %w", err)
	}
	err = c.printOutput(cmd)
	if err != nil {
		return fmt.Errorf("error starting API container: %w", err)
	}

	// Start all of the containers
	cmd, err = c.compose(composeFiles, "up -d --remove-orphans")
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
	var cmd string
	if c.daemonPath == "" {
		containerName, err := c.getAPIContainerName()
		if err != nil {
			return "", err
		}
		cmd = fmt.Sprintf("docker exec %s %s --version", shellescape.Quote(containerName), shellescape.Quote(APIBinPath))
	} else {
		cmd = fmt.Sprintf("%s --version", shellescape.Quote(c.daemonPath))
	}
	versionBytes, err := c.readOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("Could not get Rocket Pool service version: %w", err)
	}

	// Get the version string
	outputString := string(versionBytes)
	elements := strings.Fields(outputString) // Split on whitespace
	if len(elements) < 1 {
		return "", fmt.Errorf("Could not parse Rocket Pool service version number from output '%s'", outputString)
	}
	versionString := elements[len(elements)-1]

	// Make sure it's a semantic version
	version, err := semver.Make(versionString)
	if err != nil {
		return "", fmt.Errorf("Could not parse Rocket Pool service version number from output '%s': %w", outputString, err)
	}

	// Return the parsed semantic version (extra safety)
	return version.String(), nil

}

// Increments the custom nonce parameter.
// This is used for calls that involve multiple transactions, so they don't all have the same nonce.
func (c *Client) IncrementCustomNonce() {
	c.customNonce.Add(c.customNonce, big.NewInt(1))
}

// Get the current Docker image used by the given container
func (c *Client) GetDockerImage(container string) (string, error) {

	cmd := fmt.Sprintf("docker container inspect --format={{.Config.Image}} %s", container)
	image, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(image)), nil

}

// Get the current Docker image used by the given container
func (c *Client) GetDockerStatus(container string) (string, error) {

	cmd := fmt.Sprintf("docker container inspect --format={{.State.Status}} %s", container)
	status, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(status)), nil

}

// Get the time that the given container shut down
func (c *Client) GetDockerContainerShutdownTime(container string) (time.Time, error) {

	cmd := fmt.Sprintf("docker container inspect --format={{.State.FinishedAt}} %s", container)
	finishTimeBytes, err := c.readOutput(cmd)
	if err != nil {
		return time.Time{}, err
	}

	finishTime, err := time.Parse(time.RFC3339, strings.TrimSpace(string(finishTimeBytes)))
	if err != nil {
		return time.Time{}, fmt.Errorf("Error parsing validator container exit time [%s]: %w", string(finishTimeBytes), err)
	}

	return finishTime, nil

}

// Shut down a container
func (c *Client) StopContainer(container string) (string, error) {

	cmd := fmt.Sprintf("docker stop %s", container)
	output, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil

}

// Start a container
func (c *Client) StartContainer(container string) (string, error) {

	cmd := fmt.Sprintf("docker start %s", container)
	output, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil

}

// Restart a container
func (c *Client) RestartContainer(container string) (string, error) {

	cmd := fmt.Sprintf("docker restart %s", container)
	output, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil

}

// Deletes a container
func (c *Client) RemoveContainer(container string) (string, error) {

	cmd := fmt.Sprintf("docker rm %s", container)
	output, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil

}

// Deletes a container
func (c *Client) DeleteVolume(volume string) (string, error) {

	cmd := fmt.Sprintf("docker volume rm %s", volume)
	output, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil

}

// Gets the absolute file path of the client volume
func (c *Client) GetClientVolumeSource(container string, volumeTarget string) (string, error) {

	cmd := fmt.Sprintf("docker container inspect --format='{{range .Mounts}}{{if eq \"%s\" .Destination}}{{.Source}}{{end}}{{end}}' %s", volumeTarget, container)
	output, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// Gets the name of the client volume
func (c *Client) GetClientVolumeName(container string, volumeTarget string) (string, error) {

	cmd := fmt.Sprintf("docker container inspect --format='{{range .Mounts}}{{if eq \"%s\" .Destination}}{{.Name}}{{end}}{{end}}' %s", volumeTarget, container)
	output, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// Gets the disk usage of the given volume
func (c *Client) GetVolumeSize(volumeName string) (string, error) {

	cmd := fmt.Sprintf("docker system df -v --format='{{range .Volumes}}{{if eq \"%s\" .Name}}{{.Size}}{{end}}{{end}}'", volumeName)
	output, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
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

// Get the gas settings
func (c *Client) GetGasSettings() (float64, float64, uint64) {
	return c.maxFee, c.maxPrioFee, c.gasLimit
}

// Get the gas fees
func (c *Client) AssignGasSettings(maxFee float64, maxPrioFee float64, gasLimit uint64) {
	c.maxFee = maxFee
	c.maxPrioFee = maxPrioFee
	c.gasLimit = gasLimit
}

// Set the flags for ignoring EC and CC sync checks and forcing fallbacks to prevent unnecessary duplication of effort by the API during CLI commands
func (c *Client) SetClientStatusFlags(ignoreSyncCheck bool, forceFallbacks bool) {
	c.ignoreSyncCheck = ignoreSyncCheck
	c.forceFallbacks = forceFallbacks
}

// Get the provider mode and port from a legacy config's provider URL
func (c *Client) migrateProviderInfo(provider string, wsProvider string, localHostname string, clientMode *cfgtypes.Parameter, httpPortParam *cfgtypes.Parameter, wsPortParam *cfgtypes.Parameter, externalHttpUrlParam *cfgtypes.Parameter, externalWsUrlParam *cfgtypes.Parameter) error {

	// Get HTTP provider
	mode, port, err := c.getLegacyProviderInfo(provider, localHostname)
	if err != nil {
		return fmt.Errorf("error parsing %s provider: %w", localHostname, err)
	}

	// Set the mode, provider, port, and/or URL
	clientMode.Value = mode
	if mode == cfgtypes.Mode_Local {
		httpPortParam.Value = port
	} else {
		externalHttpUrlParam.Value = provider
	}

	// Get the websocket provider
	if wsProvider != "" {
		_, wsPort, err := c.getLegacyProviderInfo(wsProvider, localHostname)
		if err != nil {
			return fmt.Errorf("error parsing %s websocket provider: %w", localHostname, err)
		}
		if mode == cfgtypes.Mode_Local {
			wsPortParam.Value = wsPort
		} else {
			externalWsUrlParam.Value = wsProvider
		}
	}

	return nil

}

// Get the provider mode and port from a legacy config's provider URL
func (c *Client) getLegacyProviderInfo(provider string, localHostname string) (cfgtypes.Mode, uint16, error) {

	providerUrl, err := url.Parse(provider)
	if err != nil {
		return cfgtypes.Mode_Unknown, 0, fmt.Errorf("error parsing %s provider: %w", localHostname, err)
	}

	var mode cfgtypes.Mode
	if providerUrl.Hostname() == localHostname {
		// This is Docker mode
		mode = cfgtypes.Mode_Local
	} else {
		// This is Hybrid mode
		mode = cfgtypes.Mode_External
	}

	var port uint16
	portString := providerUrl.Port()
	if portString == "" {
		switch providerUrl.Scheme {
		case "http", "ws":
			port = 80
		case "https", "wss":
			port = 443
		default:
			return cfgtypes.Mode_Unknown, 0, fmt.Errorf("provider [%s] doesn't provide port info and it can't be inferred from the scheme", provider)
		}
	} else {
		parsedPort, err := strconv.ParseUint(portString, 0, 16)
		if err != nil {
			return cfgtypes.Mode_Unknown, 0, fmt.Errorf("invalid port [%s] in %s provider [%s]", portString, localHostname, provider)
		}
		port = uint16(parsedPort)
	}

	return mode, port, nil

}

// Sets a modern config's selected EC / mode based on a legacy config
func (c *Client) migrateEcSelection(legacySelectedClient string, ecParam *cfgtypes.Parameter, ecModeParam *cfgtypes.Parameter) error {
	// EC selection
	switch legacySelectedClient {
	case "geth":
		ecParam.Value = cfgtypes.ExecutionClient_Geth
	case "infura":
		ecParam.Value = cfgtypes.ExecutionClient_Geth
	case "pocket":
		ecParam.Value = cfgtypes.ExecutionClient_Geth
	case "custom":
		ecModeParam.Value = cfgtypes.Mode_External
	case "":
		break
	default:
		return fmt.Errorf("unknown eth1 client [%s]", legacySelectedClient)
	}

	return nil
}

// Sets a modern config's selected CC / mode based on a legacy config
func (c *Client) migrateCcSelection(legacySelectedClient string, ccParam *cfgtypes.Parameter) error {
	// CC selection
	switch legacySelectedClient {
	case "lighthouse":
		ccParam.Value = cfgtypes.ConsensusClient_Lighthouse
	case "nimbus":
		ccParam.Value = cfgtypes.ConsensusClient_Nimbus
	case "prysm":
		ccParam.Value = cfgtypes.ConsensusClient_Prysm
	case "teku":
		ccParam.Value = cfgtypes.ConsensusClient_Teku
	default:
		return fmt.Errorf("unknown eth2 client [%s]", legacySelectedClient)
	}

	return nil
}

// Migrates the parameters from a legacy eth1 config to a modern one
func (c *Client) migrateEth1Params(client string, network cfgtypes.Network, params []config.UserParam, ecCommon *config.ExecutionCommonConfig, geth *config.GethConfig, externalEc *config.ExternalExecutionConfig) error {
	for _, param := range params {
		switch param.Env {
		case "ETHSTATS_LABEL":
			if ecCommon != nil {
				ecCommon.EthstatsLabel.Value = param.Value
			}
		case "ETHSTATS_LOGIN":
			if ecCommon != nil {
				ecCommon.EthstatsLogin.Value = param.Value
			}
		case "GETH_CACHE_SIZE":
			if geth != nil {
				convertUintParam(param, &geth.CacheSize, network, 0)
			}
		case "GETH_MAX_PEERS":
			if geth != nil {
				convertUintParam(param, &geth.MaxPeers, network, 16)
			}
		case "ETH1_P2P_PORT":
			if ecCommon != nil {
				convertUintParam(param, &ecCommon.P2pPort, network, 16)
			}
		case "HTTP_PROVIDER_URL":
			if client == "custom" {
				externalEc.HttpUrl.Value = param.Value
			}
		case "WS_PROVIDER_URL":
			if client == "custom" {
				externalEc.WsUrl.Value = param.Value
			}
		}
	}

	return nil
}

// Stores a legacy parameter's value in a new parameter, replacing blank values with the appropriate default.
func convertUintParam(oldParam config.UserParam, newParam *cfgtypes.Parameter, network cfgtypes.Network, bitsize int) error {
	if newParam == nil {
		return nil
	}

	if oldParam.Value == "" {
		valIface, err := newParam.GetDefault(network)
		if err != nil {
			return fmt.Errorf("failed to get default for param [%s] on network [%v]: %w", newParam.ID, network, err)
		}
		newParam.Value = valIface
	} else {
		value, err := strconv.ParseUint(oldParam.Value, 0, bitsize)
		if err != nil {
			return fmt.Errorf("invalid legacy setting [%s] for param [%s]: %w", oldParam.Value, newParam.ID, err)
		}
		switch bitsize {
		case 0:
			newParam.Value = uint(value)
		case 16:
			newParam.Value = uint16(value)
		case 32:
			newParam.Value = uint32(value)
		case 64:
			newParam.Value = value
		default:
			return fmt.Errorf("unexpected bitsize %d", bitsize)
		}
	}

	return nil
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
	settings["EXTERNAL_IP"] = shellescape.Quote(externalIP)

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

	// API
	contents, err := envsubst.ReadFile(filepath.Join(templatesFolder, config.ApiContainerName+templateSuffix))
	if err != nil {
		return []string{}, fmt.Errorf("error reading and substituting API container template: %w", err)
	}
	apiComposePath := filepath.Join(runtimeFolder, config.ApiContainerName+composeFileSuffix)
	err = ioutil.WriteFile(apiComposePath, contents, 0664)
	if err != nil {
		return []string{}, fmt.Errorf("could not write API container file to %s: %w", apiComposePath, err)
	}
	deployedContainers = append(deployedContainers, apiComposePath)
	deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, config.ApiContainerName+composeFileSuffix))

	// Node
	contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, config.NodeContainerName+templateSuffix))
	if err != nil {
		return []string{}, fmt.Errorf("error reading and substituting node container template: %w", err)
	}
	nodeComposePath := filepath.Join(runtimeFolder, config.NodeContainerName+composeFileSuffix)
	err = ioutil.WriteFile(nodeComposePath, contents, 0664)
	if err != nil {
		return []string{}, fmt.Errorf("could not write node container file to %s: %w", nodeComposePath, err)
	}
	deployedContainers = append(deployedContainers, nodeComposePath)
	deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, config.NodeContainerName+composeFileSuffix))

	// Watchtower
	contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, config.WatchtowerContainerName+templateSuffix))
	if err != nil {
		return []string{}, fmt.Errorf("error reading and substituting watchtower container template: %w", err)
	}
	watchtowerComposePath := filepath.Join(runtimeFolder, config.WatchtowerContainerName+composeFileSuffix)
	err = ioutil.WriteFile(watchtowerComposePath, contents, 0664)
	if err != nil {
		return []string{}, fmt.Errorf("could not write watchtower container file to %s: %w", watchtowerComposePath, err)
	}
	deployedContainers = append(deployedContainers, watchtowerComposePath)
	deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, config.WatchtowerContainerName+composeFileSuffix))

	// Validator
	contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, config.ValidatorContainerName+templateSuffix))
	if err != nil {
		return []string{}, fmt.Errorf("error reading and substituting validator container template: %w", err)
	}
	validatorComposePath := filepath.Join(runtimeFolder, config.ValidatorContainerName+composeFileSuffix)
	err = ioutil.WriteFile(validatorComposePath, contents, 0664)
	if err != nil {
		return []string{}, fmt.Errorf("could not write validator container file to %s: %w", validatorComposePath, err)
	}
	deployedContainers = append(deployedContainers, validatorComposePath)
	deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, config.ValidatorContainerName+composeFileSuffix))

	// Check the EC mode to see if it needs to be deployed
	if cfg.ExecutionClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
		contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, config.Eth1ContainerName+templateSuffix))
		if err != nil {
			return []string{}, fmt.Errorf("error reading and substituting execution client container template: %w", err)
		}
		eth1ComposePath := filepath.Join(runtimeFolder, config.Eth1ContainerName+composeFileSuffix)
		err = ioutil.WriteFile(eth1ComposePath, contents, 0664)
		if err != nil {
			return []string{}, fmt.Errorf("could not write execution client container file to %s: %w", eth1ComposePath, err)
		}
		deployedContainers = append(deployedContainers, eth1ComposePath)
		deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, config.Eth1ContainerName+composeFileSuffix))
	}

	// Check the Consensus mode
	if cfg.ConsensusClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
		contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, config.Eth2ContainerName+templateSuffix))
		if err != nil {
			return []string{}, fmt.Errorf("error reading and substituting consensus client container template: %w", err)
		}
		eth2ComposePath := filepath.Join(runtimeFolder, config.Eth2ContainerName+composeFileSuffix)
		err = ioutil.WriteFile(eth2ComposePath, contents, 0664)
		if err != nil {
			return []string{}, fmt.Errorf("could not write consensus client container file to %s: %w", eth2ComposePath, err)
		}
		deployedContainers = append(deployedContainers, eth2ComposePath)
		deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, config.Eth2ContainerName+composeFileSuffix))
	}

	// Check the metrics containers
	if cfg.EnableMetrics.Value == true {
		// Grafana
		contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, config.GrafanaContainerName+templateSuffix))
		if err != nil {
			return []string{}, fmt.Errorf("error reading and substituting Grafana container template: %w", err)
		}
		grafanaComposePath := filepath.Join(runtimeFolder, config.GrafanaContainerName+composeFileSuffix)
		err = ioutil.WriteFile(grafanaComposePath, contents, 0664)
		if err != nil {
			return []string{}, fmt.Errorf("could not write Grafana container file to %s: %w", grafanaComposePath, err)
		}
		deployedContainers = append(deployedContainers, grafanaComposePath)
		deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, config.GrafanaContainerName+composeFileSuffix))

		// Node exporter
		contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, config.ExporterContainerName+templateSuffix))
		if err != nil {
			return []string{}, fmt.Errorf("error reading and substituting Node Exporter container template: %w", err)
		}
		exporterComposePath := filepath.Join(runtimeFolder, config.ExporterContainerName+composeFileSuffix)
		err = ioutil.WriteFile(exporterComposePath, contents, 0664)
		if err != nil {
			return []string{}, fmt.Errorf("could not write Node Exporter container file to %s: %w", exporterComposePath, err)
		}
		deployedContainers = append(deployedContainers, exporterComposePath)
		deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, config.ExporterContainerName+composeFileSuffix))

		// Prometheus
		contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, config.PrometheusContainerName+templateSuffix))
		if err != nil {
			return []string{}, fmt.Errorf("error reading and substituting Prometheus container template: %w", err)
		}
		prometheusComposePath := filepath.Join(runtimeFolder, config.PrometheusContainerName+composeFileSuffix)
		err = ioutil.WriteFile(prometheusComposePath, contents, 0664)
		if err != nil {
			return []string{}, fmt.Errorf("could not write Prometheus container file to %s: %w", prometheusComposePath, err)
		}
		deployedContainers = append(deployedContainers, prometheusComposePath)
		deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, config.PrometheusContainerName+composeFileSuffix))
	}

	// Check MEV-Boost
	if cfg.EnableMevBoost.Value == true && cfg.MevBoost.Mode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
		contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, config.MevBoostContainerName+templateSuffix))
		if err != nil {
			return []string{}, fmt.Errorf("error reading and substituting MEV-Boost container template: %w", err)
		}
		mevBoostComposePath := filepath.Join(runtimeFolder, config.MevBoostContainerName+composeFileSuffix)
		err = ioutil.WriteFile(mevBoostComposePath, contents, 0664)
		if err != nil {
			return []string{}, fmt.Errorf("could not write MEV-Boost container file to %s: %w", mevBoostComposePath, err)
		}
		deployedContainers = append(deployedContainers, mevBoostComposePath)
		deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, config.MevBoostContainerName+composeFileSuffix))
	}

	// Create the custom keys dir
	customKeyDir, err := homedir.Expand(filepath.Join(cfg.Smartnode.DataPath.Value.(string), "custom-keys"))
	if err != nil {
		fmt.Printf("%sWARNING: Couldn't expand the custom validator key directory (%s). You will not be able to recover any minipool keys you created outside of the Smartnode until you create the folder manually.%s\n", colorYellow, err.Error(), colorReset)
		return deployedContainers, nil
	}
	err = os.MkdirAll(customKeyDir, 0775)
	if err != nil {
		fmt.Printf("%sWARNING: Couldn't create the custom validator key directory (%s). You will not be able to recover any minipool keys you created outside of the Smartnode until you create the folder [%s] manually.%s\n", colorYellow, err.Error(), customKeyDir, colorReset)
	}

	// Create the rewards file dir
	rewardsFilePath, err := homedir.Expand(cfg.Smartnode.GetRewardsTreePath(0, false))
	if err != nil {
		fmt.Printf("%sWARNING: Couldn't expand the rewards tree file directory (%s). You will not be able to view or claim your rewards until you create the folder manually.%s\n", colorYellow, err.Error(), colorReset)
		return deployedContainers, nil
	}
	rewardsFileDir := filepath.Dir(rewardsFilePath)
	err = os.MkdirAll(rewardsFileDir, 0775)
	if err != nil {
		fmt.Printf("%sWARNING: Couldn't create the rewards tree file directory (%s). You will not be able to view or claim your rewards until you create the folder [%s] manually.%s\n", colorYellow, err.Error(), rewardsFileDir, colorReset)
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
		err = ioutil.WriteFile(composePath, contents, 0664)
		if err != nil {
			return []string{}, fmt.Errorf("could not write GWW addon container file to %s: %w", composePath, err)
		}
		deployedContainers = append(deployedContainers, composePath)
		deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, graffiti_wall_writer.GraffitiWallWriterContainerName+composeFileSuffix))
	}

	return deployedContainers, nil

}

// Call the Rocket Pool API
func (c *Client) callAPI(args string, otherArgs ...string) ([]byte, error) {
	// Sanitize and parse the args
	ignoreSyncCheckFlag, forceFallbackECFlag, args := c.getApiCallArgs(args, otherArgs...)

	// Create the command to run
	var cmd string
	if c.daemonPath == "" {
		containerName, err := c.getAPIContainerName()
		if err != nil {
			return []byte{}, err
		}
		cmd = fmt.Sprintf("docker exec %s %s %s %s %s %s api %s", shellescape.Quote(containerName), shellescape.Quote(APIBinPath), ignoreSyncCheckFlag, forceFallbackECFlag, c.getGasOpts(), c.getCustomNonce(), args)
	} else {
		cmd = fmt.Sprintf("%s --settings %s %s %s %s %s api %s",
			c.daemonPath,
			shellescape.Quote(fmt.Sprintf("%s/%s", c.configPath, SettingsFile)),
			ignoreSyncCheckFlag,
			forceFallbackECFlag,
			c.getGasOpts(),
			c.getCustomNonce(),
			args)
	}

	// Run the command
	return c.runApiCall(cmd)
}

// Call the Rocket Pool API with some custom environment variables
func (c *Client) callAPIWithEnvVars(envVars map[string]string, args string, otherArgs ...string) ([]byte, error) {
	// Sanitize and parse the args
	ignoreSyncCheckFlag, forceFallbackECFlag, args := c.getApiCallArgs(args, otherArgs...)

	// Create the command to run
	var cmd string
	if c.daemonPath == "" {
		envArgs := ""
		for key, value := range envVars {
			os.Setenv(key, shellescape.Quote(value))
			envArgs += fmt.Sprintf("-e %s ", key)
		}
		containerName, err := c.getAPIContainerName()
		if err != nil {
			return []byte{}, err
		}
		cmd = fmt.Sprintf("docker exec %s %s %s %s %s %s %s api %s", envArgs, shellescape.Quote(containerName), shellescape.Quote(APIBinPath), ignoreSyncCheckFlag, forceFallbackECFlag, c.getGasOpts(), c.getCustomNonce(), args)
	} else {
		envArgs := ""
		for key, value := range envVars {
			envArgs += fmt.Sprintf("%s=%s ", key, shellescape.Quote(value))
		}
		cmd = fmt.Sprintf("%s %s --settings %s %s %s %s %s api %s",
			envArgs,
			c.daemonPath,
			shellescape.Quote(fmt.Sprintf("%s/%s", c.configPath, SettingsFile)),
			ignoreSyncCheckFlag,
			forceFallbackECFlag,
			c.getGasOpts(),
			c.getCustomNonce(),
			args)
	}

	// Run the command
	return c.runApiCall(cmd)
}

func (c *Client) getApiCallArgs(args string, otherArgs ...string) (string, string, string) {
	// Sanitize arguments
	var sanitizedArgs []string
	for _, arg := range strings.Fields(args) {
		sanitizedArg := shellescape.Quote(arg)
		sanitizedArgs = append(sanitizedArgs, sanitizedArg)
	}
	args = strings.Join(sanitizedArgs, " ")
	if len(otherArgs) > 0 {
		for _, arg := range otherArgs {
			sanitizedArg := shellescape.Quote(arg)
			args += fmt.Sprintf(" %s", sanitizedArg)
		}
	}

	ignoreSyncCheckFlag := ""
	if c.ignoreSyncCheck {
		ignoreSyncCheckFlag = "--ignore-sync-check"
	}
	forceFallbacksFlag := ""
	if c.forceFallbacks {
		forceFallbacksFlag = "--force-fallbacks"
	}

	return ignoreSyncCheckFlag, forceFallbacksFlag, args
}

func (c *Client) runApiCall(cmd string) ([]byte, error) {
	if c.debugPrint {
		fmt.Println("To API:")
		fmt.Println(cmd)
	}

	output, err := c.readOutput(cmd)

	if c.debugPrint {
		if output != nil {
			fmt.Println("API Out:")
			fmt.Println(string(output))
		}
		if err != nil {
			fmt.Println("API Err:")
			fmt.Println(err.Error())
		}
	}

	// Reset the gas settings after the call
	c.maxFee = c.originalMaxFee
	c.maxPrioFee = c.originalMaxPrioFee
	c.gasLimit = c.originalGasLimit

	return output, err
}

// Get the API container name
func (c *Client) getAPIContainerName() (string, error) {
	cfg, _, err := c.LoadConfig()
	if err != nil {
		return "", err
	}
	if cfg.Smartnode.ProjectName.Value == "" {
		return "", errors.New("Rocket Pool docker project name not set")
	}
	return cfg.Smartnode.ProjectName.Value.(string) + APIContainerSuffix, nil
}

// Get gas price & limit flags
func (c *Client) getGasOpts() string {
	var opts string
	opts += fmt.Sprintf("--maxFee %f ", c.maxFee)
	opts += fmt.Sprintf("--maxPrioFee %f ", c.maxPrioFee)
	opts += fmt.Sprintf("--gasLimit %d ", c.gasLimit)
	return opts
}

func (c *Client) getCustomNonce() string {
	// Set the custom nonce
	nonce := ""
	if c.customNonce != nil {
		nonce = fmt.Sprintf("--nonce %s", c.customNonce.String())
	}
	return nonce
}

// Get the first downloader available to the system
func (c *Client) getDownloader() (string, error) {

	// Check for cURL
	hasCurl, err := c.readOutput("command -v curl")
	if err == nil && len(hasCurl) > 0 {
		return "curl -sL", nil
	}

	// Check for wget
	hasWget, err := c.readOutput("command -v wget")
	if err == nil && len(hasWget) > 0 {
		return "wget -qO-", nil
	}

	// Return error
	return "", errors.New("Either cURL or wget is required to begin installation.")

}

// pipeToStdOut pipes cmdOut to stdout
func pipeToStdOut(cmdOut io.Reader) {
	_, err := io.Copy(os.Stdout, cmdOut)
	if err != nil {
		log.Printf("Error piping stdout: %v", err)
	}
}

// pipeToStdErr pipes cmdErr to stderr
func pipeToStdErr(cmdErr io.Reader) {
	_, err := io.Copy(os.Stderr, cmdErr)
	if err != nil {
		log.Printf("Error piping stderr: %v", err)
	}
}

// pipeOutput pipes cmdOut and cmdErr to stdout and stderr
func pipeOutput(cmdOut, cmdErr io.Reader) {
	go pipeToStdOut(cmdOut)
	go pipeToStdErr(cmdErr)
}

// Run a command and print its output
func (c *Client) printOutput(cmdText string) error {

	// Initialize command
	cmd, err := c.newCommand(cmdText)
	if err != nil {
		return err
	}
	defer cmd.Close()

	cmdOut, cmdErr, err := cmd.OutputPipes()
	if err != nil {
		return err
	}

	// Begin piping before the command is started
	pipeOutput(cmdOut, cmdErr)

	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Wait for the command to exit
	return cmd.Wait()

}

// Run a command and return its output
func (c *Client) readOutput(cmdText string) ([]byte, error) {

	// Initialize command
	cmd, err := c.newCommand(cmdText)
	if err != nil {
		return []byte{}, err
	}
	defer func() {
		_ = cmd.Close()
	}()

	// Run command and return output
	return cmd.Output()

}
