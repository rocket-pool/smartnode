package config

import (
	"github.com/rivo/tview"
	"github.com/urfave/cli"
)

func ConfigureService(c *cli.Context) error {

	app := tview.NewApplication()
	newMainDisplay(app)
	err := app.Run()
	return err

}
