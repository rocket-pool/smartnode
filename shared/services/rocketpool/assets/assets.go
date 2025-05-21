package assets

import (
	"embed"
	"io/fs"
)

// the all: prefix is used because there are hidden files in the install directory
//
//go:embed all:install
var installFS embed.FS

//go:embed rp-update-tracker
var rpUpdateTrackerFS embed.FS

//go:embed scripts/install.sh
var installScript []byte

//go:embed scripts/install-update-tracker.sh
var installUpdateTrackerScript []byte

type ScriptWithContext struct {
	Script  []byte
	Context fs.FS
}

func InstallScript() ScriptWithContext {
	return ScriptWithContext{Script: installScript, Context: installFS}
}

func InstallUpdateTrackerScript() ScriptWithContext {
	return ScriptWithContext{Script: installUpdateTrackerScript, Context: rpUpdateTrackerFS}
}
