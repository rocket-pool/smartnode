package config

import (
	"github.com/gdamore/tcell/v2"
)

// The page wrapper for the EC config
type FallbackExecutionConfigPage struct {
	home   *settingsHome
	page   *page
	layout *standardLayout
}

// Creates a new page for the fallback Execution client settings
func NewFallbackExecutionConfigPage(home *settingsHome) *FallbackExecutionConfigPage {

	configPage := &FallbackExecutionConfigPage{
		home: home,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-execution-fallback",
		"Execution Backup (Eth1 Fallback)",
		"Select this to choose your fallback / backup Execution Client (formerly called \"ETH1 fallback client\") that the Smartnode and Beacon client will use if your main Execution client ever goes offline.",
		configPage.layout.grid,
	)

	return configPage

}

// Creates the content for the fallback Execution client settings page
func (configPage *FallbackExecutionConfigPage) createContent() {

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
