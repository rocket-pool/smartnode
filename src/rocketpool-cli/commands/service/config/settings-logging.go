package config

import snCfg "github.com/rocket-pool/smartnode/v2/shared/config"

// The page wrapper for the logging config
type LoggingConfigPage struct {
	home         *settingsHome
	page         *page
	layout       *standardLayout
	masterConfig *snCfg.SmartNodeConfig
	loggingItems []*parameterizedFormItem
}

// Creates a new page for the logging settings
func NewLoggingConfigPage(home *settingsHome) *LoggingConfigPage {
	configPage := &LoggingConfigPage{
		home:         home,
		masterConfig: home.md.Config,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-logging",
		"Logging",
		"Configure the Smart Node's daemon logs.",
		configPage.layout.grid,
	)

	return configPage
}

// Get the underlying page
func (configPage *LoggingConfigPage) getPage() *page {
	return configPage.page
}

// Creates the content for the logging settings page
func (configPage *LoggingConfigPage) createContent() {
	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Network, "Logging Settings")
	configPage.layout.setupEscapeReturnHomeHandler(configPage.home.md, configPage.home.homePage)

	// Set up the form items
	configPage.loggingItems = createParameterizedFormItems(configPage.masterConfig.Logging.GetParameters(), configPage.layout.descriptionBox)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.loggingItems...)

	// Do the initial draw
	configPage.handleLayoutChanged()
}

// Handle a bulk redraw request
func (configPage *LoggingConfigPage) handleLayoutChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.addFormItems(configPage.loggingItems)
	configPage.layout.refresh()
}
