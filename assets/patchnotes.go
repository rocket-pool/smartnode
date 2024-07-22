package assets

import (
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
)

var patchnotesTemplateFields = struct {
	ColorReset        string
	ColorBold         string
	ColorRed          string
	ColorGreen        string
	ColorYellow       string
	ColorBlue         string
	RocketPoolVersion string
}{
	ColorReset:        terminal.ColorReset,
	ColorBold:         terminal.ColorBold,
	ColorRed:          terminal.ColorRed,
	ColorGreen:        terminal.ColorGreen,
	ColorYellow:       terminal.ColorYellow,
	ColorBlue:         terminal.ColorBlue,
	RocketPoolVersion: "",
}

//go:embed patchnotes/*.tmpl
var patchnotesFS embed.FS

func loadTemplate(version string) (*template.Template, error) {
	return template.ParseFS(patchnotesFS, fmt.Sprintf("patchnotes/%s.tmpl", version))
}

// PrintPatchNotes looks for a template in ./patchnotes/ with a name matching the current
// version of smartnode and a file extension of .tmpl
//
// If it finds one, it populates it using the PatchNotes struct defined above, and prints it after printing the logo.
// It returns an error when no template exists, or the template could not be populated.
func GetPatchNotes() (string, error) {
	version := RocketPoolVersion()
	tmpl, err := loadTemplate(version)
	if err != nil {
		return "", fmt.Errorf("unable to read patch notes: %w", err)
	}

	// Set RocketPoolVersion before executing
	patchnotesTemplateFields.RocketPoolVersion = version

	notes := new(strings.Builder)
	notes.WriteString("\n")
	notes.WriteString(Logo())
	err = tmpl.Execute(notes, patchnotesTemplateFields)
	if err != nil {
		return "", fmt.Errorf("unable to populate patch notes template: %w", err)
	}

	return notes.String(), nil
}
