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

// NetworksDefaultYAML is the packaged official networks file, copied to
// ~/.rocketpool/networks-default.yml on every `rocketpool service install`.
func NetworksDefaultYAML() []byte {
	bytes, err := installFS.ReadFile("install/networks-default.yml")
	if err != nil {
		panic("embedded install/networks-default.yml is missing: " + err.Error())
	}
	return bytes
}

func ImagesMainnetEnv() []byte {
	bytes, err := installFS.ReadFile("install/mainnet.env")
	if err != nil {
		panic("embedded install/mainnet.env is missing: " + err.Error())
	}
	return bytes
}

func ImagesTestnetEnv() []byte {
	bytes, ok := EmbeddedNetworkEnv("testnet.env")
	if !ok {
		panic("embedded install/testnet.env is missing")
	}
	return bytes
}

func EmbeddedNetworkEnv(name string) ([]byte, bool) {
	bytes, err := installFS.ReadFile("install/" + name)
	if err != nil {
		return nil, false
	}
	return bytes, true
}

func InstallUpdateTrackerScript() ScriptWithContext {
	return ScriptWithContext{Script: installUpdateTrackerScript, Context: rpUpdateTrackerFS}
}
