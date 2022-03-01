package config

import (
	"github.com/gdamore/tcell/v2"
)

// The page wrapper for the Smartnode config
type SmartnodeConfigPage struct {
	home   *settingsHome
	page   *page
	layout *standardLayout
}

// Creates a new page for the Smartnode settings
func NewSmartnodeConfigPage(home *settingsHome) *SmartnodeConfigPage {

	configPage := &SmartnodeConfigPage{
		home: home,
	}

	configPage.createContent()
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
func (configPage *SmartnodeConfigPage) createContent() {

	// Create the layout
	masterConfig := configPage.home.md.Config
	layout := newStandardLayout()
	configPage.layout = layout
	layout.createForm(&masterConfig.Smartnode.Network, "Smartnode and TX Fee Settings")

	// Return to the home page after pressing Escape
	layout.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			configPage.home.md.setPage(configPage.home.homePage)
			return nil
		}
		return event
	})

	// Set up the form items
	formItems := createParameterizedFormItems(masterConfig.Smartnode.GetParameters(), layout.descriptionBox)
	for _, formItem := range formItems {
		layout.form.AddFormItem(formItem.item)
		layout.parameters[formItem.item] = formItem
	}
	layout.refresh()

}
