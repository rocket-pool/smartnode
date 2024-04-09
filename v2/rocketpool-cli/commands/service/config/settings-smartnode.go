package config

import (
	"github.com/rocket-pool/smartnode/v2/shared/config/ids"
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
		"Smart Node and TX Fees",
		"Select this to configure the settings for the Smart Node itself, including the defaults and limits on transaction fees.",
		configPage.layout.grid,
	)

	return configPage
}

// Get the underlying page
func (configPage *SmartnodeConfigPage) getPage() *page {
	return configPage.page
}

// Creates the content for the Smartnode settings page
func (configPage *SmartnodeConfigPage) createContent() {

	// Create the layout
	masterConfig := configPage.home.md.Config
	layout := newStandardLayout()
	configPage.layout = layout
	layout.createForm(&masterConfig.Network, "Smart Node and TX Fee Settings")
	layout.setupEscapeReturnHomeHandler(configPage.home.md, configPage.home.homePage)

	// Set up the form items
	formItems := createParameterizedFormItems(masterConfig.GetParameters(), layout.descriptionBox)
	for _, formItem := range formItems {
		id := formItem.parameter.GetCommon().ID
		if id == ids.ClientModeID {
			// Ignore the client mode since that's covered in the EC and BN sections
			continue
		}

		layout.form.AddFormItem(formItem.item)
		layout.parameters[formItem.item] = formItem
		if id == ids.NetworkID {
			dropDown := formItem.item.(*DropDown)
			dropDown.SetSelectedFunc(func(text string, index int) {
				newNetwork := configPage.home.md.Config.Network.Options[index].Value
				configPage.home.md.Config.ChangeNetwork(newNetwork)
				configPage.home.refresh()
			})
		}
	}
	layout.refresh()

}

// Handle a bulk redraw request
func (configPage *SmartnodeConfigPage) handleLayoutChanged() {
	configPage.layout.refresh()
}
