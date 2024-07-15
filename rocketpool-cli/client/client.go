package client

import (
	"fmt"
	"log/slog"
	"net/http/httptrace"

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

// Create new Rocket Pool client from CLI context
func NewClientFromCtx(c *cli.Context) (*Client, error) {
	snCtx := context.GetSmartNodeContext(c)
	logger := log.NewTerminalLogger(snCtx.DebugEnabled, terminalLogColor)

	// Create the tracer if required
	var tracer *httptrace.ClientTrace
	if snCtx.HttpTraceFile != nil {
		var err error
		tracer, err = createTracer(snCtx.HttpTraceFile, logger.Logger)
		if err != nil {
			logger.Error("Error creating HTTP trace", log.Err(err))
		}
	}

	// Make the client
	rpClient := &Client{
		Context: snCtx,
		Logger:  logger.Logger,
	}

	// Get the API URL
	url := snCtx.ApiUrl
	if url == nil {
		// Load the config to get the API port
		cfg, _, err := rpClient.LoadConfig()
		if err != nil {
			return nil, fmt.Errorf("error loading config: %w", err)
		}

		url, err = url.Parse(fmt.Sprintf("http://localhost:%d/%s", cfg.ApiPort.Value, config.SmartNodeApiClientRoute))
		if err != nil {
			return nil, fmt.Errorf("error parsing API URL: %w", err)
		}
	}
	rpClient.Api = client.NewApiClient(url, logger.Logger, tracer)
	return rpClient, nil
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
