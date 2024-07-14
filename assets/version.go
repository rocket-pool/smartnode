package assets

import (
	_ "embed"
	"encoding/json"
)

// Using go-embed to import version means a build pipeline can use the .txt file to tag commits/containers
// Id est, one can simply add $(jq -r .Version assets/version.json) to any command to reference the current version.
// We use json because vim (and other editors) likes to add newlines to the end of files.
//go:embed version.json
var versionJSON []byte

type version struct {
	Version string
}

// singleton to hold version after first time parsing
var v *version

func RocketPoolVersion() string {
	if v != nil {
		return v.Version
	}

	v = new(version)
	if err := json.Unmarshal(versionJSON, v); err != nil {
		panic(err)
	}
	if v.Version == "" {
		panic("version must be defined")
	}

	return v.Version
}
