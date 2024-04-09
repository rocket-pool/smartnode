package config

import (
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/v2/shared/config"
)

// The page wrapper for the metrics config
type NativeMetricsConfigPage struct {
	home             *settingsNativeHome
	page             *page
	layout           *standardLayout
	masterConfig     *config.SmartNodeConfig
	enableMetricsBox *parameterizedFormItem
}

// Creates a new page for the metrics / stats settings
func NewNativeMetricsConfigPage(home *settingsNativeHome) *NativeMetricsConfigPage {
	configPage := &NativeMetricsConfigPage{
		home:         home,
		masterConfig: home.md.Config,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-native-metrics",
		"Monitoring / Metrics",
		"Select this to configure the monitoring functions of the Daemon.",
		configPage.layout.grid,
	)

	return configPage
}

// Creates the content for the monitoring / stats settings page
func (configPage *NativeMetricsConfigPage) createContent() {
	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Network, "Monitoring / Metrics Settings")
	configPage.layout.setupEscapeReturnHomeHandler(configPage.home.md, configPage.home.homePage)

	// Set up the form items
	configPage.enableMetricsBox = createParameterizedCheckbox(&configPage.masterConfig.Metrics.EnableMetrics)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.enableMetricsBox)

	// Set up the setting callbacks
	configPage.enableMetricsBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if configPage.masterConfig.Metrics.EnableMetrics.Value == checked {
			return
		}
		configPage.masterConfig.Metrics.EnableMetrics.Value = checked
		configPage.handleEnableMetricsChanged()
	})

	// Do the initial draw
	configPage.handleEnableMetricsChanged()
}

// Handle all of the form changes when the Enable Metrics box has changed
func (configPage *NativeMetricsConfigPage) handleEnableMetricsChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.enableMetricsBox.item)

	// Only add the supporting stuff if metrics are enabled
	if !configPage.masterConfig.Metrics.EnableMetrics.Value {
		return
	}

	configPage.layout.refresh()
}
