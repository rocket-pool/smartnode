package service

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// Stops the Smart Node stack containers, prunes Docker, and restarts the Smart Node stack.
func resetDocker(c *cli.Context) error {
	fmt.Println("Once cleanup is complete, the Smart Node will restart automatically.")
	fmt.Println()

	// Stop...
	// NOTE: pauseService prompts for confirmation, so we don't need to do it here
	confirmed, err := stopService(c)
	if err != nil {
		return err
	}

	if !confirmed {
		// if the user cancelled the pause, then we cancel the rest of the operation here:
		return nil
	}

	// Prune images...
	err = pruneDocker(c)
	if err != nil {
		return fmt.Errorf("error pruning Docker: %s", err)
	}

	// Restart...
	// NOTE: startService does some other sanity checks and messages that we leverage here:
	fmt.Println("Restarting the Smart Node...")
	err = startService(c, true)
	if err != nil {
		return fmt.Errorf("error starting the Smart Node: %s", err)
	}
	return nil
}
