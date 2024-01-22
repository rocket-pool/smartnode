package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/config"
)

// The page wrapper for the metrics config
type NativeMetricsConfigPage struct {
	home             *settingsNativeHome
	page             *page
	layout           *standardLayout
	masterConfig     *config.RocketPoolConfig
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
	configPage.layout.createForm(&configPage.masterConfig.Smartnode.Network, "Monitoring / Metrics Settings")

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
	configPage.enableMetricsBox = createParameterizedCheckbox(&configPage.masterConfig.EnableMetrics)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.enableMetricsBox)

	// Set up the setting callbacks
	configPage.enableMetricsBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if configPage.masterConfig.EnableMetrics.Value == checked {
			return
		}
		configPage.masterConfig.EnableMetrics.Value = checked
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
	if configPage.masterConfig.EnableMetrics.Value == false {
		return
	}

	configPage.layout.refresh()
}
