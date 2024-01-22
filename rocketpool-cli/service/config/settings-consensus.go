package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rocket-pool/smartnode/shared/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// The page wrapper for the EC config
type ConsensusConfigPage struct {
	home                    *settingsHome
	page                    *page
	layout                  *standardLayout
	masterConfig            *config.RocketPoolConfig
	ccModeDropdown          *parameterizedFormItem
	ccDropdown              *parameterizedFormItem
	externalCcDropdown      *parameterizedFormItem
	ccCommonItems           []*parameterizedFormItem
	lighthouseItems         []*parameterizedFormItem
	lodestarItems           []*parameterizedFormItem
	nimbusItems             []*parameterizedFormItem
	prysmItems              []*parameterizedFormItem
	tekuItems               []*parameterizedFormItem
	externalLighthouseItems []*parameterizedFormItem
	externalNimbusItems     []*parameterizedFormItem
	externalLodestarItems   []*parameterizedFormItem
	externalPrysmItems      []*parameterizedFormItem
	externalTekuItems       []*parameterizedFormItem
}

// Creates a new page for the Consensus client settings
func NewConsensusConfigPage(home *settingsHome) *ConsensusConfigPage {

	configPage := &ConsensusConfigPage{
		home:         home,
		masterConfig: home.md.Config,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-consensus",
		"Consensus Client (ETH2)",
		"Select this to choose your Consensus client (formerly called \"ETH2 client\") and configure its settings.",
		configPage.layout.grid,
	)

	return configPage

}

// Get the underlying page
func (configPage *ConsensusConfigPage) getPage() *page {
	return configPage.page
}

// Creates the content for the Consensus client settings page
func (configPage *ConsensusConfigPage) createContent() {

	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Smartnode.Network, "Consensus Client (ETH2) Settings")

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
	configPage.ccModeDropdown = createParameterizedDropDown(&configPage.masterConfig.ConsensusClientMode, configPage.layout.descriptionBox)
	configPage.ccDropdown = createParameterizedDropDown(&configPage.masterConfig.ConsensusClient, configPage.layout.descriptionBox)
	configPage.externalCcDropdown = createParameterizedDropDown(&configPage.masterConfig.ExternalConsensusClient, configPage.layout.descriptionBox)
	configPage.ccCommonItems = createParameterizedFormItems(configPage.masterConfig.ConsensusCommon.GetParameters(), configPage.layout.descriptionBox)
	configPage.lighthouseItems = createParameterizedFormItems(configPage.masterConfig.Lighthouse.GetParameters(), configPage.layout.descriptionBox)
	configPage.lodestarItems = createParameterizedFormItems(configPage.masterConfig.Lodestar.GetParameters(), configPage.layout.descriptionBox)
	configPage.nimbusItems = createParameterizedFormItems(configPage.masterConfig.Nimbus.GetParameters(), configPage.layout.descriptionBox)
	configPage.prysmItems = createParameterizedFormItems(configPage.masterConfig.Prysm.GetParameters(), configPage.layout.descriptionBox)
	configPage.tekuItems = createParameterizedFormItems(configPage.masterConfig.Teku.GetParameters(), configPage.layout.descriptionBox)
	configPage.externalLighthouseItems = createParameterizedFormItems(configPage.masterConfig.ExternalLighthouse.GetParameters(), configPage.layout.descriptionBox)
	configPage.externalNimbusItems = createParameterizedFormItems(configPage.masterConfig.ExternalNimbus.GetParameters(), configPage.layout.descriptionBox)
	configPage.externalLodestarItems = createParameterizedFormItems(configPage.masterConfig.ExternalLodestar.GetParameters(), configPage.layout.descriptionBox)
	configPage.externalPrysmItems = createParameterizedFormItems(configPage.masterConfig.ExternalPrysm.GetParameters(), configPage.layout.descriptionBox)
	configPage.externalTekuItems = createParameterizedFormItems(configPage.masterConfig.ExternalTeku.GetParameters(), configPage.layout.descriptionBox)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.ccModeDropdown, configPage.ccDropdown, configPage.externalCcDropdown)
	configPage.layout.mapParameterizedFormItems(configPage.ccCommonItems...)
	configPage.layout.mapParameterizedFormItems(configPage.lighthouseItems...)
	configPage.layout.mapParameterizedFormItems(configPage.lodestarItems...)
	configPage.layout.mapParameterizedFormItems(configPage.nimbusItems...)
	configPage.layout.mapParameterizedFormItems(configPage.prysmItems...)
	configPage.layout.mapParameterizedFormItems(configPage.tekuItems...)
	configPage.layout.mapParameterizedFormItems(configPage.externalLighthouseItems...)
	configPage.layout.mapParameterizedFormItems(configPage.externalNimbusItems...)
	configPage.layout.mapParameterizedFormItems(configPage.externalLodestarItems...)
	configPage.layout.mapParameterizedFormItems(configPage.externalPrysmItems...)
	configPage.layout.mapParameterizedFormItems(configPage.externalTekuItems...)

	// Set up the setting callbacks
	configPage.ccModeDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.masterConfig.ConsensusClientMode.Value == configPage.masterConfig.ConsensusClientMode.Options[index].Value {
			return
		}
		configPage.masterConfig.ExecutionClientMode.Value = configPage.masterConfig.ExecutionClientMode.Options[index].Value
		configPage.masterConfig.ConsensusClientMode.Value = configPage.masterConfig.ConsensusClientMode.Options[index].Value
		configPage.handleCcModeChanged()
	})
	configPage.ccDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.masterConfig.ConsensusClient.Value == configPage.masterConfig.ConsensusClient.Options[index].Value {
			return
		}
		configPage.masterConfig.ConsensusClient.Value = configPage.masterConfig.ConsensusClient.Options[index].Value
		configPage.handleLocalCcChanged()
	})
	configPage.externalCcDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.masterConfig.ExternalConsensusClient.Value == configPage.masterConfig.ExternalConsensusClient.Options[index].Value {
			return
		}
		configPage.masterConfig.ExternalConsensusClient.Value = configPage.masterConfig.ExternalConsensusClient.Options[index].Value
		configPage.handleExternalCcChanged()
	})

	// Do the initial draw
	configPage.handleCcModeChanged()

}

