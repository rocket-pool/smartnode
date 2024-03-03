package config

import (
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// The page wrapper for the native config
type NativePage struct {
	home         *settingsNativeHome
	page         *page
	layout       *standardLayout
	masterConfig *config.RocketPoolConfig
	nativeItems  []*parameterizedFormItem
}

// Creates a new page for the native settings
func NewNativePage(home *settingsNativeHome) *NativePage {

	configPage := &NativePage{
		home:         home,
		masterConfig: home.md.Config,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-native",
		"Native Mode Settings",
		"Select this to change the settings that are specific to Native mode, such as which Consensus client you're using and the API URLs for your Execution and Consensus clients.",
		configPage.layout.grid,
	)

	return configPage

}

// Creates the content for the monitoring / stats settings page
func (configPage *NativePage) createContent() {

	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Smartnode.Network, "Native Mode Settings")
	configPage.layout.setupEscapeReturnHomeHandler(configPage.home.md, configPage.home.homePage)

	// Set up the form items
	configPage.nativeItems = createParameterizedFormItems(configPage.masterConfig.Native.GetParameters(), configPage.layout.descriptionBox)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.nativeItems...)
	configPage.layout.addFormItems(configPage.nativeItems)

	// Do the initial draw
	configPage.layout.refresh()
}
