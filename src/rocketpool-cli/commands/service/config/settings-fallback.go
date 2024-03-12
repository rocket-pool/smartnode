package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// The page wrapper for the fallback config
type FallbackConfigPage struct {
	home                *settingsHome
	page                *page
	layout              *standardLayout
	masterConfig        *config.RocketPoolConfig
	useFallbackBox      *parameterizedFormItem
	reconnectDelay      *parameterizedFormItem
	fallbackNormalItems []*parameterizedFormItem
	fallbackPrysmItems  []*parameterizedFormItem
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

	// Return to the home page after pressing Escape
	configPage.layout.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			// Close all dropdowns and break if one was open
			for _, param := range configPage.layout.parameters {
				dropDown, ok := param.item.(*DropDown)
				if ok && dropDown.open {
					dropDown.CloseList(configPage.home.md.app)
					return nil
				}
			}

			// Return to the home page
			configPage.home.md.setPage(configPage.home.homePage)
			return nil
		}
		return event
	})

	// Set up the form items
	configPage.useFallbackBox = createParameterizedCheckbox(&configPage.masterConfig.UseFallbackClients)
	configPage.reconnectDelay = createParameterizedStringField(&configPage.masterConfig.ReconnectDelay)
	configPage.fallbackNormalItems = createParameterizedFormItems(configPage.masterConfig.FallbackNormal.GetParameters(), configPage.layout.descriptionBox)
	configPage.fallbackPrysmItems = createParameterizedFormItems(configPage.masterConfig.FallbackPrysm.GetParameters(), configPage.layout.descriptionBox)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.useFallbackBox, configPage.reconnectDelay)
	configPage.layout.mapParameterizedFormItems(configPage.fallbackNormalItems...)
	configPage.layout.mapParameterizedFormItems(configPage.fallbackPrysmItems...)

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

	configPage.layout.refresh()
}

// Handle a bulk redraw request
func (configPage *FallbackConfigPage) handleLayoutChanged() {
	configPage.handleUseFallbackChanged()
}