// Handle all of the form changes when the CC mode has changed
func (configPage *ConsensusConfigPage) handleCcModeChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.ccModeDropdown.item)

	selectedMode := configPage.masterConfig.ConsensusClientMode.Value.(cfgtypes.Mode)
	switch selectedMode {
	case cfgtypes.Mode_Local:
		// Local (Docker mode)
		configPage.handleLocalCcChanged()

	case cfgtypes.Mode_External:
		// External (Hybrid mode)
		configPage.handleExternalCcChanged()
	}
}

// Handle all of the form changes when the CC has changed (local mode)
func (configPage *ConsensusConfigPage) handleLocalCcChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.ccModeDropdown.item)
	configPage.layout.form.AddFormItem(configPage.ccDropdown.item)
	selectedCc := configPage.masterConfig.ConsensusClient.Value.(cfgtypes.ConsensusClient)

	switch selectedCc {
	case cfgtypes.ConsensusClient_Lighthouse:
		configPage.layout.addFormItemsWithCommonParams(configPage.ccCommonItems, configPage.lighthouseItems, configPage.masterConfig.Lighthouse.UnsupportedCommonParams)
	case cfgtypes.ConsensusClient_Lodestar:
		configPage.layout.addFormItemsWithCommonParams(configPage.ccCommonItems, configPage.lodestarItems, configPage.masterConfig.Lodestar.UnsupportedCommonParams)
	case cfgtypes.ConsensusClient_Nimbus:
		configPage.layout.addFormItemsWithCommonParams(configPage.ccCommonItems, configPage.nimbusItems, configPage.masterConfig.Nimbus.UnsupportedCommonParams)
	case cfgtypes.ConsensusClient_Prysm:
		configPage.layout.addFormItemsWithCommonParams(configPage.ccCommonItems, configPage.prysmItems, configPage.masterConfig.Prysm.UnsupportedCommonParams)
	case cfgtypes.ConsensusClient_Teku:
		configPage.layout.addFormItemsWithCommonParams(configPage.ccCommonItems, configPage.tekuItems, configPage.masterConfig.Teku.UnsupportedCommonParams)
	}

	configPage.layout.refresh()
}

// Handle all of the form changes when the CC has changed (external mode)
func (configPage *ConsensusConfigPage) handleExternalCcChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.ccModeDropdown.item)
	configPage.layout.form.AddFormItem(configPage.externalCcDropdown.item)
	selectedCc := configPage.masterConfig.ExternalConsensusClient.Value.(cfgtypes.ConsensusClient)

	switch selectedCc {
	case cfgtypes.ConsensusClient_Lighthouse:
		configPage.layout.addFormItems(configPage.externalLighthouseItems)
	case cfgtypes.ConsensusClient_Nimbus:
		configPage.layout.addFormItems(configPage.externalNimbusItems)
	case cfgtypes.ConsensusClient_Lodestar:
		configPage.layout.addFormItems(configPage.externalLodestarItems)
	case cfgtypes.ConsensusClient_Prysm:
		configPage.layout.addFormItems(configPage.externalPrysmItems)
	case cfgtypes.ConsensusClient_Teku:
		configPage.layout.addFormItems(configPage.externalTekuItems)
	}

	configPage.layout.refresh()
}

// Handle a bulk redraw request
func (configPage *ConsensusConfigPage) handleLayoutChanged() {
	configPage.handleCcModeChanged()
}
