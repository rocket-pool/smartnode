package migration

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/rocket-pool/smartnode/v2/shared/config/ids"
)

const (
	v1LegacyVersionMax   string = "1.99.99"
	legacyRootConfigName string = "root"
	legacyVersionKey     string = "version"
)

type ConfigUpgrader struct {
	Version     *version.Version
	UpgradeFunc func(serializedConfig map[string]any) (map[string]any, error)
}

func UpdateConfig(serializedConfig map[string]any) (map[string]any, error) {

	// Get the config's version
	configVersion, err := getVersionFromConfig(serializedConfig)
	if err != nil {
		return nil, err
	}

	// Create versions
	v1, err := parseVersion(v1LegacyVersionMax)
	if err != nil {
		return nil, err
	}

	// Create the collection of upgraders
	upgraders := []ConfigUpgrader{
		{
			Version:     v1,
			UpgradeFunc: upgradeFromV1,
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
		return serializedConfig, nil
	}

	// If there are upgrades, start at the first applicable index and apply them all in series
	for i := targetIndex; i < len(upgraders); i++ {
		upgrader := upgraders[i]
		serializedConfig, err = upgrader.UpgradeFunc(serializedConfig)
		if err != nil {
			return nil, fmt.Errorf("error applying upgrade for config version %s: %w", upgrader.Version.String(), err)
		}
	}

	return serializedConfig, nil

}

// Get the Smartnode version that the given config was built with
func getVersionFromConfig(serializedConfig map[string]any) (*version.Version, error) {
	var configVersionString string
	configVersionEntry, exists := serializedConfig[ids.VersionID]
	if !exists {
		// Check to see if this is a legacy config
		rootConfigEntry, exists := serializedConfig[legacyRootConfigName]
		if !exists {
			return nil, fmt.Errorf("expected a top-level setting named '%s' but it didn't exist", ids.VersionID)
		}

		// Ok, we have a legacy config - get its version
		rootConfig, ok := rootConfigEntry.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("detected a Smart Node v1 config; it has an entry named [%s] but it is not a map, it's a %s", legacyRootConfigName, reflect.TypeOf(rootConfigEntry))
		}

		configVersionEntry, exists := rootConfig[legacyVersionKey]
		if !exists {
			return nil, fmt.Errorf("detected a Smart Node v1 config but it is missing an entry named [%s.%s]", legacyRootConfigName, legacyVersionKey)
		}

		configVersionString, ok = configVersionEntry.(string)
		if !ok {
			return nil, fmt.Errorf("detected a Smart Node v1 config; it has an entry named [%s.%s] but it is not a string, it's a %s", legacyRootConfigName, legacyVersionKey, reflect.TypeOf(configVersionEntry))
		}
	} else {
		var ok bool
		configVersionString, ok = configVersionEntry.(string)
		if !ok {
			return nil, fmt.Errorf("config has an entry named [%s] but it is not a string, it's a %s", ids.VersionID, reflect.TypeOf(configVersionEntry))
		}
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
