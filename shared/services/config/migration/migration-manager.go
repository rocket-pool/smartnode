package migration

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
)

type ConfigUpgrader struct {
	Version     *version.Version
	UpgradeFunc func(serializedConfig map[string]map[string]string) error
}

func UpdateConfig(serializedConfig map[string]map[string]string) error {

	// Get the config's version
	configVersion, err := getVersionFromConfig(serializedConfig)
	if err != nil {
		return err
	}

	// Create versions
	v131, err := parseVersion("1.3.1")
	if err != nil {
		return err
	}

	// Create the collection of upgraders
	upgraders := []ConfigUpgrader{
		{
			Version:     v131,
			UpgradeFunc: upgradeFromV131,
		},
	}

	// Find the index of the provided config's version
	targetIndex := -1
	for i, upgrader := range upgraders {
		if configVersion.LessThanOrEqual(upgrader.Version) {
			targetIndex = i
		}
	}

	// If there are no upgrades to apply, return
	if targetIndex == -1 {
		return nil
	}

	// If there are upgrades, start at the first applicable index and apply them all in series
	for i := targetIndex; i < len(upgraders); i++ {
		upgrader := upgraders[i]
		err = upgrader.UpgradeFunc(serializedConfig)
		if err != nil {
			return fmt.Errorf("error applying upgrade for config version %s: %w", upgrader.Version.String(), err)
		}
	}

	return nil

}

// Get the Smartnode version that the given config was built with
func getVersionFromConfig(serializedConfig map[string]map[string]string) (*version.Version, error) {
	rootConfig, exists := serializedConfig["root"]
	if !exists {
		return nil, fmt.Errorf("expected a section called `root` but it didn't exist")
	}

	configVersionString, exists := rootConfig["version"]
	if !exists {
		return nil, fmt.Errorf("expected a `root` setting named `version` but it didn't exist")
	}

	configVersion, err := version.NewVersion(strings.TrimPrefix(configVersionString, "v"))
	if err != nil {
		return nil, fmt.Errorf("error parsing version [%s] from config file: %w", configVersionString, err)
	}

	return configVersion, nil
}

// Parses a version string into a semantic version
func parseVersion(versionString string) (*version.Version, error) {
	parsedVersion, err := version.NewSemver(versionString)
	if err != nil {
		return nil, fmt.Errorf("error parsing version %s: %w", versionString, err)
	}
	return parsedVersion, nil
}
