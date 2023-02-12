package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/urfave/cli"
	"golang.org/x/crypto/openpgp"

	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Settings
const (
	ExporterContainerSuffix         string = "_exporter"
	ValidatorContainerSuffix        string = "_validator"
	BeaconContainerSuffix           string = "_eth2"
	ExecutionContainerSuffix        string = "_eth1"
	NodeContainerSuffix             string = "_node"
	ApiContainerSuffix              string = "_api"
	WatchtowerContainerSuffix       string = "_watchtower"
	PruneProvisionerContainerSuffix string = "_prune_provisioner"
	EcMigratorContainerSuffix       string = "_ec_migrator"
	clientDataVolumeName            string = "/ethclient"
	dataFolderVolumeName            string = "/.rocketpool/data"

	PruneFreeSpaceRequired uint64 = 50 * 1024 * 1024 * 1024
	dockerImageRegex       string = ".*/(?P<image>.*):.*"
	colorReset             string = "\033[0m"
	colorBold              string = "\033[1m"
	colorRed               string = "\033[31m"
	colorYellow            string = "\033[33m"
	colorGreen             string = "\033[32m"
	colorLightBlue         string = "\033[36m"
	clearLine              string = "\033[2K"
)

// Creates CLI argument flags from the parameters of the configuration struct
func createFlagsFromConfigParams(sectionName string, params []*cfgtypes.Parameter, configFlags []cli.Flag, network cfgtypes.Network) []cli.Flag {
	for _, param := range params {
		var paramName string
		if sectionName == "" {
			paramName = param.ID
		} else {
			paramName = fmt.Sprintf("%s-%s", sectionName, param.ID)
		}

		defaultVal, err := param.GetDefault(network)
		if err != nil {
			panic(fmt.Sprintf("Error getting default value for [%s]: %s\n", paramName, err.Error()))
		}

		switch param.Type {
		case cfgtypes.ParameterType_Bool:
			configFlags = append(configFlags, cli.BoolFlag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: bool\n", param.Description),
			})
		case cfgtypes.ParameterType_Int:
			configFlags = append(configFlags, cli.IntFlag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: int\n", param.Description),
				Value: int(defaultVal.(int64)),
			})
		case cfgtypes.ParameterType_Float:
			configFlags = append(configFlags, cli.Float64Flag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: float\n", param.Description),
				Value: defaultVal.(float64),
			})
		case cfgtypes.ParameterType_String:
			configFlags = append(configFlags, cli.StringFlag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: string\n", param.Description),
				Value: defaultVal.(string),
			})
		case cfgtypes.ParameterType_Uint:
			configFlags = append(configFlags, cli.UintFlag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: uint\n", param.Description),
				Value: uint(defaultVal.(uint64)),
			})
		case cfgtypes.ParameterType_Uint16:
			configFlags = append(configFlags, cli.UintFlag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: uint16\n", param.Description),
				Value: uint(defaultVal.(uint16)),
			})
		case cfgtypes.ParameterType_Choice:
			optionStrings := []string{}
			for _, option := range param.Options {
				optionStrings = append(optionStrings, fmt.Sprint(option.Value))
			}
			configFlags = append(configFlags, cli.StringFlag{
				Name:  paramName,
				Usage: fmt.Sprintf("%s\n\tType: choice\n\tOptions: %s\n", param.Description, strings.Join(optionStrings, ", ")),
				Value: fmt.Sprint(defaultVal),
			})
		}
	}

	return configFlags
}

func getHttpClientWithTimeout() *http.Client {
	return &http.Client{
		Timeout: time.Second * 5,
	}
}

func checkSignature(signatureUrl string, pubkeyUrl string, verification_target *os.File) error {
	pubkeyResponse, err := http.Get(pubkeyUrl)
	if err != nil {
		return err
	}
	defer pubkeyResponse.Body.Close()
	if pubkeyResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("public key request failed with code %d", pubkeyResponse.StatusCode)
	}
	keyring, err := openpgp.ReadArmoredKeyRing(pubkeyResponse.Body)
	if err != nil {
		return fmt.Errorf("error while reading public key: %w", err)
	}

	signatureResponse, err := http.Get(signatureUrl)
	if err != nil {
		return err
	}
	defer signatureResponse.Body.Close()
	if signatureResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("signature request failed with code %d", signatureResponse.StatusCode)
	}

	entity, err := openpgp.CheckDetachedSignature(keyring, verification_target, signatureResponse.Body)
	if err != nil {
		return fmt.Errorf("error while verifying signature: %w", err)
	}

	for _, v := range entity.Identities {
		fmt.Printf("Signed by: %s", v.Name)
	}
	return nil
}

