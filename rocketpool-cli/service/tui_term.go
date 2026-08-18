package service

import (
	"fmt"
	"os"
	"strings"

	"github.com/rivo/tview"

	cliconfig "github.com/rocket-pool/smartnode/rocketpool-cli/service/config"
	snconfig "github.com/rocket-pool/smartnode/shared/services/config"
)

func unsupportedTUITermError(term, termProgram string) error {
	if term == "" {
		term = "<unset>"
	}
	reason := fmt.Sprintf("TERM=%s", term)
	if termProgram != "" {
		reason += fmt.Sprintf(" (TERM_PROGRAM=%s)", termProgram)
	}
	return fmt.Errorf("the configuration UI does not support this terminal (%s).\nRetry with: TERM=xterm-256color rocketpool service config", reason)
}

func isUnsupportedTUITerm(term, termProgram string) bool {
	if term == "" || strings.EqualFold(term, "dumb") {
		return true
	}
	if strings.Contains(strings.ToLower(term), "ghostty") {
		return true
	}
	return strings.EqualFold(termProgram, "ghostty")
}

func runConfigTUI(oldCfg, cfg *snconfig.RocketPoolConfig, isNew, isUpdate, isNative bool) (*cliconfig.MainDisplay, error) {
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")
	if isUnsupportedTUITerm(term, termProgram) {
		return nil, unsupportedTUITermError(term, termProgram)
	}

	app := tview.NewApplication()
	md := cliconfig.NewMainDisplay(app, oldCfg, cfg, isNew, isUpdate, isNative)
	if err := app.Run(); err != nil {
		return nil, unsupportedTUITermError(term, termProgram)
	}
	return md, nil
}
