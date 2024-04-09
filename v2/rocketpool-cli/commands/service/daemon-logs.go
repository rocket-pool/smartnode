package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/urfave/cli/v2"
)

// View the daemon logs
func daemonLogs(c *cli.Context, serviceNames ...string) error {
	lines := c.String(tailFlag.Name)
	lineArg := "--lines="
	if lines == "all" {
		lineArg += "+0"
	} else {
		lineArg += lines
	}

	// Get client
	rp := client.NewClientFromCtx(c)
	cfg, _, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading Smart Node configuration: %w", err)
	}

	// TODO: Get log paths from service names
	logPaths := []string{}
	for _, service := range serviceNames {
		switch service {
		// Vanilla
		case "api", "a":
			logPaths = append(logPaths, cfg.GetApiLogFilePath())
		case "tasks", "t":
			logPaths = append(logPaths, cfg.GetTasksLogFilePath())
		case "watchtower", "w":
			logPaths = append(logPaths, cfg.GetWatchtowerLogFilePath())

		// Modules
		default:
			return fmt.Errorf("unknown service name: %s", service)
		}
	}

	// Print service logs
	return rp.PrintDaemonLogs(getComposeFiles(c), lineArg, logPaths...)
}
