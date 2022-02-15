package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

type SmartnodeConfigPage struct {
	home     *settingsHome
	page     *page
	layout   *standardLayout
	paramMap map[string]tview.FormItem
}

// Creates a new page for the Smartnode settings
func NewSmartnodeConfigPage(home *settingsHome) *SmartnodeConfigPage {

	configPage := &SmartnodeConfigPage{
		home:     home,
		paramMap: map[string]tview.FormItem{},
	}

	configPage.createSettingSmartnodeContent()
	configPage.page = newPage(
		home.homePage,
		"settings-smartnode",
		"Smartnode and TX Fees",
		"Select this to configure the settings for the Smartnode itself, including the defaults and limits on transaction fees.",
		configPage.layout.grid,
	)

	return configPage

}

// Creates the content for the Smartnode settings page
func (configPage *SmartnodeConfigPage) createSettingSmartnodeContent() {

	// Create the layout
	masterConfig := configPage.home.md.config
	layout := newStandardLayout()
	layout.createFormForConfig(masterConfig.Smartnode, masterConfig.Smartnode.Network.Value.(config.Network), "Smartnode and TX Fee Settings")

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
