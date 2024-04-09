package config

import (
	"github.com/rocket-pool/node-manager-core/config"
)

// The page wrapper for the native config
type NativePage struct {
	home        *settingsNativeHome
	page        *page
	layout      *standardLayout
	nativeItems []*parameterizedFormItem
}

// Creates a new page for the native settings
func NewNativePage(home *settingsNativeHome) *NativePage {
	configPage := &NativePage{
		home: home,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-native",
		"Native Mode Settings",
		"Select this to change the settings that are specific to Native mode, such as which Beacon Node you're using and the API URLs for your Execution Client and Beacon Node.",
		configPage.layout.grid,
	)

	return configPage
}

// Creates the content for the monitoring / stats settings page
func (configPage *NativePage) createContent() {
	// Create the layout
	masterConfig := configPage.home.md.Config
	layout := newStandardLayout()
	configPage.layout = layout
	configPage.layout.createForm(&masterConfig.Network, "Native Mode Settings")
	configPage.layout.setupEscapeReturnHomeHandler(configPage.home.md, configPage.home.homePage)

	// Set up the form items
	configPage.nativeItems = createParameterizedFormItems([]config.IParameter{
		&masterConfig.ExternalExecutionClient.ExecutionClient,
		&masterConfig.ExternalExecutionClient.HttpUrl,
		&masterConfig.ExternalBeaconClient.BeaconNode,
		&masterConfig.ExternalBeaconClient.HttpUrl,
		&masterConfig.ValidatorClient.NativeValidatorRestartCommand,
		&masterConfig.ValidatorClient.NativeValidatorStopCommand,
	}, configPage.layout.descriptionBox)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.nativeItems...)
	configPage.layout.addFormItems(configPage.nativeItems)

	// Do the initial draw
	configPage.layout.refresh()
}
