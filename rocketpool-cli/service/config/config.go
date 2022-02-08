package config

import (
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/urfave/cli"
)

func ConfigureService(c *cli.Context) error {

	app := tview.NewApplication()
	config := config.NewConfiguration()
	newMainDisplay(app, config)
	err := app.Run()
	return err

}
