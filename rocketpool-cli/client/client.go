package client

import (
	"fmt"
	"log/slog"
	"path/filepath"

	docker "github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/context"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/urfave/cli/v2"
)

// Config
const (
	InstallerName              string = "install.sh"
	UpdateTrackerInstallerName string = "install-update-tracker.sh"
	InstallerURL               string = "https://github.com/rocket-pool/smartnode/releases/download/%s/" + InstallerName
	UpdateTrackerURL           string = "https://github.com/rocket-pool/smartnode/releases/download/%s/" + UpdateTrackerInstallerName

	SettingsFile       string = "user-settings.yml"
	BackupSettingsFile string = "user-settings-backup.yml"

	terminalLogColor color.Attribute = color.FgHiYellow
)

// Rocket Pool client
type Client struct {
	Api      *client.ApiClient
	Context  *context.SmartNodeContext
	Logger   *slog.Logger
	docker   *docker.Client
	cfg      *config.SmartNodeConfig
	isNewCfg bool
}

// Create new Rocket Pool client from CLI context without checking for sync status
// Only use this function from commands that may work if the Daemon service doesn't exist
// Most users should call NewClientFromCtx(c).WithStatus() or NewClientFromCtx(c).WithReady()
func NewClientFromCtx(c *cli.Context) *Client {
	snCtx := context.GetSmartNodeContext(c)
	socketPath := filepath.Join(snCtx.ConfigPath, config.SmartNodeCliSocketFilename)

	// Make the client
	logger := log.NewTerminalLogger(snCtx.DebugEnabled, terminalLogColor)
	client := &Client{
		Api:     client.NewApiClient(config.SmartNodeApiClientRoute, socketPath, logger.Logger),
		Context: snCtx,
		Logger:  logger.Logger,
	}
	return client
}

// Get the Docker client
func (c *Client) GetDocker() (*docker.Client, error) {
	if c.docker == nil {
		var err error
		c.docker, err = docker.NewClientWithOpts(docker.WithAPIVersionNegotiation())
		if err != nil {
			return nil, fmt.Errorf("error creating Docker client: %w", err)
		}
	}

	return c.docker, nil
}
