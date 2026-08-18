package service

import (
	"errors"
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/terminfo"
	"github.com/rivo/tview"

	cliconfig "github.com/rocket-pool/smartnode/rocketpool-cli/service/config"
	snconfig "github.com/rocket-pool/smartnode/shared/services/config"
)

func unsupportedTUITermError(term string) error {
	if term == "" {
		term = "<unset>"
	}
	return fmt.Errorf("the configuration UI does not support this terminal (TERM=%s).\nRetry with: TERM=xterm-256color rocketpool service config", term)
}

func runConfigTUI(oldCfg, cfg *snconfig.RocketPoolConfig, isNew, isUpdate, isNative bool) (*cliconfig.MainDisplay, error) {
	term := os.Getenv("TERM")
	// tcell.LookupTerminfo discards ErrTermNotFound and falls back to infocmp,
	// which returns a raw exec.ExitError. Check the static database first.
	if _, err := terminfo.LookupTerminfo(term); errors.Is(err, tcell.ErrTermNotFound) {
		return nil, unsupportedTUITermError(term)
	}

	app := tview.NewApplication()
	md := cliconfig.NewMainDisplay(app, oldCfg, cfg, isNew, isUpdate, isNative)
	if err := app.Run(); err != nil {
		return nil, err
	}
	return md, nil
}
