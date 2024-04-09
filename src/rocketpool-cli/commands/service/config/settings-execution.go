package config

import (
	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/node-manager-core/config/ids"
	snCfg "github.com/rocket-pool/smartnode/v2/shared/config"
)

// The page wrapper for the EC config
type ExecutionConfigPage struct {
	home               *settingsHome
	page               *page
	layout             *standardLayout
	masterConfig       *snCfg.SmartNodeConfig
	clientModeDropdown *parameterizedFormItem
	localEcDropdown    *parameterizedFormItem
	externalEcDropdown *parameterizedFormItem
	localEcItems       []*parameterizedFormItem
	gethItems          []*parameterizedFormItem
	nethermindItems    []*parameterizedFormItem
	besuItems          []*parameterizedFormItem
	rethItems          []*parameterizedFormItem
	externalEcItems    []*parameterizedFormItem
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
		"Execution Client",
		"Select this to choose your Execution Client and configure its settings.",
		configPage.layout.grid,
	)

	return configPage
}

// Get the underlying page
func (configPage *ExecutionConfigPage) getPage() *page {
	return configPage.page
}

// Creates the content for the Execution client settings page
func (configPage *ExecutionConfigPage) createContent() {
	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Network, "Execution Client Settings")
	configPage.layout.setupEscapeReturnHomeHandler(configPage.home.md, configPage.home.homePage)

	// Set up the form items
	configPage.clientModeDropdown = createParameterizedDropDown(&configPage.masterConfig.ClientMode, configPage.layout.descriptionBox)
	configPage.localEcDropdown = createParameterizedDropDown(&configPage.masterConfig.LocalExecutionClient.ExecutionClient, configPage.layout.descriptionBox)
	configPage.externalEcDropdown = createParameterizedDropDown(&configPage.masterConfig.ExternalExecutionClient.ExecutionClient, configPage.layout.descriptionBox)
	configPage.localEcItems = createParameterizedFormItems(configPage.masterConfig.LocalExecutionClient.GetParameters(), configPage.layout.descriptionBox)
	configPage.gethItems = createParameterizedFormItems(configPage.masterConfig.LocalExecutionClient.Geth.GetParameters(), configPage.layout.descriptionBox)
	configPage.nethermindItems = createParameterizedFormItems(configPage.masterConfig.LocalExecutionClient.Nethermind.GetParameters(), configPage.layout.descriptionBox)
	configPage.besuItems = createParameterizedFormItems(configPage.masterConfig.LocalExecutionClient.Besu.GetParameters(), configPage.layout.descriptionBox)
	configPage.rethItems = createParameterizedFormItems(configPage.masterConfig.LocalExecutionClient.Reth.GetParameters(), configPage.layout.descriptionBox)
	configPage.externalEcItems = createParameterizedFormItems(configPage.masterConfig.ExternalExecutionClient.GetParameters(), configPage.layout.descriptionBox)

	// Take the client selections out since they're done explicitly
	localEcItems := []*parameterizedFormItem{}
	for _, item := range configPage.localEcItems {
		if item.parameter.GetCommon().ID == ids.EcID {
			continue
		}
		localEcItems = append(localEcItems, item)
	}
	configPage.localEcItems = localEcItems

	externalEcItems := []*parameterizedFormItem{}
	for _, item := range configPage.externalEcItems {
		if item.parameter.GetCommon().ID == ids.EcID {
			continue
		}
		externalEcItems = append(externalEcItems, item)
	}
	configPage.externalEcItems = externalEcItems

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.clientModeDropdown, configPage.localEcDropdown, configPage.externalEcDropdown)
	configPage.layout.mapParameterizedFormItems(configPage.localEcItems...)
	configPage.layout.mapParameterizedFormItems(configPage.gethItems...)
	configPage.layout.mapParameterizedFormItems(configPage.nethermindItems...)
	configPage.layout.mapParameterizedFormItems(configPage.besuItems...)
	configPage.layout.mapParameterizedFormItems(configPage.rethItems...)
	configPage.layout.mapParameterizedFormItems(configPage.externalEcItems...)

	// Set up the setting callbacks
	configPage.clientModeDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.masterConfig.ClientMode.Value == configPage.masterConfig.ClientMode.Options[index].Value {
			return
		}
		configPage.masterConfig.ClientMode.Value = configPage.masterConfig.ClientMode.Options[index].Value
		configPage.handleEcModeChanged()
	})
	configPage.localEcDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.masterConfig.LocalExecutionClient.ExecutionClient.Value == configPage.masterConfig.LocalExecutionClient.ExecutionClient.Options[index].Value {
			return
		}
		configPage.masterConfig.LocalExecutionClient.ExecutionClient.Value = configPage.masterConfig.LocalExecutionClient.ExecutionClient.Options[index].Value
		configPage.handleLocalEcChanged()
	})
	configPage.externalEcDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.masterConfig.ExternalExecutionClient.ExecutionClient.Value == configPage.masterConfig.ExternalExecutionClient.ExecutionClient.Options[index].Value {
			return
		}
		configPage.masterConfig.ExternalExecutionClient.ExecutionClient.Value = configPage.masterConfig.ExternalExecutionClient.ExecutionClient.Options[index].Value
		configPage.handleExternalEcChanged()
	})

	// Do the initial draw
	configPage.handleEcModeChanged()
}

// Handle all of the form changes when the EC mode has changed
func (configPage *ExecutionConfigPage) handleEcModeChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.clientModeDropdown.item)

	selectedMode := configPage.masterConfig.ClientMode.Value
	switch selectedMode {
	case config.ClientMode_Local:
		// Local (Docker mode)
		configPage.handleLocalEcChanged()

	case config.ClientMode_External:
		// External (Hybrid mode)
		configPage.handleExternalEcChanged()
	}
}

// Handle all of the form changes when the local EC has changed
func (configPage *ExecutionConfigPage) handleLocalEcChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.clientModeDropdown.item)
	configPage.layout.form.AddFormItem(configPage.localEcDropdown.item)
	selectedEc := configPage.masterConfig.LocalExecutionClient.ExecutionClient.Value

	switch selectedEc {
	case config.ExecutionClient_Geth:
		configPage.layout.addFormItemsWithCommonParams(configPage.localEcItems, configPage.gethItems, nil)
	case config.ExecutionClient_Nethermind:
		configPage.layout.addFormItemsWithCommonParams(configPage.localEcItems, configPage.nethermindItems, nil)
	case config.ExecutionClient_Besu:
		configPage.layout.addFormItemsWithCommonParams(configPage.localEcItems, configPage.besuItems, nil)
	case config.ExecutionClient_Reth:
		configPage.layout.addFormItemsWithCommonParams(configPage.localEcItems, configPage.rethItems, nil)
	}

	configPage.layout.refresh()
}

// Handle all of the form changes when the external EC has changed
func (configPage *ExecutionConfigPage) handleExternalEcChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.clientModeDropdown.item)
	configPage.layout.form.AddFormItem(configPage.externalEcDropdown.item)
	for _, param := range configPage.externalEcItems {
		configPage.layout.form.AddFormItem(param.item)
	}
	configPage.layout.refresh()
}

// Handle a bulk redraw request
func (configPage *ExecutionConfigPage) handleLayoutChanged() {
	configPage.handleEcModeChanged()
}
