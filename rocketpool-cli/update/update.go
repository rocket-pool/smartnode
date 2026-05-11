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

	"github.com/rocket-pool/smartnode/rocketpool-cli/update/assets"
	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

const (
	downloadUrlFormatString = "https://github.com/rocket-pool/smartnode/releases/latest/download/rocketpool-cli-%s-%s"
	downloadDirName         = ".smartnode_downloads"
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
	color.RedPrintf("    %s\n", err.Error())
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

func Update(yes bool, skipSignatureVerification bool, force bool) error {
	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	// Get the config
	cfg, _, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading configuration: %w", err)
	}

	if cfg.IsNativeMode {
		fmt.Println("You are using Native Mode.")
		fmt.Println("The Smart Node cannot update the cli for you (yet), you'll have to do it manually.")
		fmt.Println("Please follow the instructions at https://docs.rocketpool.net/node-staking/updates#updating-the-smartnode-stack")
		fmt.Println()
		return fmt.Errorf("native mode not supported yet")
	}

	oldBinaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting path of current executable: %w", err)
	}

	// Validate the OS and architecture
	err = validateOsArch()
	if err != nil {
		return err
	}

	fmt.Printf("Your detected os/architecture is %s/%s.\n", color.Green(runtime.GOOS), color.Green(runtime.GOARCH))
	fmt.Println()

	if !yes {
		ok := prompt.Confirm("The cli at %s will be replaced. Continue?", color.Yellow(oldBinaryPath))
		if !ok {
			return nil
		}
	}
	fmt.Printf("Replacing the cli at %s with the latest version...\n", oldBinaryPath)

	downloadDir := filepath.Join(filepath.Dir(oldBinaryPath), downloadDirName)
	err = os.MkdirAll(downloadDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating download directory: %w", err)
	}
	defer func() {
		// Ignore errors, future downloads will hopefully succeed at deleting the directory
		_ = os.RemoveAll(downloadDir)
	}()

	// Create a file that is executable and has the correct permissions
	tempFile, err := os.CreateTemp(downloadDir, "rocketpool-cli-update-*.bin")
	if err != nil {
		return fmt.Errorf("error creating temporary file: %w", err)
	}
	err = tempFile.Chmod(0755)
	if err != nil {
		return fmt.Errorf("error setting temporary file permissions: %w", err)
	}
	defer tempFile.Close()

	// Download the new binary
	downloadUrl := fmt.Sprintf(downloadUrlFormatString, runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Downloading the new binary from %s\n", color.Green(downloadUrl))
	fmt.Println()
	response, err := http.Get(downloadUrl)
	if err != nil {
		return fmt.Errorf("error downloading new binary: %w", err)
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("error downloading new binary: %s", response.Status)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if !skipSignatureVerification {
		// Download the signature file
		fmt.Println("Verifying the binary signature")
		signatureUrl := fmt.Sprintf("%s.sig", downloadUrl)
		fmt.Printf("Downloading the signature from %s\n", signatureUrl)
		signatureResponse, err := http.Get(signatureUrl)
		if err != nil {
			return fmt.Errorf("error downloading signature: %w", err)
		}
		defer func() {
			_ = signatureResponse.Body.Close()
		}()
		if signatureResponse.StatusCode != 200 {
			return fmt.Errorf("error downloading signature: %s", signatureResponse.Status)
		}
		teeReader := io.TeeReader(response.Body, tempFile)
		signer, err := assets.VerifySignedBinary(teeReader, signatureResponse.Body)
		if err != nil {
			return fmt.Errorf("error verifying signed binary: %w", err)
		}
		fmt.Printf("Signed by %s\n", color.Green(signer.PrimaryKey.KeyIdString()))
		for _, identity := range signer.Identities {
			fmt.Printf("  %s\n", color.Green(identity.Name))
		}
	} else {
		_, err = io.Copy(tempFile, response.Body)
		if err != nil {
			return fmt.Errorf("error copying new binary to temporary file: %w", err)
		}
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

	if strings.EqualFold(shared.RocketPoolVersion(), newVersion) && !force {
		color.GreenPrintf("You are already on the latest version of smartnode: %s.", newVersion)
		return nil
	}

	fmt.Printf("Updating from %s to %s\n", color.Yellow(shared.RocketPoolVersion()), color.Green(newVersion))

	// Rename the temporary file to the actual binary
	err = os.Rename(tempFile.Name(), oldBinaryPath)
	if err != nil {
		return fmt.Errorf("error replacing binary: %w", err)
	}

	fmt.Println()
	color.GreenPrintln("The cli has been updated.")
	fmt.Println()

	if !yes {
		if !prompt.Confirm("Would you like to automatically stop, upgrade, and restart the service to complete the update?") {
			printPartialSuccessNextSteps()
			return nil
		}
	}

	fmt.Println("=========================================")
	fmt.Println("========= Stopping service... ===========")
	fmt.Println("=========================================")
	stopCmd := []string{"service", "stop"}
	cmd = forkCommand(oldBinaryPath, yes, stopCmd...)
	err = cmd.Run()
	if err != nil {
		errorPartialSuccess(err)
		return nil
	}

	fmt.Println("=========================================")
	fmt.Println("========= Upgrading service... ==========")
	fmt.Println("=========================================")
	cmd = forkCommand(oldBinaryPath, yes, "service", "install", "-d")
	err = cmd.Run()
	if err != nil {
		errorPartialSuccess(err)
		return nil
	}

	fmt.Println("=========================================")
	fmt.Println("========= Starting service... ===========")
	fmt.Println("=========================================")
	cmd = forkCommand(oldBinaryPath, yes, "service", "start")
	err = cmd.Run()
	if err != nil {
		errorPartialSuccess(err)
		return nil
	}

	color.GreenPrintf("The upgrade to Smart Node %s has been completed.", newVersion)
	fmt.Println()
	color.YellowPrintln("Please monitor your validators for a few minutes for issues.")
	fmt.Println()

	return nil
}
