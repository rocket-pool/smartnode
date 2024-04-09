package config

import (
	"github.com/rivo/tview"
	"github.com/rocket-pool/node-manager-core/config/ids"
	"github.com/rocket-pool/smartnode/v2/shared/config"
)

// The page wrapper for the fallback config
type FallbackConfigPage struct {
	home           *settingsHome
	page           *page
	layout         *standardLayout
	masterConfig   *config.SmartNodeConfig
	useFallbackBox *parameterizedFormItem
	fallbackItems  []*parameterizedFormItem
}

// Creates a new page for the fallback client settings
func NewFallbackConfigPage(home *settingsHome) *FallbackConfigPage {
	configPage := &FallbackConfigPage{
		home:         home,
		masterConfig: home.md.Config,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-fallback",
		"Fallback Clients",
		"Select this to specify a secondary, externally-managed Execution Client and Beacon Node pair. The Smart Node and Validator Client will use if your main clients ever go offline.",
		configPage.layout.grid,
	)

	return configPage
}

// Get the underlying page
func (configPage *FallbackConfigPage) getPage() *page {
	return configPage.page
}

// Creates the content for the fallback client settings page
func (configPage *FallbackConfigPage) createContent() {
	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Network, "Fallback Client Settings")
	configPage.layout.setupEscapeReturnHomeHandler(configPage.home.md, configPage.home.homePage)

	// Set up the form items
	configPage.useFallbackBox = createParameterizedCheckbox(&configPage.masterConfig.Fallback.UseFallbackClients)
	configPage.fallbackItems = createParameterizedFormItems(configPage.masterConfig.Fallback.GetParameters(), configPage.layout.descriptionBox)

	// Take the enable out since it's done explicitly
	fallbackItems := []*parameterizedFormItem{}
	for _, item := range configPage.fallbackItems {
		if item.parameter.GetCommon().ID == ids.FallbackUseFallbackClientsID {
			continue
		}
		fallbackItems = append(fallbackItems, item)
	}
	configPage.fallbackItems = fallbackItems

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.useFallbackBox)
	configPage.layout.mapParameterizedFormItems(configPage.fallbackItems...)

	// Set up the setting callbacks
	configPage.useFallbackBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if configPage.masterConfig.Fallback.UseFallbackClients.Value == checked {
			return
		}
		configPage.masterConfig.Fallback.UseFallbackClients.Value = checked
		configPage.handleUseFallbackChanged()
	})

	// Do the initial draw
	configPage.handleUseFallbackChanged()
}

// Handle all of the form changes when the Use Fallback box has changed
func (configPage *FallbackConfigPage) handleUseFallbackChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.useFallbackBox.item)

	// Only add the supporting stuff if external clients are enabled
	if !configPage.masterConfig.Fallback.UseFallbackClients.Value {
		return
	}

	configPage.layout.addFormItems(configPage.fallbackItems)
	configPage.layout.refresh()
}

// Handle a bulk redraw request
func (configPage *FallbackConfigPage) handleLayoutChanged() {
	configPage.handleUseFallbackChanged()
}
