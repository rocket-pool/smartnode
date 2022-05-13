package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// The page wrapper for the fallback EC config
type FallbackExecutionConfigPage struct {
	home                    *settingsHome
	page                    *page
	layout                  *standardLayout
	masterConfig            *config.RocketPoolConfig
	useFallbackEcBox        *parameterizedFormItem
	reconnectDelay          *parameterizedFormItem
	fallbackEcModeDropdown  *parameterizedFormItem
	fallbackEcDropdown      *parameterizedFormItem
	fallbackEcCommonItems   []*parameterizedFormItem
	fallbackInfuraItems     []*parameterizedFormItem
	fallbackPocketItems     []*parameterizedFormItem
	fallbackExternalECItems []*parameterizedFormItem
}

// Creates a new page for the fallback Execution client settings
func NewFallbackExecutionConfigPage(home *settingsHome) *FallbackExecutionConfigPage {

	configPage := &FallbackExecutionConfigPage{
		home:         home,
		masterConfig: home.md.Config,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-execution-fallback",
		"Execution Backup (ETH1 Fallback)",
		"Select this to choose your fallback / backup Execution Client (formerly called \"ETH1 fallback client\") that the Smartnode and Beacon client will use if your main Execution client ever goes offline.",
		configPage.layout.grid,
	)

	return configPage

}

// Get the underlying page
func (configPage *FallbackExecutionConfigPage) getPage() *page {
	return configPage.page
}

// Creates the content for the fallback Execution client settings page
func (configPage *FallbackExecutionConfigPage) createContent() {

	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Smartnode.Network, "Fallback Execution Client (ETH1) Settings")

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
	configPage.useFallbackEcBox = createParameterizedCheckbox(&configPage.masterConfig.UseFallbackExecutionClient)
	configPage.reconnectDelay = createParameterizedStringField(&configPage.masterConfig.ReconnectDelay)
	configPage.fallbackEcModeDropdown = createParameterizedDropDown(&configPage.masterConfig.FallbackExecutionClientMode, configPage.layout.descriptionBox)
	configPage.fallbackEcDropdown = createParameterizedDropDown(&configPage.masterConfig.FallbackExecutionClient, configPage.layout.descriptionBox)
	configPage.fallbackEcCommonItems = createParameterizedFormItems(configPage.masterConfig.FallbackExecutionCommon.GetParameters(), configPage.layout.descriptionBox)
	configPage.fallbackInfuraItems = createParameterizedFormItems(configPage.masterConfig.FallbackInfura.GetParameters(), configPage.layout.descriptionBox)
	configPage.fallbackPocketItems = createParameterizedFormItems(configPage.masterConfig.FallbackPocket.GetParameters(), configPage.layout.descriptionBox)
	configPage.fallbackExternalECItems = createParameterizedFormItems(configPage.masterConfig.FallbackExternalExecution.GetParameters(), configPage.layout.descriptionBox)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.useFallbackEcBox, configPage.reconnectDelay, configPage.fallbackEcModeDropdown, configPage.fallbackEcDropdown)
	configPage.layout.mapParameterizedFormItems(configPage.fallbackEcCommonItems...)
	configPage.layout.mapParameterizedFormItems(configPage.fallbackInfuraItems...)
	configPage.layout.mapParameterizedFormItems(configPage.fallbackPocketItems...)
	configPage.layout.mapParameterizedFormItems(configPage.fallbackExternalECItems...)

	// Set up the setting callbacks
	configPage.useFallbackEcBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if configPage.masterConfig.UseFallbackExecutionClient.Value == checked {
			return
		}
		configPage.masterConfig.UseFallbackExecutionClient.Value = checked
		configPage.handleUseFallbackEcChanged()
	})
	configPage.fallbackEcModeDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.masterConfig.FallbackExecutionClientMode.Value == configPage.masterConfig.FallbackExecutionClientMode.Options[index].Value {
			return
		}
		configPage.masterConfig.FallbackExecutionClientMode.Value = configPage.masterConfig.FallbackExecutionClientMode.Options[index].Value
		configPage.handleFallbackEcModeChanged()
	})
	configPage.fallbackEcDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.masterConfig.FallbackExecutionClient.Value == configPage.masterConfig.FallbackExecutionClient.Options[index].Value {
			return
		}
		configPage.masterConfig.FallbackExecutionClient.Value = configPage.masterConfig.FallbackExecutionClient.Options[index].Value
		configPage.handleLocalFallbackEcChanged()
	})

	// Do the initial draw
	configPage.handleUseFallbackEcChanged()
}

// Handle all of the form changes when the Use Fallback EC box has changed
func (configPage *FallbackExecutionConfigPage) handleUseFallbackEcChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.useFallbackEcBox.item)

	// Only add the supporting stuff if external clients are enabled
	if configPage.masterConfig.UseFallbackExecutionClient.Value == false {
		return
	}
	configPage.layout.form.AddFormItem(configPage.reconnectDelay.item)
	configPage.handleFallbackEcModeChanged()
}

// Handle all of the form changes when the fallback EC mode has changed
func (configPage *FallbackExecutionConfigPage) handleFallbackEcModeChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.useFallbackEcBox.item)
	configPage.layout.form.AddFormItem(configPage.reconnectDelay.item)
	configPage.layout.form.AddFormItem(configPage.fallbackEcModeDropdown.item)

	selectedMode := configPage.masterConfig.FallbackExecutionClientMode.Value.(config.Mode)
	switch selectedMode {
	case config.Mode_Local:
		// Local (Docker mode)
		configPage.handleLocalFallbackEcChanged()

	case config.Mode_External:
		// External (Hybrid mode)
		for _, param := range configPage.fallbackExternalECItems {
			configPage.layout.form.AddFormItem(param.item)
		}
		configPage.layout.refresh()
	}
}

// Handle all of the form changes when the fallback EC has changed
func (configPage *FallbackExecutionConfigPage) handleLocalFallbackEcChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.useFallbackEcBox.item)
	configPage.layout.form.AddFormItem(configPage.reconnectDelay.item)
	configPage.layout.form.AddFormItem(configPage.fallbackEcModeDropdown.item)
	configPage.layout.form.AddFormItem(configPage.fallbackEcDropdown.item)
	selectedEc := configPage.masterConfig.FallbackExecutionClient.Value.(config.ExecutionClient)

	switch selectedEc {
	case config.ExecutionClient_Infura:
		configPage.layout.addFormItemsWithCommonParams(configPage.fallbackEcCommonItems, configPage.fallbackInfuraItems, configPage.masterConfig.FallbackInfura.UnsupportedCommonParams)
	case config.ExecutionClient_Pocket:
		configPage.layout.addFormItemsWithCommonParams(configPage.fallbackEcCommonItems, configPage.fallbackPocketItems, configPage.masterConfig.FallbackPocket.UnsupportedCommonParams)
	}

	configPage.layout.refresh()
}

// Handle a bulk redraw request
func (configPage *FallbackExecutionConfigPage) handleLayoutChanged() {
	configPage.handleUseFallbackEcChanged()
}
