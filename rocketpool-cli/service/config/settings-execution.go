package config

import (
	"github.com/gdamore/tcell/v2"
)

// The page wrapper for the EC config
type ExecutionConfigPage struct {
	home   *settingsHome
	page   *page
	layout *standardLayout
}

// Creates a new page for the Execution client settings
func NewExecutionConfigPage(home *settingsHome) *ExecutionConfigPage {

	configPage := &ExecutionConfigPage{
		home: home,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-execution",
		"Execution Client (ETH1)",
		"Select this to choose your Execution client (formerly called \"ETH1 client\") and configure its settings.",
		configPage.layout.grid,
	)

	return configPage

}

// Creates the content for the Execution client settings page
func (configPage *ExecutionConfigPage) createContent() {

	// Create the layout
	masterConfig := configPage.home.md.config
	layout := newStandardLayout()
	layout.createForm(&masterConfig.Smartnode.Network, "Execution Client (Eth1) Settings")

	// Return to the home page after pressing Escape
	layout.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			configPage.home.md.setPage(configPage.home.homePage)
			return nil
		}
		return event
	})

	// Return the standard layout's grid
	configPage.layout = layout

}
