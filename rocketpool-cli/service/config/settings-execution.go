package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// The page wrapper for the EC config
type ExecutionConfigPage struct {
	home   *settingsHome
	page   *page
	layout *standardLayout
}

// A manager for handling changes to the EC layout
type executionClientLayoutManager struct {
	layout          *standardLayout
	masterConfig    *config.MasterConfig
	ecModeDropdown  *parameterizedFormItem
	ecDropdown      *parameterizedFormItem
	ecCommonItems   []*parameterizedFormItem
	gethItems       []*parameterizedFormItem
	infuraItems     []*parameterizedFormItem
	pocketItems     []*parameterizedFormItem
	externalEcItems []*parameterizedFormItem
}

// Creates a new page for the Execution client settings
func NewExecutionConfigPage(home *settingsHome) *ExecutionConfigPage {

	configPage := &ExecutionConfigPage{
		home: home,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-execution",
		"Execution Client (ETH1)",
		"Select this to choose your Execution client (formerly called \"ETH1 client\") and configure its settings.",
		configPage.layout.grid,
	)

	return configPage

}

// Creates the content for the Execution client settings page
func (configPage *ExecutionConfigPage) createContent() {

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
	ecModeDropdown := createParameterizedDropDown(&masterConfig.ExecutionClientMode, layout.descriptionBox)
	ecDropdown := createParameterizedDropDown(&masterConfig.ExecutionClient, layout.descriptionBox)
	ecCommonItems := createParameterizedFormItems(masterConfig.ExecutionCommon.GetParameters(), layout.descriptionBox)
	gethItems := createParameterizedFormItems(masterConfig.Geth.GetParameters(), layout.descriptionBox)
	infuraItems := createParameterizedFormItems(masterConfig.Infura.GetParameters(), layout.descriptionBox)
	pocketItems := createParameterizedFormItems(masterConfig.Pocket.GetParameters(), layout.descriptionBox)
	externalEcItems := createParameterizedFormItems(masterConfig.ExternalExecution.GetParameters(), layout.descriptionBox)

	layout.mapParameterizedFormItems(ecModeDropdown, ecDropdown)
	layout.mapParameterizedFormItems(ecCommonItems...)
	layout.mapParameterizedFormItems(gethItems...)
	layout.mapParameterizedFormItems(infuraItems...)
	layout.mapParameterizedFormItems(pocketItems...)
	layout.mapParameterizedFormItems(externalEcItems...)

	manager := &executionClientLayoutManager{
		layout:          layout,
		masterConfig:    masterConfig,
		ecModeDropdown:  ecModeDropdown,
		ecDropdown:      ecDropdown,
		ecCommonItems:   ecCommonItems,
		gethItems:       gethItems,
		infuraItems:     infuraItems,
		pocketItems:     pocketItems,
		externalEcItems: externalEcItems,
	}

	ecModeDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if masterConfig.ExecutionClientMode.Value == masterConfig.ExecutionClientMode.Options[index].Value {
			return
		}
		masterConfig.ExecutionClientMode.Value = masterConfig.ExecutionClientMode.Options[index].Value
		manager.handleEcModeChanged()
	})

	ecDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if masterConfig.ExecutionClient.Value == masterConfig.ExecutionClient.Options[index].Value {
			return
		}
		masterConfig.ExecutionClient.Value = masterConfig.ExecutionClient.Options[index].Value
		manager.handleLocalEcChanged()
	})

	manager.handleEcModeChanged()

}

// Handle all of the form changes when the EC mode has changed
func (manager *executionClientLayoutManager) handleEcModeChanged() {
	manager.layout.form.Clear(true)
	manager.layout.form.AddFormItem(manager.ecModeDropdown.item)

	selectedMode := manager.masterConfig.ExecutionClientMode.Value.(config.Mode)
	switch selectedMode {
	case config.Mode_Local:
		// Local (Docker mode)
		manager.handleLocalEcChanged()

	case config.Mode_External:
		// External (Hybrid mode)
		for _, param := range manager.externalEcItems {
			manager.layout.form.AddFormItem(param.item)
		}
		manager.layout.refresh()
	}
}

// Handle all of the form changes when the EC has changed
func (manager *executionClientLayoutManager) handleLocalEcChanged() {
	manager.layout.form.Clear(true)
	manager.layout.form.AddFormItem(manager.ecModeDropdown.item)
	manager.layout.form.AddFormItem(manager.ecDropdown.item)
	selectedEc := manager.masterConfig.ExecutionClient.Value.(config.ExecutionClient)

	switch selectedEc {
	case config.ExecutionClient_Geth:
		manager.layout.addFormItemsWithCommonParams(manager.ecCommonItems, manager.gethItems, manager.masterConfig.Geth.UnsupportedCommonParams)
	case config.ExecutionClient_Infura:
		manager.layout.addFormItemsWithCommonParams(manager.ecCommonItems, manager.infuraItems, manager.masterConfig.Infura.UnsupportedCommonParams)
	case config.ExecutionClient_Pocket:
		manager.layout.addFormItemsWithCommonParams(manager.ecCommonItems, manager.pocketItems, manager.masterConfig.Pocket.UnsupportedCommonParams)
	}

	manager.layout.refresh()
}
