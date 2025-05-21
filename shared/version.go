package shared

import (
	_ "embed"
	"strings"
)

//go:embed version.txt
var rocketPoolVersion string

func RocketPoolVersion() string {
	return strings.TrimSpace(rocketPoolVersion)
}

//go:embed logo.txt
var logo string

func Logo() string {
	return strings.TrimSpace(logo)
}
