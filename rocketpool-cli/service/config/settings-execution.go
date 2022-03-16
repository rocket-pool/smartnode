package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

// The page wrapper for the EC config
type ExecutionConfigPage struct {
	home            *settingsHome
	page            *page
	layout          *standardLayout
	masterConfig    *config.RocketPoolConfig
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
		home:         home,
		masterConfig: home.md.Config,
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
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Smartnode.Network, "Execution Client (Eth1) Settings")

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
	configPage.ecModeDropdown = createParameterizedDropDown(&configPage.masterConfig.ExecutionClientMode, configPage.layout.descriptionBox)
	configPage.ecDropdown = createParameterizedDropDown(&configPage.masterConfig.ExecutionClient, configPage.layout.descriptionBox)
	configPage.ecCommonItems = createParameterizedFormItems(configPage.masterConfig.ExecutionCommon.GetParameters(), configPage.layout.descriptionBox)
	configPage.gethItems = createParameterizedFormItems(configPage.masterConfig.Geth.GetParameters(), configPage.layout.descriptionBox)
	configPage.infuraItems = createParameterizedFormItems(configPage.masterConfig.Infura.GetParameters(), configPage.layout.descriptionBox)
	configPage.pocketItems = createParameterizedFormItems(configPage.masterConfig.Pocket.GetParameters(), configPage.layout.descriptionBox)
	configPage.externalEcItems = createParameterizedFormItems(configPage.masterConfig.ExternalExecution.GetParameters(), configPage.layout.descriptionBox)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.ecModeDropdown, configPage.ecDropdown)
	configPage.layout.mapParameterizedFormItems(configPage.ecCommonItems...)
	configPage.layout.mapParameterizedFormItems(configPage.gethItems...)
	configPage.layout.mapParameterizedFormItems(configPage.infuraItems...)
	configPage.layout.mapParameterizedFormItems(configPage.pocketItems...)
	configPage.layout.mapParameterizedFormItems(configPage.externalEcItems...)

	// Set up the setting callbacks
	configPage.ecModeDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.masterConfig.ExecutionClientMode.Value == configPage.masterConfig.ExecutionClientMode.Options[index].Value {
			return
		}
		configPage.masterConfig.ExecutionClientMode.Value = configPage.masterConfig.ExecutionClientMode.Options[index].Value
		configPage.handleEcModeChanged()
	})
	configPage.ecDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.masterConfig.ExecutionClient.Value == configPage.masterConfig.ExecutionClient.Options[index].Value {
			return
		}
		configPage.masterConfig.ExecutionClient.Value = configPage.masterConfig.ExecutionClient.Options[index].Value
		configPage.handleLocalEcChanged()
	})

	// Do the initial draw
	configPage.handleEcModeChanged()

}

// Handle all of the form changes when the EC mode has changed
func (configPage *ExecutionConfigPage) handleEcModeChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.ecModeDropdown.item)

	selectedMode := configPage.masterConfig.ExecutionClientMode.Value.(config.Mode)
	switch selectedMode {
	case config.Mode_Local:
		// Local (Docker mode)
		configPage.handleLocalEcChanged()

	case config.Mode_External:
		// External (Hybrid mode)
		for _, param := range configPage.externalEcItems {
			configPage.layout.form.AddFormItem(param.item)
		}
		configPage.layout.refresh()
	}
}

// Handle all of the form changes when the EC has changed
func (configPage *ExecutionConfigPage) handleLocalEcChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.ecModeDropdown.item)
	configPage.layout.form.AddFormItem(configPage.ecDropdown.item)
	selectedEc := configPage.masterConfig.ExecutionClient.Value.(config.ExecutionClient)

	switch selectedEc {
	case config.ExecutionClient_Geth:
		configPage.layout.addFormItemsWithCommonParams(configPage.ecCommonItems, configPage.gethItems, configPage.masterConfig.Geth.UnsupportedCommonParams)
	case config.ExecutionClient_Infura:
		configPage.layout.addFormItemsWithCommonParams(configPage.ecCommonItems, configPage.infuraItems, configPage.masterConfig.Infura.UnsupportedCommonParams)
	case config.ExecutionClient_Pocket:
		configPage.layout.addFormItemsWithCommonParams(configPage.ecCommonItems, configPage.pocketItems, configPage.masterConfig.Pocket.UnsupportedCommonParams)
	}

	configPage.layout.refresh()
}
