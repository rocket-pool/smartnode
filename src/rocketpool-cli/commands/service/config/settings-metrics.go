package config

import (
	"github.com/rivo/tview"
	"github.com/rocket-pool/node-manager-core/config"
	snCfg "github.com/rocket-pool/smartnode/v2/shared/config"
)

// The page wrapper for the metrics config
type MetricsConfigPage struct {
	home                       *settingsHome
	page                       *page
	layout                     *standardLayout
	masterConfig               *snCfg.SmartNodeConfig
	enableMetricsBox           *parameterizedFormItem
	enableOdaoMetricsBox       *parameterizedFormItem
	ecMetricsPortBox           *parameterizedFormItem
	bnMetricsPortBox           *parameterizedFormItem
	vcMetricsPortBox           *parameterizedFormItem
	nodeMetricsPortBox         *parameterizedFormItem
	exporterMetricsPortBox     *parameterizedFormItem
	grafanaItems               []*parameterizedFormItem
	prometheusItems            []*parameterizedFormItem
	exporterItems              []*parameterizedFormItem
	enableBitflyNodeMetricsBox *parameterizedFormItem
	bitflyNodeMetricsItems     []*parameterizedFormItem
}

// Creates a new page for the metrics / stats settings
func NewMetricsConfigPage(home *settingsHome) *MetricsConfigPage {
	configPage := &MetricsConfigPage{
		home:         home,
		masterConfig: home.md.Config,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-metrics",
		"Monitoring / Metrics",
		"Select this to configure the monitoring and statistics gathering parts of the Smart Node, such as Grafana and Prometheus.",
		configPage.layout.grid,
	)

	return configPage
}

// Get the underlying page
func (configPage *MetricsConfigPage) getPage() *page {
	return configPage.page
}

// Creates the content for the monitoring / stats settings page
func (configPage *MetricsConfigPage) createContent() {
	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Network, "Monitoring / Metrics Settings")
	configPage.layout.setupEscapeReturnHomeHandler(configPage.home.md, configPage.home.homePage)

	// Set up the form items
	configPage.enableMetricsBox = createParameterizedCheckbox(&configPage.masterConfig.Metrics.EnableMetrics)
	configPage.enableOdaoMetricsBox = createParameterizedCheckbox(&configPage.masterConfig.Metrics.EnableOdaoMetrics)
	configPage.ecMetricsPortBox = createParameterizedUint16Field(&configPage.masterConfig.Metrics.EcMetricsPort)
	configPage.bnMetricsPortBox = createParameterizedUint16Field(&configPage.masterConfig.Metrics.BnMetricsPort)
	configPage.vcMetricsPortBox = createParameterizedUint16Field(&configPage.masterConfig.ValidatorClient.VcCommon.MetricsPort)
	configPage.nodeMetricsPortBox = createParameterizedUint16Field(&configPage.masterConfig.Metrics.DaemonMetricsPort)
	configPage.exporterMetricsPortBox = createParameterizedUint16Field(&configPage.masterConfig.Metrics.ExporterMetricsPort)
	configPage.grafanaItems = createParameterizedFormItems(configPage.masterConfig.Metrics.Grafana.GetParameters(), configPage.layout.descriptionBox)
	configPage.prometheusItems = createParameterizedFormItems(configPage.masterConfig.Metrics.Prometheus.GetParameters(), configPage.layout.descriptionBox)
	configPage.exporterItems = createParameterizedFormItems(configPage.masterConfig.Metrics.Exporter.GetParameters(), configPage.layout.descriptionBox)
	configPage.enableBitflyNodeMetricsBox = createParameterizedCheckbox(&configPage.masterConfig.Metrics.EnableBitflyNodeMetrics)
	configPage.bitflyNodeMetricsItems = createParameterizedFormItems(configPage.masterConfig.Metrics.BitflyNodeMetrics.GetParameters(), configPage.layout.descriptionBox)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.enableMetricsBox, configPage.enableOdaoMetricsBox, configPage.ecMetricsPortBox, configPage.bnMetricsPortBox, configPage.vcMetricsPortBox, configPage.nodeMetricsPortBox, configPage.exporterMetricsPortBox)
	configPage.layout.mapParameterizedFormItems(configPage.grafanaItems...)
	configPage.layout.mapParameterizedFormItems(configPage.prometheusItems...)
	configPage.layout.mapParameterizedFormItems(configPage.exporterItems...)
	configPage.layout.mapParameterizedFormItems(configPage.enableBitflyNodeMetricsBox)
	configPage.layout.mapParameterizedFormItems(configPage.bitflyNodeMetricsItems...)

	// Set up the setting callbacks
	configPage.enableMetricsBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if configPage.masterConfig.Metrics.EnableMetrics.Value == checked {
			return
		}
		configPage.masterConfig.Metrics.EnableMetrics.Value = checked
		configPage.handleLayoutChanged()
	})
	configPage.enableBitflyNodeMetricsBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if configPage.masterConfig.Metrics.EnableBitflyNodeMetrics.Value == checked {
			return
		}
		configPage.masterConfig.Metrics.EnableBitflyNodeMetrics.Value = checked
		configPage.handleLayoutChanged()
	})

	// Do the initial draw
	configPage.handleLayoutChanged()
}

// Handle all of the form changes when the Enable Metrics box has changed
func (configPage *MetricsConfigPage) handleLayoutChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.enableMetricsBox.item)

	if configPage.masterConfig.Metrics.EnableMetrics.Value {
		configPage.layout.addFormItems([]*parameterizedFormItem{configPage.enableOdaoMetricsBox})
		if configPage.masterConfig.IsLocalMode() {
			configPage.layout.addFormItems([]*parameterizedFormItem{configPage.ecMetricsPortBox, configPage.bnMetricsPortBox})
		}
		configPage.layout.addFormItems([]*parameterizedFormItem{configPage.vcMetricsPortBox, configPage.nodeMetricsPortBox, configPage.exporterMetricsPortBox})
		configPage.layout.addFormItems(configPage.grafanaItems)
		configPage.layout.addFormItems(configPage.prometheusItems)
		configPage.layout.addFormItems(configPage.exporterItems)
	}

	if configPage.masterConfig.IsLocalMode() {
		switch configPage.masterConfig.LocalBeaconClient.BeaconNode.Value {
		case config.BeaconNode_Teku, config.BeaconNode_Lighthouse, config.BeaconNode_Lodestar:
			configPage.layout.form.AddFormItem(configPage.enableBitflyNodeMetricsBox.item)
			if configPage.masterConfig.Metrics.EnableBitflyNodeMetrics.Value {
				configPage.layout.addFormItems(configPage.bitflyNodeMetricsItems)
			}
		}
	}

	configPage.layout.refresh()
}
