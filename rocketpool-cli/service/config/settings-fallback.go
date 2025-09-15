package config

import (
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// The page wrapper for the fallback config
type FallbackConfigPage struct {
	home                 *settingsHome
	page                 *page
	layout               *standardLayout
	masterConfig         *config.RocketPoolConfig
	useFallbackBox       *parameterizedFormItem
	reconnectDelay       *parameterizedFormItem
	fallbackNormalItems  []*parameterizedFormItem
	fallbackPrysmItems   []*parameterizedFormItem
	prioritizeValidation *parameterizedFormItem
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
		"Select this to manage a separate pair of externally-managed Execution and Consensus clients that the Smartnode and Validator Client will use if your main Execution or Consensus clients ever go offline.",
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
	configPage.layout.createForm(&configPage.masterConfig.Smartnode.Network, "Fallback Client Settings")
	configPage.layout.setupEscapeReturnHomeHandler(configPage.home.md, configPage.home.homePage)

	// Set up the form items
	configPage.useFallbackBox = createParameterizedCheckbox(&configPage.masterConfig.UseFallbackClients)
	configPage.reconnectDelay = createParameterizedStringField(&configPage.masterConfig.ReconnectDelay)
	configPage.fallbackNormalItems = createParameterizedFormItems(configPage.masterConfig.FallbackNormal.GetParameters(), configPage.layout.descriptionBox)
	configPage.fallbackPrysmItems = createParameterizedFormItems(configPage.masterConfig.FallbackPrysm.GetParameters(), configPage.layout.descriptionBox)
	configPage.prioritizeValidation = createParameterizedCheckbox(&configPage.masterConfig.PrioritizeValidation)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.useFallbackBox, configPage.reconnectDelay)
	configPage.layout.mapParameterizedFormItems(configPage.fallbackNormalItems...)
	configPage.layout.mapParameterizedFormItems(configPage.fallbackPrysmItems...)
	configPage.layout.mapParameterizedFormItems(configPage.prioritizeValidation)

	// Set up the setting callbacks
	configPage.useFallbackBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if configPage.masterConfig.UseFallbackClients.Value == checked {
			return
		}
		configPage.masterConfig.UseFallbackClients.Value = checked
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
	if configPage.masterConfig.UseFallbackClients.Value == false {
		return
	}
	configPage.layout.form.AddFormItem(configPage.reconnectDelay.item)

	cc, _ := configPage.masterConfig.GetSelectedConsensusClient()
	switch cc {
	case cfgtypes.ConsensusClient_Prysm:
		configPage.layout.addFormItems(configPage.fallbackPrysmItems)
	default:
		configPage.layout.addFormItems(configPage.fallbackNormalItems)
	}
	configPage.layout.form.AddFormItem(configPage.prioritizeValidation.item)

	configPage.layout.refresh()
}

// Handle a bulk redraw request
func (configPage *FallbackConfigPage) handleLayoutChanged() {
	configPage.handleUseFallbackChanged()
}
