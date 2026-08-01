package wallet

import (
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/term"

	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/color"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// How often the CLI asks the daemon how far along a running recovery is
const recoveryProgressInterval = 2 * time.Second

// Returns true if a recovery is running, in which case the caller should stop
func checkForRunningKeyRecovery(rp *rocketpool.Client, cfg *config.RocketPoolConfig) (bool, error) {
	response, err := rp.GetKeyRecoveryStatus()
	if err != nil {
		return false, fmt.Errorf("error checking for a running key recovery: %w", err)
	}
	if !response.Recovery.Running {
		return false, nil
	}
	recovery := response.Recovery

	color.YellowPrintln("A validator key recovery is already running on the node daemon.")
	fmt.Println()
	fmt.Printf("  Operation:   %s\n", recovery.Operation)
	fmt.Printf("  Running for: %s\n", time.Duration(recovery.ElapsedSeconds*float64(time.Second)).Round(time.Second))
	if recovery.TotalKnown {
		fmt.Printf("  Progress:    %d of %d validator keys recovered\n", recovery.KeysFound, recovery.KeysTotal)
	} else {
		fmt.Println("  Progress:    still looking up this node's validator keys")
	}
	fmt.Println()
	fmt.Println("This usually means a recovery was started earlier and the CLI was closed before it")
	fmt.Println("finished. The daemon keeps working in the background, so your keys are still being")
	fmt.Println("recovered. Closing the CLI does not stop it. Nodes with many validators can take")
	fmt.Println("several minutes.")
	fmt.Println()
	fmt.Println("Wait for it to finish, then run this command again to see the result.")
	fmt.Println()
	printRecoveryKillInstructions(cfg)
	return true, nil
}

// Polls the daemon and prints progress while a recovery request is in flight.
func startRecoveryProgressReporter(rp *rocketpool.Client) func() {
	// Overwriting a single line only makes sense on a terminal; anywhere else (a
	// pipe, a log file) each update has to go on its own line
	interactive := term.IsTerminal(int(os.Stdout.Fd()))
	return startProgressReporter(os.Stdout, interactive, recoveryProgressInterval, func() (api.KeyRecoveryStatus, error) {
		response, err := rp.GetKeyRecoveryStatus()
		return response.Recovery, err
	})
}

// The polling loop behind startRecoveryProgressReporter
func startProgressReporter(out io.Writer, interactive bool, interval time.Duration, poll func() (api.KeyRecoveryStatus, error)) func() {
	done := make(chan struct{})
	finished := make(chan struct{})

	go func() {
		defer close(finished)

		wrote := false
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				// Close off the line the updates have been overwriting
				if wrote && interactive {
					_, _ = fmt.Fprintln(out)
				}
				return
			case <-ticker.C:
				recovery, err := poll()
				// A failed probe says nothing about the recovery itself. that is
				// reported by the call this runs alongside, so just try again
				if err != nil || !recovery.Running {
					continue
				}
				line := formatRecoveryProgress(recovery)
				if interactive {
					_, _ = fmt.Fprintf(out, "\r\033[K%s", line)
				} else {
					_, _ = fmt.Fprintln(out, line)
				}
				wrote = true
			}
		}
	}()

	return func() {
		close(done)
		<-finished
	}
}

// Renders a one-line summary of a running recovery
func formatRecoveryProgress(recovery api.KeyRecoveryStatus) string {
	elapsed := time.Duration(recovery.ElapsedSeconds * float64(time.Second)).Round(time.Second)
	switch {
	case !recovery.TotalKnown:
		return fmt.Sprintf("  Looking up this node's validator keys... (%s elapsed)", elapsed)
	case recovery.KeysTotal == 0:
		return fmt.Sprintf("  No validator keys to recover (%s elapsed)", elapsed)
	default:
		return fmt.Sprintf("  Recovered %d of %d validator keys... (%s elapsed)", recovery.KeysFound, recovery.KeysTotal, elapsed)
	}
}

// Explains how to stop a recovery that is genuinely stuck
func printRecoveryKillInstructions(cfg *config.RocketPoolConfig) {
	fmt.Println("If you are sure it is stuck, restart the node daemon to stop it:")
	fmt.Println()
	if cfg.IsNativeMode {
		fmt.Println("    You are using Native Mode, so restart the service that runs the node daemon,")
		fmt.Println("    for example:  sudo systemctl restart <your node daemon service>")
	} else {
		projectName := cfg.Smartnode.ProjectName.Value.(string)
		fmt.Printf("    docker restart %s_%s\n", projectName, config.NodeContainerName)
	}
	fmt.Println()
	color.YellowPrintln("WARNING: this aborts the recovery partway through, so some validator keys may be")
	color.YellowPrintln("written and others not. Run `rocketpool wallet rebuild` afterwards to finish the job.")
	fmt.Println()
}
