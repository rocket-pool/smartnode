package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// The page wrapper for the metrics config
type MetricsConfigPage struct {
	home             *settingsHome
	page             *page
	layout           *standardLayout
	masterConfig     *config.RocketPoolConfig
	enableMetricsBox *parameterizedFormItem
	grafanaItems     []*parameterizedFormItem
	prometheusItems  []*parameterizedFormItem
	exporterItems    []*parameterizedFormItem
}

// Creates a new page for the metrics / stats settings
func NewMetricsConfigPage(home *settingsHome) *MetricsConfigPage {

	configPage := &MetricsConfigPage{
		home:         home,
		masterConfig: home.md.config,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-metrics",
		"Monitoring / Metrics",
		"Select this to configure the monitoring and statistics gathering parts of the Smartnode, such as Grafana and Prometheus.",
		configPage.layout.grid,
	)

	return configPage

}

// Creates the content for the monitoring / stats settings page
func (configPage *MetricsConfigPage) createContent() {

	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Smartnode.Network, "Monitoring / Metrics Settings")

	// Return to the home page after pressing Escape
	configPage.layout.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			configPage.home.md.setPage(configPage.home.homePage)
			return nil
		}
		return event
	})

	// Set up the form items
	configPage.enableMetricsBox = createParameterizedCheckbox(&configPage.masterConfig.EnableMetrics)
	configPage.grafanaItems = createParameterizedFormItems(configPage.masterConfig.Grafana.GetParameters(), configPage.layout.descriptionBox)
	configPage.prometheusItems = createParameterizedFormItems(configPage.masterConfig.Prometheus.GetParameters(), configPage.layout.descriptionBox)
	configPage.exporterItems = createParameterizedFormItems(configPage.masterConfig.Exporter.GetParameters(), configPage.layout.descriptionBox)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.enableMetricsBox)
	configPage.layout.mapParameterizedFormItems(configPage.grafanaItems...)
	configPage.layout.mapParameterizedFormItems(configPage.prometheusItems...)
	configPage.layout.mapParameterizedFormItems(configPage.exporterItems...)

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
func (configPage *MetricsConfigPage) handleEnableMetricsChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.enableMetricsBox.item)

	// Only add the supporting stuff if metrics are enabled
	if configPage.masterConfig.EnableMetrics.Value == false {
		return
	}

	configPage.layout.addFormItems(configPage.grafanaItems)
	configPage.layout.addFormItems(configPage.prometheusItems)
	configPage.layout.addFormItems(configPage.exporterItems)

	configPage.layout.refresh()
}
