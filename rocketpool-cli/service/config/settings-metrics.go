package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// The page wrapper for the metrics config
type MetricsConfigPage struct {
	home                       *settingsHome
	page                       *page
	layout                     *standardLayout
	masterConfig               *config.RocketPoolConfig
	enableMetricsBox           *parameterizedFormItem
	enableOdaoMetricsBox       *parameterizedFormItem
	ecMetricsPortBox           *parameterizedFormItem
	bnMetricsPortBox           *parameterizedFormItem
	vcMetricsPortBox           *parameterizedFormItem
	nodeMetricsPortBox         *parameterizedFormItem
	exporterMetricsPortBox     *parameterizedFormItem
	watchtowerMetricsPortBox   *parameterizedFormItem
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
		"Select this to configure the monitoring and statistics gathering parts of the Smartnode, such as Grafana and Prometheus.",
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
	configPage.enableOdaoMetricsBox = createParameterizedCheckbox(&configPage.masterConfig.EnableODaoMetrics)
	configPage.ecMetricsPortBox = createParameterizedUint16Field(&configPage.masterConfig.EcMetricsPort)
	configPage.bnMetricsPortBox = createParameterizedUint16Field(&configPage.masterConfig.BnMetricsPort)
	configPage.vcMetricsPortBox = createParameterizedUint16Field(&configPage.masterConfig.VcMetricsPort)
	configPage.nodeMetricsPortBox = createParameterizedUint16Field(&configPage.masterConfig.NodeMetricsPort)
	configPage.exporterMetricsPortBox = createParameterizedUint16Field(&configPage.masterConfig.ExporterMetricsPort)
	configPage.watchtowerMetricsPortBox = createParameterizedUint16Field(&configPage.masterConfig.WatchtowerMetricsPort)
	configPage.grafanaItems = createParameterizedFormItems(configPage.masterConfig.Grafana.GetParameters(), configPage.layout.descriptionBox)
	configPage.prometheusItems = createParameterizedFormItems(configPage.masterConfig.Prometheus.GetParameters(), configPage.layout.descriptionBox)
	configPage.exporterItems = createParameterizedFormItems(configPage.masterConfig.Exporter.GetParameters(), configPage.layout.descriptionBox)
	configPage.enableBitflyNodeMetricsBox = createParameterizedCheckbox(&configPage.masterConfig.EnableBitflyNodeMetrics)
	configPage.bitflyNodeMetricsItems = createParameterizedFormItems(configPage.masterConfig.BitflyNodeMetrics.GetParameters(), configPage.layout.descriptionBox)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.enableMetricsBox, configPage.enableOdaoMetricsBox, configPage.ecMetricsPortBox, configPage.bnMetricsPortBox, configPage.vcMetricsPortBox, configPage.nodeMetricsPortBox, configPage.exporterMetricsPortBox, configPage.watchtowerMetricsPortBox)
	configPage.layout.mapParameterizedFormItems(configPage.grafanaItems...)
	configPage.layout.mapParameterizedFormItems(configPage.prometheusItems...)
	configPage.layout.mapParameterizedFormItems(configPage.exporterItems...)
	configPage.layout.mapParameterizedFormItems(configPage.enableBitflyNodeMetricsBox)
	configPage.layout.mapParameterizedFormItems(configPage.bitflyNodeMetricsItems...)

	// Set up the setting callbacks
	configPage.enableMetricsBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if configPage.masterConfig.EnableMetrics.Value == checked {
			return
		}
		configPage.masterConfig.EnableMetrics.Value = checked
		configPage.handleLayoutChanged()
	})
	configPage.enableBitflyNodeMetricsBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if configPage.masterConfig.EnableBitflyNodeMetrics.Value == checked {
			return
		}
		configPage.masterConfig.EnableBitflyNodeMetrics.Value = checked
		configPage.handleLayoutChanged()
	})

	// Do the initial draw
	configPage.handleLayoutChanged()
}

// Handle all of the form changes when the Enable Metrics box has changed
func (configPage *MetricsConfigPage) handleLayoutChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.enableMetricsBox.item)

	if configPage.masterConfig.EnableMetrics.Value == true {
		configPage.layout.addFormItems([]*parameterizedFormItem{configPage.enableOdaoMetricsBox, configPage.ecMetricsPortBox, configPage.bnMetricsPortBox, configPage.vcMetricsPortBox, configPage.nodeMetricsPortBox, configPage.exporterMetricsPortBox, configPage.watchtowerMetricsPortBox})
		configPage.layout.addFormItems(configPage.grafanaItems)
		configPage.layout.addFormItems(configPage.prometheusItems)
		configPage.layout.addFormItems(configPage.exporterItems)
	}

	switch configPage.masterConfig.ConsensusClient.Value.(cfgtypes.ConsensusClient) {
	case cfgtypes.ConsensusClient_Teku, cfgtypes.ConsensusClient_Lighthouse, cfgtypes.ConsensusClient_Lodestar:
		configPage.layout.form.AddFormItem(configPage.enableBitflyNodeMetricsBox.item)
		if configPage.masterConfig.EnableBitflyNodeMetrics.Value == true {
			configPage.layout.addFormItems(configPage.bitflyNodeMetricsItems)
		}
	}

	configPage.layout.refresh()
}
