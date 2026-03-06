package update

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rocket-pool/smartnode/shared"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

const (
	downloadUrlFormatString = "https://github.com/rocket-pool/smartnode/releases/latest/download/rocketpool-cli-%s-%s"
)

func validateOsArch() error {

	switch runtime.GOARCH {
	case "amd64":
	case "arm64":
	default:
		return fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}

	switch runtime.GOOS {
	case "linux":
	case "darwin":
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return nil
}

func errorPartialSuccess(err error) {
	fmt.Fprintln(os.Stderr, "An error occurred after the cli binary was updated, but before the service was.")
	fmt.Fprintln(os.Stderr, "The error was:")
	fmt.Fprintf(os.Stderr, "    %s\n", cliutils.Red(err.Error()))
	fmt.Fprintln(os.Stderr)
	printPartialSuccessNextSteps()
	os.Exit(1)
}

func printPartialSuccessNextSteps() {
	fmt.Println("Please complete the following steps to complete the update:")
	fmt.Println("    Run `rocketpool service stop` to stop the service")
	fmt.Println("    Run `rocketpool service install -d` to upgrade the service")
	fmt.Println("    Run `rocketpool service start` to start the service")
}

func forkCommand(binaryPath string, yes bool, args ...string) *exec.Cmd {
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if yes {
		cmd.Args = append(cmd.Args, "--yes")
	}
	return cmd
}

func Update(c *cli.Context) error {
	// Get the pwd and argv[0]
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	argv0 := os.Args[0]

	oldBinaryPath := filepath.Join(pwd, argv0)

	// Validate the OS and architecture
	err = validateOsArch()
	if err != nil {
		return err
	}

	fmt.Printf("Your detected os/architecture is %s/%s.\n", cliutils.Green(runtime.GOOS), cliutils.Green(runtime.GOARCH))
	fmt.Println()

	if !c.Bool("yes") {
		ok := prompt.Confirm("The cli at %s will be replaced. Continue?", cliutils.Yellow(oldBinaryPath))
		if !ok {
			return nil
		}
	}
	fmt.Printf("Replacing the cli at %s with the latest version...\n", oldBinaryPath)

	// Create a temporary directory to download the new binary to
	tempDir, err := os.MkdirTemp("", "rocketpool-cli-update-")
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a file that is executable and has the correct permissions
	tempFile, err := os.CreateTemp(tempDir, "rocketpool-cli-update-*.bin")
	if err != nil {
		return fmt.Errorf("error creating temporary file: %w", err)
	}
	tempFile.Chmod(0755)

	// Download the new binary
	downloadUrl := fmt.Sprintf(downloadUrlFormatString, runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Downloading the new binary from %s\n", cliutils.Green(downloadUrl))
	fmt.Println()
	response, err := http.Get(downloadUrl)
	if err != nil {
		return fmt.Errorf("error downloading new binary: %w", err)
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("error downloading new binary: %s", response.Status)
	}
	defer response.Body.Close()
	_, err = io.Copy(tempFile, response.Body)
	if err != nil {
		return fmt.Errorf("error copying new binary to temporary file: %w", err)
	}
	tempFile.Close()

	// Fork off a process to get the new binary's version and compare
	// it with the current version
	cmd := exec.Command(tempFile.Name(), "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error getting new binary's version: %w", err)
	}
	newVersion := strings.TrimSpace(string(output))
	newVersion = strings.TrimPrefix(newVersion, "rocketpool version ")

	if strings.EqualFold(shared.RocketPoolVersion(), newVersion) && !c.Bool("force") {
		fmt.Println(cliutils.Green(fmt.Sprintf("You are already on the latest version of smartnode: %s.", newVersion)))
		return nil
	}

	fmt.Printf("Updating from %s to %s\n", cliutils.Yellow(shared.RocketPoolVersion()), cliutils.Green(newVersion))

	// Rename the temporary file to the actual binary
	err = os.Rename(tempFile.Name(), oldBinaryPath)
	if err != nil {
		return fmt.Errorf("error replacing binary: %w", err)
	}

	fmt.Println()
	fmt.Println(cliutils.Green("The cli has been updated."))
	fmt.Println()

	if !c.Bool("yes") {
		if !prompt.Confirm("Would you like to automatically stop, upgrade, and restart the service to complete the update?") {
			printPartialSuccessNextSteps()
			return nil
		}
	}

	fmt.Println("=========================================")
	fmt.Println("========= Stopping service... ===========")
	fmt.Println("=========================================")
	stopCmd := []string{"service", "stop"}
	cmd = forkCommand(oldBinaryPath, c.Bool("yes"), stopCmd...)
	err = cmd.Run()
	if err != nil {
		errorPartialSuccess(err)
		return nil
	}

	fmt.Println("=========================================")
	fmt.Println("========= Upgrading service... ==========")
	fmt.Println("=========================================")
	cmd = forkCommand(oldBinaryPath, c.Bool("yes"), "service", "install", "-d")
	err = cmd.Run()
	if err != nil {
		errorPartialSuccess(err)
		return nil
	}

	fmt.Println("=========================================")
	fmt.Println("========= Starting service... ===========")
	fmt.Println("=========================================")
	cmd = forkCommand(oldBinaryPath, c.Bool("yes"), "service", "start")
	err = cmd.Run()
	if err != nil {
		errorPartialSuccess(err)
		return nil
	}

	fmt.Println(cliutils.Green(fmt.Sprintf("The upgrade to Smart Node %s has been completed.", newVersion)))
	fmt.Println()
	fmt.Println(cliutils.Yellow("Please monitor your validators for a few minutes for issues."))
	fmt.Println()

	return nil
}
