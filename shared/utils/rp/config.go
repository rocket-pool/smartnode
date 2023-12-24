package rp

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/alessio/shellescape"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"gopkg.in/yaml.v2"
)

const (
	upgradeFlagFile string = ".firstrun"
)

// Loads a config without updating it if it exists
func LoadConfigFromFile(path string) (*config.RocketPoolConfig, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, nil
	}

	cfg, err := config.LoadFromFile(path)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Saves a config and removes the upgrade flag file
func SaveConfig(cfg *config.RocketPoolConfig, directory, filename string) error {
	path := filepath.Join(directory, filename)

	settings := cfg.Serialize()
	configBytes, err := yaml.Marshal(settings)
	if err != nil {
		return fmt.Errorf("could not serialize settings file: %w", err)
	}

	// Make a tmp file
	// The empty string directs CreateTemp to use the OS's $TMPDIR (or GetTempPath) on windows
	// The * in the second string is replaced with random characters by CreateTemp
	f, err := os.CreateTemp(directory, ".tmp-"+filename+"-*")
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("could not create file to save config to disk... do you need to clean your tmpdir (%s)?: %w", os.TempDir(), err)
		}

		return fmt.Errorf("could not create file to save config to disk: %w", err)
	}
	// Clean up the temporary files
	// This prevents us from filling up `directory` with partially written files on failure
	// If the file is successfully written, it fails with an error since it will be renamed
	// before it is deleted, which we explicitly ignore / don't care about.
	defer func() {
		// Clean up tmp files, if any found
		oldFiles, err := filepath.Glob(filepath.Join(directory, ".tmp-"+filename+"-*"))
		if err != nil {
			// Only possible error is ErrBadPattern, which we should catch
			// during development, since the pattern is a comptime constant.
			panic(err.Error())
		}

		for _, match := range oldFiles {
			os.RemoveAll(match)
		}
	}()

	// Save the serialized settings to the temporary file
	if _, err := f.Write(configBytes); err != nil {
		return fmt.Errorf("could not write Rocket Pool config to %s: %w", shellescape.Quote(path), err)
	}

	// Close the file for writing
	if err := f.Close(); err != nil {
		return fmt.Errorf("error saving Rocket Pool config to %s: %w", shellescape.Quote(path), err)
	}

	// Rename the temp file to overwrite the actual file.
	// On Unix systems this operation is atomic and won't fail if the disk is now full
	if err := os.Rename(f.Name(), path); err != nil {
		return fmt.Errorf("error replacing old Rocket Pool config with %s: %w", f.Name(), err)
	}

	// Just in case the rename didn't overwrite (and preserve the perms of) the original file, set them now.
	if err := os.Chmod(path, 0664); err != nil {
		return fmt.Errorf("error updating permissions of %s: %w", path, err)
	}

	return nil

}

// Checks if this is the first run of the configurator after an install
func IsFirstRun(configDir string) bool {
	upgradeFilePath := filepath.Join(configDir, upgradeFlagFile)

	// Load the config normally if the upgrade flag file isn't there
	_, err := os.Stat(upgradeFilePath)
	return !os.IsNotExist(err)
}

// Remove the upgrade flag file
func RemoveUpgradeFlagFile(configDir string) error {

	// Check for the upgrade flag file
	upgradeFilePath := filepath.Join(configDir, upgradeFlagFile)
	_, err := os.Stat(upgradeFilePath)
	if os.IsNotExist(err) {
		return nil
	}

	// Delete the upgrade flag file
	err = os.Remove(upgradeFilePath)
	if err != nil {
		return fmt.Errorf("error removing upgrade flag file: %w", err)
	}

	return nil

}
