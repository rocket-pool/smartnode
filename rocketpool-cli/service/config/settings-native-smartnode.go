package config

import (
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
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
		"Smartnode and TX Fees",
		"Select this to configure the settings for the Smartnode itself, including the defaults and limits on transaction fees.",
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
	layout.createForm(&masterConfig.Smartnode.Network, "Smartnode and TX Fee Settings")
	layout.setupEscapeReturnHomeHandler(configPage.home.md, configPage.home.homePage)

	// Set up the form items
	formItems := createParameterizedFormItems(masterConfig.Smartnode.GetParameters(), layout.descriptionBox)
	for _, formItem := range formItems {
		if formItem.parameter.ID == config.ProjectNameID {
			// Ignore the project name ID since it doesn't apply to native mode
			continue
		}

		layout.form.AddFormItem(formItem.item)
		layout.parameters[formItem.item] = formItem
		if formItem.parameter.ID == config.NetworkID {
			dropDown := formItem.item.(*DropDown)
			dropDown.SetSelectedFunc(func(text string, index int) {
				newNetwork := configPage.home.md.Config.Smartnode.Network.Options[index].Value.(cfgtypes.Network)
				configPage.home.md.Config.ChangeNetwork(newNetwork)
				configPage.home.refresh()
			})
		}
	}
	layout.refresh()

}
