package client

import (
	"os"

	"github.com/fatih/color"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/rocketpool"
	"github.com/urfave/cli"
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

// Rocket Pool client
type Client struct {
	Api *rocketpool.ApiRequester

	configPath string
	daemonPath string
	debugPrint bool
}

// Create new Rocket Pool client from CLI context without checking for sync status
// Only use this function from commands that may work if the Daemon service doesn't exist
// Most users should call NewClientFromCtx(c).WithStatus() or NewClientFromCtx(c).WithReady()
func NewClientFromCtx(c *cli.Context) *Client {
	socketPath := os.ExpandEnv(c.GlobalString("api-socket-path"))
	client := &Client{
		configPath: os.ExpandEnv(c.GlobalString("config-path")),
		daemonPath: os.ExpandEnv(c.GlobalString("daemon-path")),
		debugPrint: c.GlobalBool("debug"),

		Api: rocketpool.NewApiRequester(socketPath),
	}
	return client
}