// Update the Rocket Pool CLI
func updateCLI(c *cli.Context) error {
	// Check the latest version published to the Github repository
	client := getHttpClientWithTimeout()
	resp, err := client.Get("https://api.github.com/repos/rocket-pool/smartnode-install/releases/latest")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with code %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var apiResponse map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return fmt.Errorf("could not decode Github API response: %w", err)
	}
	latestVersion, err := semver.Make(strings.TrimLeft(apiResponse["name"].(string), "v"))
	if err != nil {
		return fmt.Errorf("could not parse latest Rocket Pool version number from API response '%s': %w", apiResponse["name"].(string), err)
	}

	// Check this version against the currently installed version
	if !c.Bool("force") {
		currentVersion, err := semver.Make(shared.RocketPoolVersion)
		if err != nil {
			return fmt.Errorf("could not parse local Rocket Pool version number '%s': %w", shared.RocketPoolVersion, err)
		}
		switch latestVersion.Compare(currentVersion) {
		case 1:
			fmt.Printf("Newer version avilable online (%s). Downloading...\n", latestVersion.String())
		case 0:
			fmt.Printf("Already on latest version (%s). Aborting update\n", latestVersion.String())
			return nil
		default:
			fmt.Printf("Online version (%s) is lower than running version (%s). Aborting update\n", latestVersion.String(), currentVersion.String())
			return nil
		}
	} else {
		fmt.Printf("Forced update to %s. Downloading...\n", latestVersion.String())
	}

	// Download the new binary to same folder as the running RP binary, as `rocketpool-vX.X.X`
	var ClientURL = fmt.Sprintf("https://github.com/rocket-pool/smartnode-install/releases/download/v%s/rocketpool-cli-%s-%s", latestVersion.String(), runtime.GOOS, runtime.GOARCH)
	resp, err = http.Get(ClientURL)
	if err != nil {
		return fmt.Errorf("error while downloading %s: %w", ClientURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with code %d", resp.StatusCode)
	}

	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error while determining running rocketpool location: %w", err)
	}
	var rpBinDir = filepath.Dir(ex)
	var fileName = filepath.Join(rpBinDir, "rocketpool-v"+latestVersion.String())
	output, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("error while creating %s: %w", fileName, err)
	}
	defer output.Close()

	_, err = io.Copy(output, resp.Body)
	if err != nil {
		return fmt.Errorf("error while downloading %s: %w", ClientURL, err)
	}

	// Verify the signature of the downloaded binary
	if !c.Bool("skip-signature-verification") {
		var pubkeyUrl = fmt.Sprintf("https://github.com/rocket-pool/smartnode-install/releases/download/v%s/smartnode-signing-key-v3.asc", latestVersion.String())
		output.Seek(0, io.SeekStart)
		err = checkSignature(ClientURL+".sig", pubkeyUrl, output)
		if err != nil {
			return fmt.Errorf("error while verifying GPG signature: %w", err)
		}
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to update? Current Rocketpool Client will be replaced.")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Do the switcheroo - move `rocketpool-vX.X.X` to the location of the current Rocketpool Client
	err = os.Remove(ex)
	if err != nil {
		return fmt.Errorf("error while removing old rocketpool binary: %w", err)
	}
	err = os.Rename(fileName, ex)
	if err != nil {
		return fmt.Errorf("error while writing new rocketpool binary: %w", err)
	}

	fmt.Printf("Updated Rocketpool Client to v%s. Please run `rocketpool service install` to finish the installation and update your smartstack.\n", latestVersion.String())
	return nil
}

// Register commands
func RegisterCommands(app *cli.App, name string, aliases []string) {

	configFlags := []cli.Flag{}
	cfgTemplate := config.NewRocketPoolConfig("", false)
	network := cfgTemplate.Smartnode.Network.Value.(cfgtypes.Network)

	// Root params
	configFlags = createFlagsFromConfigParams("", cfgTemplate.GetParameters(), configFlags, network)

	// Subconfigs
	for sectionName, subconfig := range cfgTemplate.GetSubconfigs() {
		configFlags = createFlagsFromConfigParams(sectionName, subconfig.GetParameters(), configFlags, network)
	}

	app.Commands = append(app.Commands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Subcommands: []cli.Command{
			{
				Name:      "cli",
				Aliases:   []string{"c"},
				Usage:     "Update the Rocket Pool CLI",
				UsageText: "rocketpool update cli [options]",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "force, f",
						Usage: "Force update, even if same version or lower",
					},
					cli.BoolFlag{
						Name:  "skip-signature-verification, s",
						Usage: "Skip signature verification",
					},
					cli.BoolFlag{
						Name:  "yes, y",
						Usage: "Automatically confirm update",
					},
				},
				Action: func(c *cli.Context) error {

					// Validate args
					if err := cliutils.ValidateArgCount(c, 0); err != nil {
						return err
					}

					// Run command
					return updateCLI(c)

				},
			},
		},
	})
}
