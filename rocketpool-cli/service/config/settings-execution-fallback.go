package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// The page wrapper for the fallback EC config
type FallbackExecutionConfigPage struct {
	home   *settingsHome
	page   *page
	layout *standardLayout
}

// A manager for handling changes to the fallback EC layout
type fallbackExecutionClientLayoutManager struct {
	layout                  *standardLayout
	masterConfig            *config.MasterConfig
	useFallbackEcBox        *parameterizedFormItem
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
		home: home,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-execution-fallback",
		"Execution Backup (Eth1 Fallback)",
		"Select this to choose your fallback / backup Execution Client (formerly called \"ETH1 fallback client\") that the Smartnode and Beacon client will use if your main Execution client ever goes offline.",
		configPage.layout.grid,
	)

	return configPage

}

// Creates the content for the fallback Execution client settings page
func (configPage *FallbackExecutionConfigPage) createContent() {

	// Create the layout
	masterConfig := configPage.home.md.config
	layout := newStandardLayout()
	configPage.layout = layout
	layout.createForm(&masterConfig.Smartnode.Network, "Execution Client (Eth1) Settings")

	// Return to the home page after pressing Escape
	layout.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			configPage.home.md.setPage(configPage.home.homePage)
			return nil
		}
		return event
	})

	// Set up the form items
	useFallbackEcBox := createParameterizedCheckbox(&masterConfig.UseFallbackExecutionClient)
	fallbackEcModeDropdown := createParameterizedDropDown(&masterConfig.FallbackExecutionClientMode, layout.descriptionBox)
	fallbackEcDropdown := createParameterizedDropDown(&masterConfig.FallbackExecutionClient, layout.descriptionBox)
	fallbackEcCommonItems := createParameterizedFormItems(masterConfig.FallbackExecutionCommon.GetParameters(), layout.descriptionBox)
	fallbackInfuraItems := createParameterizedFormItems(masterConfig.FallbackInfura.GetParameters(), layout.descriptionBox)
	fallbackPocketItems := createParameterizedFormItems(masterConfig.FallbackPocket.GetParameters(), layout.descriptionBox)
	fallbackExternalECItems := createParameterizedFormItems(masterConfig.FallbackExternalExecution.GetParameters(), layout.descriptionBox)

	layout.mapParameterizedFormItems(useFallbackEcBox, fallbackEcModeDropdown, fallbackEcDropdown)
	layout.mapParameterizedFormItems(fallbackEcCommonItems...)
	layout.mapParameterizedFormItems(fallbackInfuraItems...)
	layout.mapParameterizedFormItems(fallbackPocketItems...)
	layout.mapParameterizedFormItems(fallbackExternalECItems...)

	manager := &fallbackExecutionClientLayoutManager{
		layout:                  layout,
		masterConfig:            masterConfig,
		useFallbackEcBox:        useFallbackEcBox,
		fallbackEcModeDropdown:  fallbackEcModeDropdown,
		fallbackEcDropdown:      fallbackEcDropdown,
		fallbackEcCommonItems:   fallbackEcCommonItems,
		fallbackInfuraItems:     fallbackInfuraItems,
		fallbackPocketItems:     fallbackPocketItems,
		fallbackExternalECItems: fallbackExternalECItems,
	}

	useFallbackEcBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if masterConfig.UseFallbackExecutionClient.Value == checked {
			return
		}
		masterConfig.UseFallbackExecutionClient.Value = checked
		manager.handleUseFallbackEcChanged()
	})

	fallbackEcModeDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if masterConfig.FallbackExecutionClientMode.Value == masterConfig.FallbackExecutionClientMode.Options[index].Value {
			return
		}
		masterConfig.FallbackExecutionClientMode.Value = masterConfig.FallbackExecutionClientMode.Options[index].Value
		manager.handleFallbackEcModeChanged()
	})

	fallbackEcDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if masterConfig.FallbackExecutionClient.Value == masterConfig.FallbackExecutionClient.Options[index].Value {
			return
		}
		masterConfig.FallbackExecutionClient.Value = masterConfig.FallbackExecutionClient.Options[index].Value
		manager.handleLocalFallbackEcChanged()
	})

	manager.handleUseFallbackEcChanged()
}

// Handle all of the form changes when the Use Fallback EC box has changed
func (manager *fallbackExecutionClientLayoutManager) handleUseFallbackEcChanged() {
	manager.layout.form.Clear(true)
	manager.layout.form.AddFormItem(manager.useFallbackEcBox.item)

	// Only add the supporting stuff if external clients are enabled
	if manager.masterConfig.UseFallbackExecutionClient.Value == false {
		return
	}
	manager.handleFallbackEcModeChanged()
}

// Handle all of the form changes when the fallback EC mode has changed
func (manager *fallbackExecutionClientLayoutManager) handleFallbackEcModeChanged() {
	manager.layout.form.Clear(true)
	manager.layout.form.AddFormItem(manager.useFallbackEcBox.item)
	manager.layout.form.AddFormItem(manager.fallbackEcModeDropdown.item)

	selectedMode := manager.masterConfig.FallbackExecutionClientMode.Value.(config.Mode)
	switch selectedMode {
	case config.Mode_Local:
		// Local (Docker mode)
		manager.handleLocalFallbackEcChanged()

	case config.Mode_External:
		// External (Hybrid mode)
		for _, param := range manager.fallbackExternalECItems {
			manager.layout.form.AddFormItem(param.item)
		}
		manager.layout.refresh()
	}
}

// Handle all of the form changes when the fallback EC has changed
func (manager *fallbackExecutionClientLayoutManager) handleLocalFallbackEcChanged() {
	manager.layout.form.Clear(true)
	manager.layout.form.AddFormItem(manager.useFallbackEcBox.item)
	manager.layout.form.AddFormItem(manager.fallbackEcModeDropdown.item)
	manager.layout.form.AddFormItem(manager.fallbackEcDropdown.item)
	selectedEc := manager.masterConfig.FallbackExecutionClient.Value.(config.ExecutionClient)

	switch selectedEc {
	case config.ExecutionClient_Infura:
		manager.layout.addFormItemsWithCommonParams(manager.fallbackEcCommonItems, manager.fallbackInfuraItems, manager.masterConfig.FallbackInfura.UnsupportedCommonParams)
	case config.ExecutionClient_Pocket:
		manager.layout.addFormItemsWithCommonParams(manager.fallbackEcCommonItems, manager.fallbackPocketItems, manager.masterConfig.FallbackPocket.UnsupportedCommonParams)
	}

	manager.layout.refresh()
}
