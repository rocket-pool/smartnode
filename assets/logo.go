package assets

import (
	_ "embed"
)

//go:embed logo.txt
var logo string

// Logo returns the rocket pool ascii art, with padding and newlines
func Logo() string {
	return logo
}
