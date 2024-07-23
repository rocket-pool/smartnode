package settings

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	// Environment variable to set the system path for unit tests
	TestSystemDirEnvVar string = "SMARTNODE_TEST_SYSTEM_DIR"

	// System dir path for Linux
	linuxSystemDir string = "/usr/share/rocketpool"

	// Subfolders under the system dir
	scriptsDir        string = "scripts"
	templatesDir      string = "templates"
	overrideSourceDir string = "override"
	networksDir       string = "networks"
)

// Holds the location of various Smart Node system directories
type SystemSettings struct {
	// The system path for Smart Node scripts used in the Docker containers
	ScriptsDir string

	// The system path for Smart Node templates
	TemplatesDir string

	// The system path for the source files to put in the user's override directory
	OverrideSourceDir string

	// The system path for built-in network settings and resource definitions
	NetworksDir string
}

func NewSystemSettings() *SystemSettings {
	systemDir := os.Getenv(TestSystemDirEnvVar)
	if systemDir == "" {
		switch runtime.GOOS {
		// This is where to add different paths for different OS's like macOS
		default:
			// By default just use the Linux path
			systemDir = linuxSystemDir
		}
	}

	return &SystemSettings{
		ScriptsDir:        filepath.Join(systemDir, scriptsDir),
		TemplatesDir:      filepath.Join(systemDir, templatesDir),
		OverrideSourceDir: filepath.Join(systemDir, overrideSourceDir),
		NetworksDir:       filepath.Join(systemDir, networksDir),
	}
}
