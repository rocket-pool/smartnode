package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rocket-pool/smartnode/shared/config"
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

	// Return to the home page after pressing Escape
	configPage.layout.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Return to the home page
		if event.Key() == tcell.KeyEsc {
			// Close all dropdowns and break if one was open
			for _, param := range configPage.layout.parameters {
				dropDown, ok := param.item.(*DropDown)
				if ok && dropDown.open {
					dropDown.CloseList(configPage.home.md.app)
					return nil
				}
			}

			configPage.home.md.setPage(configPage.home.homePage)
			return nil
		}
		return event
	})

	// Set up the form items
	configPage.nativeItems = createParameterizedFormItems(configPage.masterConfig.Native.GetParameters(), configPage.layout.descriptionBox)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.nativeItems...)
	configPage.layout.addFormItems(configPage.nativeItems)

	// Do the initial draw
	configPage.layout.refresh()
}
