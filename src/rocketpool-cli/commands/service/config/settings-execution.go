package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/node-manager-core/config/ids"
	snCfg "github.com/rocket-pool/smartnode/shared/config"
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
	configPage.clientModeDropdown = createParameterizedDropDown(&configPage.masterConfig.ClientMode, configPage.layout.descriptionBox)
	configPage.localEcDropdown = createParameterizedDropDown(&configPage.masterConfig.LocalExecutionConfig.ExecutionClient, configPage.layout.descriptionBox)
	configPage.externalEcDropdown = createParameterizedDropDown(&configPage.masterConfig.ExternalExecutionConfig.ExecutionClient, configPage.layout.descriptionBox)
	configPage.localEcItems = createParameterizedFormItems(configPage.masterConfig.LocalExecutionConfig.GetParameters(), configPage.layout.descriptionBox)
	configPage.gethItems = createParameterizedFormItems(configPage.masterConfig.LocalExecutionConfig.Geth.GetParameters(), configPage.layout.descriptionBox)
	configPage.nethermindItems = createParameterizedFormItems(configPage.masterConfig.LocalExecutionConfig.Nethermind.GetParameters(), configPage.layout.descriptionBox)
	configPage.besuItems = createParameterizedFormItems(configPage.masterConfig.LocalExecutionConfig.Besu.GetParameters(), configPage.layout.descriptionBox)
	configPage.externalEcItems = createParameterizedFormItems(configPage.masterConfig.ExternalExecutionConfig.GetParameters(), configPage.layout.descriptionBox)

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
		if configPage.masterConfig.LocalExecutionConfig.ExecutionClient.Value == configPage.masterConfig.LocalExecutionConfig.ExecutionClient.Options[index].Value {
			return
		}
		configPage.masterConfig.LocalExecutionConfig.ExecutionClient.Value = configPage.masterConfig.LocalExecutionConfig.ExecutionClient.Options[index].Value
		configPage.handleLocalEcChanged()
	})
	configPage.externalEcDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.masterConfig.ExternalExecutionConfig.ExecutionClient.Value == configPage.masterConfig.ExternalExecutionConfig.ExecutionClient.Options[index].Value {
			return
		}
		configPage.masterConfig.ExternalExecutionConfig.ExecutionClient.Value = configPage.masterConfig.ExternalExecutionConfig.ExecutionClient.Options[index].Value
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
	selectedEc := configPage.masterConfig.LocalExecutionConfig.ExecutionClient.Value

	switch selectedEc {
	case config.ExecutionClient_Geth:
		configPage.layout.addFormItemsWithCommonParams(configPage.localEcItems, configPage.gethItems, nil)
	case config.ExecutionClient_Nethermind:
		configPage.layout.addFormItemsWithCommonParams(configPage.localEcItems, configPage.nethermindItems, nil)
	case config.ExecutionClient_Besu:
		configPage.layout.addFormItemsWithCommonParams(configPage.localEcItems, configPage.besuItems, nil)
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
