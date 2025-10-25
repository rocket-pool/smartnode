package rocketpool

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/goccy/go-json"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"

	"github.com/alessio/shellescape"
	"github.com/blang/semver/v4"
	"github.com/mitchellh/go-homedir"
	"github.com/rocket-pool/smartnode/addons/graffiti_wall_writer"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool/assets"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool/template"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
)

// Config
const (
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

	nethermindAdminUrl          string = "http://127.0.0.1:7434"
	pruneStarterContainerSuffix string = "_nm_prune_starter"

	DebugColor = color.FgYellow
)

// When printing sync percents, we should avoid printing 100%.
// This function is only called if we're still syncing,
// and the `%0.2f` token will round up if we're above 99.99%.
func SyncRatioToPercent(in float64) float64 {
	return math.Min(99.99, in*100)
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

func (c *Client) ConfigPath() string {
	return c.configPath
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

// Load the Prometheus template, do a template variable substitution, and save it
func (c *Client) UpdatePrometheusConfiguration(config *config.RocketPoolConfig) error {
	prometheusTemplatePath, err := homedir.Expand(fmt.Sprintf("%s/%s", c.configPath, PrometheusConfigTemplate))
	if err != nil {
		return fmt.Errorf("Error expanding Prometheus template path: %w", err)
	}

	prometheusConfigPath, err := homedir.Expand(fmt.Sprintf("%s/%s", c.configPath, PrometheusFile))
	if err != nil {
		return fmt.Errorf("Error expanding Prometheus config file path: %w", err)
	}

	t := template.Template{
		Src: prometheusTemplatePath,
		Dst: prometheusConfigPath,
	}

	return t.Write(config)
}

func (c *Client) runScript(script assets.ScriptWithContext, verbose bool, flags []string) error {
	// Make a tmpdir
	tmpdir, err := os.MkdirTemp("", "rocketpool-")
	if err != nil {
		return fmt.Errorf("error creating tmpdir: %w", err)
	}
	if verbose {
		fmt.Printf("Verbose mode enabled, tmpdir %s will not be removed\n", tmpdir)
	} else {
		defer os.RemoveAll(tmpdir)
	}

	// Create a file in the tmpdir
	scriptPathName := filepath.Join(tmpdir, "script.sh")
	scriptFile, err := os.Create(scriptPathName)
	if err != nil {
		return fmt.Errorf("error creating script file: %w", err)
	}
	if err := scriptFile.Chmod(0700); err != nil {
		return fmt.Errorf("error setting script file permissions: %w", err)
	}
	// write the script to the file
	_, err = scriptFile.Write(script.Script)
	if err != nil {
		return fmt.Errorf("error writing script to file: %w", err)
	}
	scriptFile.Close()

	// Copy the context to the tmpdir
	// If we upgrade to go 1.23+ we can probably use os.CopyFS() instead
	err = fs.WalkDir(script.Context, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dstpath := filepath.Join(tmpdir, path)
		if d.IsDir() {
			// If d is a directory, create it.
			if verbose {
				fmt.Printf("Creating directory: %s\n", path)
			}
			err = os.MkdirAll(dstpath, 0755)
			if err != nil {
				return err
			}
			return nil
		}

		// d is a file, copy it.
		if verbose {
			fmt.Printf("Copying file: %s\n", path)
		}
		scriptFile, err := os.Create(dstpath)
		if err != nil {
			return err
		}
		content, err := fs.ReadFile(script.Context, path)
		if err != nil {
			return err
		}
		_, err = scriptFile.Write(content)
		if err != nil {
			return err
		}
		err = scriptFile.Close()
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error copying script files: %w", err)
	}

	// Initialize command
	cmdString := fmt.Sprintf("%s %s", scriptPathName, strings.Join(flags, " "))
	if verbose {
		fmt.Printf("Running script: %s\n", cmdString)
	}
	cmd, err := c.newCommand(cmdString)
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
		return fmt.Errorf("Could not run script: %s", errMessage)
	}
	return nil

}

// Install the Rocket Pool service
func (c *Client) InstallService(verbose, noDeps bool, path string, dataPath string) error {

	// Get installation script flags
	flags := []string{}
	if path != "" {
		flags = append(flags, fmt.Sprintf("-p %s", shellescape.Quote(path)))
	}
	if noDeps {
		flags = append(flags, "-d")
	}
	if dataPath != "" {
		flags = append(flags, fmt.Sprintf("-u %s", dataPath))
	}

	// Load the installation script
	return c.runScript(assets.InstallScript(), verbose, flags)
}

// Install the update tracker
func (c *Client) InstallUpdateTracker(verbose bool) error {

	return c.runScript(assets.InstallUpdateTrackerScript(), verbose, nil)
}

// Start the Rocket Pool service
func (c *Client) StartService(composeFiles []string) error {

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
	cmd = fmt.Sprintf("rm -rf %s", path)
	// The directory contains root-owned paths, so delete it as root
	_, err = c.readOutputSudo(cmd)
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

// Deletes a docker image
func (c *Client) DeleteDockerImage(id string) (string, error) {

	cmd := fmt.Sprintf("docker image rm %s", shellescape.Quote(id))
	output, err := c.readOutput(cmd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil

}

// Runs docker system prune remove all unused containers, networks, and unused images
func (c *Client) DockerSystemPrune(deleteAllImages bool) error {

	// NOTE: explicitly *NOT* using the --all flag, as it would remove all images,
	//   not just unused ones, and we use this command to preserve the current
	//   smartnode stack images.
	cmd := "docker system prune -f"
	if deleteAllImages {
		cmd += " --all"
	}
	err := c.printOutput(cmd)
	if err != nil {
		return fmt.Errorf("error running docker system prune: %w", err)
	}
	return nil
}

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

func (c *Client) PullComposeImages(composeFiles []string) error {
	cmd, err := c.compose(composeFiles, "pull -q")
	if err != nil {
		return err
	}
	err = c.printOutput(cmd)
	if err != nil {
		return fmt.Errorf("error pulling images: %w", err)
	}
	return nil
}

type DockerImage struct {
	Repository string `json:"Repository"`
	Tag        string `json:"Tag"`
	ID         string `json:"ID"`
}

func (img *DockerImage) TagString() string {
	return fmt.Sprintf("%s:%s", img.Repository, img.Tag)
}

func (img *DockerImage) String() string {
	return fmt.Sprintf("%s:%s (%s)", img.Repository, img.Tag, img.ID)
}

// Returns all Docker images on the system
func (c *Client) GetAllDockerImages() ([]DockerImage, error) {
	cmd := "docker images -a --format json"
	responseBytes, err := c.readOutput(cmd)
	if err != nil {
		return nil, err
	}

	// docker images output puts each image as a json object on a new line (JSONL)
	var images []DockerImage
	lines := strings.SplitSeq(string(responseBytes), "\n")
	for line := range lines {
		if line == "" {
			continue
		}
		var image DockerImage
		if err := json.Unmarshal([]byte(line), &image); err != nil {
			return nil, fmt.Errorf("could not decode docker image: %w", err)
		}
		images = append(images, image)
	}

	return images, nil
}

type DockerContainer struct {
	Names  string `json:"Names"`
	State  string `json:"State"`
	Mounts string `json:"Mounts"`
}

func (c *DockerContainer) HasVolume(volume string) bool {
	mounts := strings.Split(c.Mounts, ",")
	return slices.Contains(mounts, volume)
}

// Returns all Docker containers on the system with the given prefix
func (c *Client) GetContainersByPrefix(prefix string) ([]DockerContainer, error) {
	cmd := fmt.Sprintf("docker container ls -a --no-trunc --format json --filter label=com.docker.compose.project=%s", prefix)
	output, err := c.readOutput(cmd)
	if err != nil {
		return nil, err
	}

	// docker container ls output puts each container as a json object on a new line (JSONL)
	lines := strings.Split(string(output), "\n")
	out := make([]DockerContainer, 0, len(lines))
	for _, l := range lines {
		if l == "" {
			continue
		}
		var container DockerContainer
		if err := json.Unmarshal([]byte(l), &container); err != nil {
			return nil, fmt.Errorf("could not decode docker container: %w", err)
		}
		out = append(out, container)
	}
	return out, nil
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
func (c *Client) RunPruneProvisioner(container, volume string) error {

	// Run the prune provisioner
	cmd := fmt.Sprintf("docker run --rm --name %s -v %s:/ethclient alpine:latest sh -c 'touch /ethclient/prune.lock'", container, volume)
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

// Curls the Nethermind admin URL to trigger pruning
func (c *Client) RunNethermindPruneStarter(executionContainerName string) error {
	retryCount := 5
	retryTime := 3 * time.Second

	for i := 0; i < retryCount; i++ {
		command := fmt.Sprintf(`-m 30 -H "Content-Type: application/json" -X POST --data '{"jsonrpc":"2.0","method":"admin_prune","params":[],"id":%d}' %s`, i+1, nethermindAdminUrl)
		cmdText := fmt.Sprintf(`docker run --quiet --rm  --name curl%s --network container:%s curlimages/curl -Ss %s`, pruneStarterContainerSuffix, executionContainerName, command)

		if i != 0 {
			fmt.Printf("Trying again in %v... (%d/%d)\n", retryTime, i+1, retryCount)
			time.Sleep(retryTime)
		}

		cmd, err := c.newCommand(cmdText)
		if err != nil {
			return fmt.Errorf("error creating command for prune starter: %w", err)
		}

		stdOut, stdErr, err := cmd.OutputPipes()
		if err != nil {
			return fmt.Errorf("error getting output pipes for prune starter: %w", err)
		}

		err = cmd.Start()
		if err != nil {
			return fmt.Errorf("error running prune starter: %w", err)
		}
		defer func() {
			_ = cmd.Wait()
		}()

		// Check for a curl error
		stdErrTextBytes, err := io.ReadAll(stdErr)
		if err != nil {
			return fmt.Errorf("error reading error from prune starter: %w", err)
		}
		stdErrText := string(stdErrTextBytes)
		if stdErrText != "" {
			fmt.Printf("Error while curling the Nethermind admin URL: %s\n", stdErrText)
			continue
		}

		// Grab the response
		stdOutText, err := io.ReadAll(stdOut)
		if err != nil {
			return fmt.Errorf("error reading response from prune starter: %w", err)
		}
		// Parse the response as JSON
		var response map[string]any
		err = json.Unmarshal(stdOutText, &response)
		if err != nil {
			return fmt.Errorf("error parsing response from prune starter: %w", err)
		}

		if errObject, ok := response["error"].(map[string]any); ok {
			fmt.Printf("Error starting prune: code %d, message = %s, data = %s\n", errObject["code"], errObject["message"], errObject["data"])
			continue
		} else {
			fmt.Printf("Success: Pruning is now \"%s\"\n", response["result"])
			fmt.Println("Your main execution client is now pruning. You can follow its progress with `rocketpool service logs eth1`.")
			fmt.Println("NOTE: While pruning, you **cannot** interrupt the client (e.g. by restarting) or you risk corrupting the database!")
			fmt.Println("You must let it run to completion!")
			break
		}

	}
	return nil
}

// Deletes the node wallet and all validator keys, and restarts the Docker containers
func (c *Client) PurgeAllKeys(composeFiles []string) error {

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
	cmd := fmt.Sprintf("rm -f %s", walletPath)
	// The file is owned by root, so delete as root
	_, err = c.readOutputSudo(cmd)
	if err != nil {
		return fmt.Errorf("error deleting wallet: %w", err)
	}

	// Delete the password
	passwordPath, err := homedir.Expand(cfg.Smartnode.GetPasswordPathInCLI())
	if err != nil {
		return fmt.Errorf("error loading password path: %w", err)
	}
	fmt.Println("Deleting password...")
	cmd = fmt.Sprintf("rm -f %s", passwordPath)
	// The file is owned by root, so delete as root
	_, err = c.readOutputSudo(cmd)
	if err != nil {
		return fmt.Errorf("error deleting password: %w", err)
	}

	// Delete the validators dir
	validatorsPath, err := homedir.Expand(cfg.Smartnode.GetValidatorKeychainPathInCLI())
	if err != nil {
		return fmt.Errorf("error loading validators folder path: %w", err)
	}
	fmt.Println("Deleting validator keys...")
	cmd = fmt.Sprintf("rm -rf %s/*", validatorsPath)
	// The validators path can be created by the smartnode daemon (owned by root, 0600)
	// So delete its contents as root, otherwise the * won't expand.
	// NB: we delete the contents of the folder instead of recreating the folder
	// This way, if the drive is full, we don't release the directory inode and fail to recreate it.
	_, err = c.readOutputSudo(cmd)
	if err != nil {
		return fmt.Errorf("error deleting validator keys: %w", err)
	}
	// Also delete hidden files
	cmd = fmt.Sprintf("rm -rf %s/.[a-zA-Z0-9]*", validatorsPath)
	// also as root, so bash can expand the regex
	_, err = c.readOutputSudo(cmd)
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
		return "", fmt.Errorf("Settings file not found. Please run `rocketpool service config` to set up your Smart Node before starting it.")
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
	return fmt.Sprintf("COMPOSE_PROJECT_NAME=%s docker compose --project-directory %s %s %s", cfg.Smartnode.ProjectName.Value.(string), shellescape.Quote(expandedConfigPath), strings.Join(composeFileFlags, " "), args), nil

}

// Deploys all of the appropriate docker compose template files and provisions them based on the provided configuration
func (c *Client) deployTemplates(cfg *config.RocketPoolConfig, rocketpoolDir string) ([]string, error) {

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

	composePaths := template.ComposePaths{
		RuntimePath:  runtimeFolder,
		TemplatePath: templatesFolder,
		OverridePath: overrideFolder,
	}

	// Read and substitute the templates
	deployedContainers := []string{}

	// These containers always run
	toDeploy := []string{
		config.ApiContainerName,
		config.NodeContainerName,
		config.WatchtowerContainerName,
		config.ValidatorContainerName,
	}

	// Check if we are running the Execution Layer locally
	if cfg.ExecutionClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
		toDeploy = append(toDeploy, config.Eth1ContainerName)
	}

	// Check if we are running the Consensus Layer locally
	if cfg.ConsensusClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
		toDeploy = append(toDeploy, config.Eth2ContainerName)
	}

	// Check the metrics containers
	if cfg.EnableMetrics.Value == true {
		toDeploy = append(toDeploy,
			config.GrafanaContainerName,
			config.ExporterContainerName,
			config.PrometheusContainerName,
		)
	}

	// Check if Alerts are enabled
	if cfg.Alertmanager.EnableAlerting.Value == true {
		toDeploy = append(toDeploy,
			config.AlertmanagerContainerName,
		)
	}

	// Check if we are running the Mev-Boost container locally
	if cfg.EnableMevBoost.Value == true && cfg.MevBoost.Mode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
		toDeploy = append(toDeploy, config.MevBoostContainerName)
	}

	for _, containerName := range toDeploy {
		containers, err := composePaths.File(containerName).Write(cfg)
		if err != nil {
			return []string{}, fmt.Errorf("could not create %s container definition: %w", containerName, err)
		}
		deployedContainers = append(deployedContainers, containers...)
	}

	// Create the custom keys dir
	customKeyDir, err := homedir.Expand(filepath.Join(cfg.Smartnode.DataPath.Value.(string), "custom-keys"))
	if err != nil {
		fmt.Printf("%sWARNING: Couldn't expand the custom validator key directory (%s). You will not be able to recover any minipool keys you created outside of the Smart Node until you create the folder manually.%s\n", colorYellow, err.Error(), colorReset)
		return deployedContainers, nil
	}
	err = os.MkdirAll(customKeyDir, 0775)
	if err != nil {
		fmt.Printf("%sWARNING: Couldn't create the custom validator key directory (%s). You will not be able to recover any minipool keys you created outside of the Smart Node until you create the folder [%s] manually.%s\n", colorYellow, err.Error(), customKeyDir, colorReset)
	}

	// Create the rewards file dir
	rewardsFileDir, err := homedir.Expand(cfg.Smartnode.GetRewardsTreeDirectory(false))
	if err != nil {
		fmt.Printf("%sWARNING: Couldn't expand the rewards tree file directory (%s). You will not be able to view or claim your rewards until you create the folder manually.%s\n", colorYellow, err.Error(), colorReset)
		return deployedContainers, nil
	}
	err = os.MkdirAll(rewardsFileDir, 0775)
	if err != nil {
		fmt.Printf("%sWARNING: Couldn't create the rewards tree file directory (%s). You will not be able to view or claim your rewards until you create the folder [%s] manually.%s\n", colorYellow, err.Error(), rewardsFileDir, colorReset)
	}

	return c.composeAddons(cfg, rocketpoolDir, deployedContainers)

}

// Handle composing for addons
func (c *Client) composeAddons(cfg *config.RocketPoolConfig, rocketpoolDir string, deployedContainers []string) ([]string, error) {

	// GWW
	if cfg.GraffitiWallWriter.GetEnabledParameter().Value == true {

		composePaths := template.ComposePaths{
			RuntimePath:  filepath.Join(rocketpoolDir, runtimeDir, "addons", "gww"),
			TemplatePath: filepath.Join(rocketpoolDir, templatesDir, "addons", "gww"),
			OverridePath: filepath.Join(rocketpoolDir, overrideDir, "addons", "gww"),
		}

		// Make the addon folder
		err := os.MkdirAll(composePaths.RuntimePath, 0775)
		if err != nil {
			return []string{}, fmt.Errorf("error creating addon runtime folder (%s): %w", composePaths.RuntimePath, err)
		}

		containers, err := composePaths.File(graffiti_wall_writer.GraffitiWallWriterContainerName).Write(cfg)
		if err != nil {
			return []string{}, fmt.Errorf("could not create gww container definition: %w", err)
		}
		deployedContainers = append(deployedContainers, containers...)
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
	for arg := range strings.FieldsSeq(args) {
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

// Run a command as root and return its output
func (c *Client) readOutputSudo(rootCmdText string) ([]byte, error) {
	var escCmd string
	for _, escalationCommand := range []string{"sudo", "doas"} {
		exists, err := c.checkIfCommandExists(escalationCommand)
		if err != nil {
			return nil, fmt.Errorf("error checking if %s exists: %w", escalationCommand, err)
		}
		if exists {
			escCmd = escalationCommand
			break
		}
	}
	if escCmd == "" {
		return nil, fmt.Errorf("no privilege escalation command found")
	}

	return c.readOutput(fmt.Sprintf("%s bash -c %s", escCmd, shellescape.Quote(rootCmdText)))
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

// Gets the container prefix from the settings
func (c *Client) GetContainerPrefix() (string, error) {
	cfg, isNew, err := c.LoadConfig()
	if err != nil {
		return "", err
	}
	if isNew {
		return "", fmt.Errorf("Settings file not found. Please run `rocketpool service config` to set up your Smart Node.")
	}

	return cfg.Smartnode.ProjectName.Value.(string), nil
}
