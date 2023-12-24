package rocketpool

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
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
	"github.com/rocket-pool/smartnode/addons/rescue_node"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
)

// Config
const (
	InstallerURL     string = "https://github.com/rocket-pool/smartnode-install/releases/download/%s/install.sh"
	UpdateTrackerURL string = "https://github.com/rocket-pool/smartnode-install/releases/download/%s/install-update-tracker.sh"

	SettingsFile             string = "user-settings.yml"
	BackupSettingsFile       string = "user-settings-backup.yml"
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

// When printing sync percents, we should avoid printing 100%.
// This function is only called if we're still syncing,
// and the `%0.2f` token will round up if we're above 99.99%.
func SyncRatioToPercent(in float64) float64 {
	return math.Min(99.99, in*100)
}

// Get the external IP address. Try finding an IPv4 address first to:
// * Improve peer discovery and node performance
// * Avoid unnecessary container restarts caused by switching between IPv4 and IPv6
func getExternalIP() (net.IP, error) {
	// Try IPv4 first
	ip4Consensus := externalip.DefaultConsensus(nil, nil)
	err := ip4Consensus.UseIPProtocol(4)
	if err != nil {
		// This will only happen if we pass something other than 0/4/6 to UseIPProtocol
		panic(err)
	}
	if ip, err := ip4Consensus.ExternalIP(); err == nil {
		return ip, nil
	}

	// Try IPv6 as fallback
	ip6Consensus := externalip.DefaultConsensus(nil, nil)
	err = ip6Consensus.UseIPProtocol(6)
	if err != nil {
		// This will only happen if we pass something other than 0/4/6 to UseIPProtocol
		panic(err)
	}
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

func getClientStatusString(clientStatus api.ClientStatus) string {
	if clientStatus.IsSynced {
		return "synced and ready"
	}

	if clientStatus.IsWorking {
		return fmt.Sprintf("syncing (%.2f%%)", SyncRatioToPercent(clientStatus.SyncProgress))
	}

	return fmt.Sprintf("unavailable (%s)", clientStatus.Error)
}

// Check the status of the Execution and Consensus client(s) and provision the API with them
func checkClientStatus(rp *Client) (bool, error) {

	// Check if the primary clients are up, synced, and able to respond to requests - if not, forces the use of the fallbacks for this command
	response, err := rp.GetClientStatus()
	if err != nil {
		return false, err
	}

	ecMgrStatus := response.EcManagerStatus
	bcMgrStatus := response.BcManagerStatus

	// Primary EC and CC are good
	if ecMgrStatus.PrimaryClientStatus.IsSynced && bcMgrStatus.PrimaryClientStatus.IsSynced {
		rp.SetClientStatusFlags(true, false)
		return true, nil
	}

	// Get the status messages
	primaryEcStatus := getClientStatusString(ecMgrStatus.PrimaryClientStatus)
	primaryBcStatus := getClientStatusString(bcMgrStatus.PrimaryClientStatus)
	fallbackEcStatus := getClientStatusString(ecMgrStatus.FallbackClientStatus)
	fallbackBcStatus := getClientStatusString(bcMgrStatus.FallbackClientStatus)

	// Check the fallbacks if enabled
	if ecMgrStatus.FallbackEnabled && bcMgrStatus.FallbackEnabled {

		// Fallback EC and CC are good
		if ecMgrStatus.FallbackClientStatus.IsSynced && bcMgrStatus.FallbackClientStatus.IsSynced {
			fmt.Printf("%sNOTE: primary clients are not ready, using fallback clients...\n\tPrimary EC status: %s\n\tPrimary CC status: %s%s\n\n", colorYellow, primaryEcStatus, primaryBcStatus, colorReset)
			rp.SetClientStatusFlags(true, true)
			return true, nil
		}

		// Both pairs aren't ready
		fmt.Printf("Error: neither primary nor fallback client pairs are ready.\n\tPrimary EC status: %s\n\tFallback EC status: %s\n\tPrimary CC status: %s\n\tFallback CC status: %s\n", primaryEcStatus, fallbackEcStatus, primaryBcStatus, fallbackBcStatus)
		return false, nil
	}

	// Primary isn't ready and fallback isn't enabled
	fmt.Printf("Error: primary client pair isn't ready and fallback clients aren't enabled.\n\tPrimary EC status: %s\n\tPrimary CC status: %s\n", primaryEcStatus, primaryBcStatus)
	return false, nil
}

// Create new Rocket Pool client from CLI context without checking for sync status
// Only use this function from commands that may work if the Daemon service doesn't exist
// Most users should call NewClientFromCtx(c).WithStatus() or NewClientFromCtx(c).WithReady()
func NewClientFromCtx(c *cli.Context) *Client {

	// Return client
	client := &Client{
		configPath:         os.ExpandEnv(c.GlobalString("config-path")),
		daemonPath:         os.ExpandEnv(c.GlobalString("daemon-path")),
		maxFee:             c.GlobalFloat64("maxFee"),
		maxPrioFee:         c.GlobalFloat64("maxPrioFee"),
		gasLimit:           c.GlobalUint64("gasLimit"),
		originalMaxFee:     c.GlobalFloat64("maxFee"),
		originalMaxPrioFee: c.GlobalFloat64("maxPrioFee"),
		originalGasLimit:   c.GlobalUint64("gasLimit"),
		debugPrint:         c.GlobalBool("debug"),
		forceFallbacks:     false,
		ignoreSyncCheck:    false,
	}

	if nonce, ok := c.App.Metadata["nonce"]; ok {
		client.customNonce = nonce.(*big.Int)
	}

	return client
}

// Check the status of a newly created client and return it
// Only use this function from commands that may work without the clients being synced-
// most users should use WithReady instead
func (c *Client) WithStatus() (*Client, bool, error) {
	ready, err := checkClientStatus(c)
	if err != nil {
		c.Close()
		return nil, false, err
	}

	return c, ready, nil
}

// Check the status of a newly created client and ensure the eth clients are synced and ready
func (c *Client) WithReady() (*Client, error) {
	_, ready, err := c.WithStatus()
	if err != nil {
		return nil, err
	}

	if !ready {
		c.Close()
		return nil, fmt.Errorf("clients not ready")
	}

	return c, nil
}

// Close client remote connection
func (c *Client) Close() {
	if c == nil {
		return
	}

	if c.client == nil {
		return
	}

	_ = c.client.Close()
}

// Load the config
// Returns the RocketPoolConfig and whether or not it was newly generated
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

	if cfg != nil {
		// A config was loaded, return it now
		return cfg, false, nil
	}

	// Config wasn't loaded, but there was no error- we should create one.
	return config.NewRocketPoolConfig(c.configPath, c.daemonPath != ""), true, nil
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
	settingsFileDirectoryPath, err := homedir.Expand(c.configPath)
	if err != nil {
		return err
	}
	return rp.SaveConfig(cfg, settingsFileDirectoryPath, SettingsFile)
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
	err = os.WriteFile(prometheusConfigPath, contents, 0664)
	if err != nil {
		return fmt.Errorf("Could not write Prometheus config file to %s: %w", shellescape.Quote(prometheusConfigPath), err)
	}
	err = os.Chmod(prometheusConfigPath, 0664)
	if err != nil {
		return fmt.Errorf("Could not set Prometheus config file %s permissions: %w", shellescape.Quote(prometheusConfigPath), err)
	}

	return nil
}

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
	cmd, err := c.newCommand(fmt.Sprintf("sh -s -- %s", strings.Join(flags, " ")))
	if err != nil {
		return err
	}
	defer func() {
		_ = cmd.Close()
	}()

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
		"-v", shellescape.Quote(version),
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
	cmd, err := c.newCommand(fmt.Sprintf("sh -c \"%s %s\"", path, strings.Join(flags, " ")))
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
		sanitizedStrings[i] = shellescape.Quote(serviceName)
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

// Get the command used to escalate privileges on the system
func (c *Client) getEscalationCommand() (string, error) {
	// Check for sudo first
	sudo := "sudo"
	exists, err := c.checkIfCommandExists(sudo)
	if err != nil {
		return "", fmt.Errorf("error checking if %s exists: %w", sudo, err)
	}
	if exists {
		return sudo, nil
	}

	// Check for doas next
	doas := "doas"
	exists, err = c.checkIfCommandExists(doas)
	if err != nil {
		return "", fmt.Errorf("error checking if %s exists: %w", doas, err)
	}
	if exists {
		return doas, nil
	}

	return "", fmt.Errorf("no privilege escalation command found")
}

func (c *Client) checkIfCommandExists(command string) (bool, error) {
	// Run `type` to check for existence
	cmd := fmt.Sprintf("type %s", command)
	output, err := c.readOutput(cmd)

	if err != nil {
		exitErr, isExitErr := err.(*exec.ExitError)
		if isExitErr && exitErr.ProcessState.ExitCode() == 127 {
			// Command not found
			return false, nil
		} else {
			return false, fmt.Errorf("error checking if %s exists: %w", command, err)
		}
	} else {
		if strings.Contains(string(output), fmt.Sprintf("%s is", command)) {
			return true, nil
		} else {
			return false, fmt.Errorf("unexpected output when checking for %s: %s", command, string(output))
		}
	}
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

	// Include all of the relevant docker compose definition files
	composeFileFlags := []string{}
	for _, container := range deployedContainers {
		composeFileFlags = append(composeFileFlags, fmt.Sprintf("-f %s", shellescape.Quote(container)))
	}
	for _, container := range composeFiles {
		composeFileFlags = append(composeFileFlags, fmt.Sprintf("-f %s", shellescape.Quote(container)))
	}

	// Return command
	return fmt.Sprintf("docker compose --project-directory %s %s %s", shellescape.Quote(expandedConfigPath), strings.Join(composeFileFlags, " "), args), nil

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
	//
	// TODO: Instead of substituting the templates with env vars, switch to the text/template package.
	// However, rescue node plugin will also need to be updated, as it relies on this behavior.
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
	err = os.WriteFile(apiComposePath, contents, 0664)
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
	err = os.WriteFile(nodeComposePath, contents, 0664)
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
	err = os.WriteFile(watchtowerComposePath, contents, 0664)
	if err != nil {
		return []string{}, fmt.Errorf("could not write watchtower container file to %s: %w", watchtowerComposePath, err)
	}
	deployedContainers = append(deployedContainers, watchtowerComposePath)
	deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, config.WatchtowerContainerName+composeFileSuffix))

	// Validator
	// Check if Rescue Node is in-use
	cc, _ := cfg.GetSelectedConsensusClient()
	cleanup, err := cfg.RescueNode.(*rescue_node.RescueNode).ApplyValidatorOverrides(cc)
	if err != nil {
		return []string{}, fmt.Errorf("error using Rescue Node: %w", err)
	}
	contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, config.ValidatorContainerName+templateSuffix))
	if err != nil {
		return []string{}, fmt.Errorf("error reading and substituting validator container template: %w", err)
	}
	validatorComposePath := filepath.Join(runtimeFolder, config.ValidatorContainerName+composeFileSuffix)
	err = os.WriteFile(validatorComposePath, contents, 0664)
	if err != nil {
		return []string{}, fmt.Errorf("could not write validator container file to %s: %w", validatorComposePath, err)
	}
	deployedContainers = append(deployedContainers, validatorComposePath)
	deployedContainers = append(deployedContainers, filepath.Join(overrideFolder, config.ValidatorContainerName+composeFileSuffix))
	// Unset custom env vars from Rescue Node
	cleanup()

	// Check the EC mode to see if it needs to be deployed
	if cfg.ExecutionClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
		contents, err = envsubst.ReadFile(filepath.Join(templatesFolder, config.Eth1ContainerName+templateSuffix))
		if err != nil {
			return []string{}, fmt.Errorf("error reading and substituting execution client container template: %w", err)
		}
		eth1ComposePath := filepath.Join(runtimeFolder, config.Eth1ContainerName+composeFileSuffix)
		err = os.WriteFile(eth1ComposePath, contents, 0664)
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
		err = os.WriteFile(eth2ComposePath, contents, 0664)
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
		err = os.WriteFile(grafanaComposePath, contents, 0664)
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
		err = os.WriteFile(exporterComposePath, contents, 0664)
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
		err = os.WriteFile(prometheusComposePath, contents, 0664)
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
		err = os.WriteFile(mevBoostComposePath, contents, 0664)
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
		err = os.WriteFile(composePath, contents, 0664)
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

// Run a command and print its output
func (c *Client) printOutput(cmdText string) error {

	// Initialize command
	cmd, err := c.newCommand(cmdText)
	if err != nil {
		return err
	}
	defer cmd.Close()

	cmd.SetStdout(os.Stdout)
	cmd.SetStderr(os.Stderr)

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
