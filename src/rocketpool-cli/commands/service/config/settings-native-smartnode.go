package config

import (
	snids "github.com/rocket-pool/smartnode/v2/shared/config/ids"
)

// The page wrapper for the Smartnode config
type NativeSmartnodeConfigPage struct {
	home   *settingsNativeHome
	page   *page
	layout *standardLayout
}

// Creates a new page for the Native Smartnode settings
func NewNativeSmartnodeConfigPage(home *settingsNativeHome) *NativeSmartnodeConfigPage {
	configPage := &NativeSmartnodeConfigPage{
		home: home,
	}

	configPage.createContent()
	configPage.page = newPage(
		home.homePage,
		"settings-native-smartnode",
		"Smart Node and TX Fees",
		"Select this to configure the settings for the Smart Node itself, including the defaults and limits on transaction fees.",
		configPage.layout.grid,
	)

	return configPage
}

// Creates the content for the Smartnode settings page
func (configPage *NativeSmartnodeConfigPage) createContent() {
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
		if id == snids.ClientModeID {
			// Ignore the client mode since that's covered in the EC and BN sections
			continue
		}
		if id == snids.ProjectNameID {
			// Ignore the project name ID since it doesn't apply to native mode
			continue
		}

		layout.form.AddFormItem(formItem.item)
		layout.parameters[formItem.item] = formItem
		if id == snids.NetworkID {
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
