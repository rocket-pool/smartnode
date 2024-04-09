package config

import (
	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/node-manager-core/config/ids"
	snCfg "github.com/rocket-pool/smartnode/v2/shared/config"
	snids "github.com/rocket-pool/smartnode/v2/shared/config/ids"
)

// The page wrapper for the BN config
type BeaconConfigPage struct {
	home               *settingsHome
	page               *page
	layout             *standardLayout
	masterConfig       *snCfg.SmartNodeConfig
	clientModeDropdown *parameterizedFormItem
	localBnDropdown    *parameterizedFormItem
	externalBnDropdown *parameterizedFormItem
	localBnItems       []*parameterizedFormItem
	lighthouseBnItems  []*parameterizedFormItem
	lodestarBnItems    []*parameterizedFormItem
	nimbusBnItems      []*parameterizedFormItem
	prysmBnItems       []*parameterizedFormItem
	tekuBnItems        []*parameterizedFormItem
	externalBnItems    []*parameterizedFormItem
	vcCommonItems      []*parameterizedFormItem
	lighthouseVcItems  []*parameterizedFormItem
	lodestarVcItems    []*parameterizedFormItem
	nimbusVcItems      []*parameterizedFormItem
	prysmVcItems       []*parameterizedFormItem
	tekuVcItems        []*parameterizedFormItem
}

// Creates a new page for the Beacon Node settings
func NewBeaconConfigPage(home *settingsHome) *BeaconConfigPage {
	configPage := &BeaconConfigPage{
		home:         home,
		masterConfig: home.md.Config,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-beacon",
		"Beacon Node",
		"Select this to choose your Beacon Node and configure its settings.",
		configPage.layout.grid,
	)

	return configPage
}

// Get the underlying page
func (configPage *BeaconConfigPage) getPage() *page {
	return configPage.page
}

// Creates the content for the Consensus client settings page
func (configPage *BeaconConfigPage) createContent() {
	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Network, "Beacon Node Settings")
	configPage.layout.setupEscapeReturnHomeHandler(configPage.home.md, configPage.home.homePage)

	// Set up the form items
	configPage.clientModeDropdown = createParameterizedDropDown(&configPage.masterConfig.ClientMode, configPage.layout.descriptionBox)
	configPage.localBnDropdown = createParameterizedDropDown(&configPage.masterConfig.LocalBeaconClient.BeaconNode, configPage.layout.descriptionBox)
	configPage.externalBnDropdown = createParameterizedDropDown(&configPage.masterConfig.ExternalBeaconClient.BeaconNode, configPage.layout.descriptionBox)
	configPage.localBnItems = createParameterizedFormItems(configPage.masterConfig.LocalBeaconClient.GetParameters(), configPage.layout.descriptionBox)
	configPage.lighthouseBnItems = createParameterizedFormItems(configPage.masterConfig.LocalBeaconClient.Lighthouse.GetParameters(), configPage.layout.descriptionBox)
	configPage.lodestarBnItems = createParameterizedFormItems(configPage.masterConfig.LocalBeaconClient.Lodestar.GetParameters(), configPage.layout.descriptionBox)
	configPage.nimbusBnItems = createParameterizedFormItems(configPage.masterConfig.LocalBeaconClient.Nimbus.GetParameters(), configPage.layout.descriptionBox)
	configPage.prysmBnItems = createParameterizedFormItems(configPage.masterConfig.LocalBeaconClient.Prysm.GetParameters(), configPage.layout.descriptionBox)
	configPage.tekuBnItems = createParameterizedFormItems(configPage.masterConfig.LocalBeaconClient.Teku.GetParameters(), configPage.layout.descriptionBox)
	configPage.externalBnItems = createParameterizedFormItems(configPage.masterConfig.ExternalBeaconClient.GetParameters(), configPage.layout.descriptionBox)
	configPage.vcCommonItems = createParameterizedFormItems(configPage.masterConfig.ValidatorClient.VcCommon.GetParameters(), configPage.layout.descriptionBox)
	configPage.lighthouseVcItems = createParameterizedFormItems(configPage.masterConfig.ValidatorClient.Lighthouse.GetParameters(), configPage.layout.descriptionBox)
	configPage.lodestarVcItems = createParameterizedFormItems(configPage.masterConfig.ValidatorClient.Lodestar.GetParameters(), configPage.layout.descriptionBox)
	configPage.nimbusVcItems = createParameterizedFormItems(configPage.masterConfig.ValidatorClient.Nimbus.GetParameters(), configPage.layout.descriptionBox)
	configPage.prysmVcItems = createParameterizedFormItems(configPage.masterConfig.ValidatorClient.Prysm.GetParameters(), configPage.layout.descriptionBox)
	configPage.tekuVcItems = createParameterizedFormItems(configPage.masterConfig.ValidatorClient.Teku.GetParameters(), configPage.layout.descriptionBox)

	// Take out native mode stuff
	vcCommonItems := []*parameterizedFormItem{}
	for _, item := range configPage.vcCommonItems {
		id := item.parameter.GetCommon().ID
		if id == snids.NativeValidatorRestartCommandID || id == snids.NativeValidatorStopCommandID {
			continue
		}
		vcCommonItems = append(vcCommonItems, item)
	}
	configPage.vcCommonItems = vcCommonItems

	// Take the client selections out since they're done explicitly
	localBnItems := []*parameterizedFormItem{}
	for _, item := range configPage.localBnItems {
		if item.parameter.GetCommon().ID == ids.BnID {
			continue
		}
		localBnItems = append(localBnItems, item)
	}
	configPage.localBnItems = localBnItems

	externalBnItems := []*parameterizedFormItem{}
	for _, item := range configPage.externalBnItems {
		if item.parameter.GetCommon().ID == ids.BnID {
			continue
		}
		externalBnItems = append(externalBnItems, item)
	}
	configPage.externalBnItems = externalBnItems

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.clientModeDropdown, configPage.localBnDropdown, configPage.externalBnDropdown)
	configPage.layout.mapParameterizedFormItems(configPage.localBnItems...)
	configPage.layout.mapParameterizedFormItems(configPage.lighthouseBnItems...)
	configPage.layout.mapParameterizedFormItems(configPage.lodestarBnItems...)
	configPage.layout.mapParameterizedFormItems(configPage.nimbusBnItems...)
	configPage.layout.mapParameterizedFormItems(configPage.prysmBnItems...)
	configPage.layout.mapParameterizedFormItems(configPage.tekuBnItems...)
	configPage.layout.mapParameterizedFormItems(configPage.externalBnItems...)
	configPage.layout.mapParameterizedFormItems(configPage.vcCommonItems...)
	configPage.layout.mapParameterizedFormItems(configPage.lighthouseVcItems...)
	configPage.layout.mapParameterizedFormItems(configPage.lodestarVcItems...)
	configPage.layout.mapParameterizedFormItems(configPage.nimbusVcItems...)
	configPage.layout.mapParameterizedFormItems(configPage.prysmVcItems...)
	configPage.layout.mapParameterizedFormItems(configPage.tekuVcItems...)

	// Set up the setting callbacks
	configPage.clientModeDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.masterConfig.ClientMode.Value == configPage.masterConfig.ClientMode.Options[index].Value {
			return
		}
		configPage.masterConfig.ClientMode.Value = configPage.masterConfig.ClientMode.Options[index].Value
		configPage.handleClientModeChanged()
	})
	configPage.localBnDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.masterConfig.LocalBeaconClient.BeaconNode.Value == configPage.masterConfig.LocalBeaconClient.BeaconNode.Options[index].Value {
			return
		}
		configPage.masterConfig.LocalBeaconClient.BeaconNode.Value = configPage.masterConfig.LocalBeaconClient.BeaconNode.Options[index].Value
		configPage.handleLocalBnChanged()
	})
	configPage.externalBnDropdown.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.masterConfig.ExternalBeaconClient.BeaconNode.Value == configPage.masterConfig.ExternalBeaconClient.BeaconNode.Options[index].Value {
			return
		}
		configPage.masterConfig.ExternalBeaconClient.BeaconNode.Value = configPage.masterConfig.ExternalBeaconClient.BeaconNode.Options[index].Value
		configPage.handleExternalBnChanged()
	})

	// Do the initial draw
	configPage.handleClientModeChanged()
}

// Handle all of the form changes when the client mode has changed
func (configPage *BeaconConfigPage) handleClientModeChanged() {
	selectedMode := configPage.masterConfig.ClientMode.Value
	switch selectedMode {
	case config.ClientMode_Local:
		// Local (Docker mode)
		configPage.handleLocalBnChanged()

	case config.ClientMode_External:
		// External (Hybrid mode)
		configPage.handleExternalBnChanged()
	}
}

// Handle all of the form changes when the BN has changed (local mode)
func (configPage *BeaconConfigPage) handleLocalBnChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.clientModeDropdown.item)
	configPage.layout.form.AddFormItem(configPage.localBnDropdown.item)
	selectedBn := configPage.masterConfig.LocalBeaconClient.BeaconNode.Value

	switch selectedBn {
	case config.BeaconNode_Lighthouse:
		configPage.layout.addFormItemsWithCommonParams(configPage.localBnItems, configPage.lighthouseBnItems, nil)
		configPage.layout.addFormItemsWithCommonParams(configPage.vcCommonItems, configPage.lighthouseVcItems, nil)
	case config.BeaconNode_Lodestar:
		configPage.layout.addFormItemsWithCommonParams(configPage.localBnItems, configPage.lodestarBnItems, nil)
		configPage.layout.addFormItemsWithCommonParams(configPage.vcCommonItems, configPage.lodestarVcItems, nil)
	case config.BeaconNode_Nimbus:
		configPage.layout.addFormItemsWithCommonParams(configPage.localBnItems, configPage.nimbusBnItems, nil)
		configPage.layout.addFormItemsWithCommonParams(configPage.vcCommonItems, configPage.nimbusVcItems, nil)
	case config.BeaconNode_Prysm:
		configPage.layout.addFormItemsWithCommonParams(configPage.localBnItems, configPage.prysmBnItems, nil)
		configPage.layout.addFormItemsWithCommonParams(configPage.vcCommonItems, configPage.prysmVcItems, nil)
	case config.BeaconNode_Teku:
		configPage.layout.addFormItemsWithCommonParams(configPage.localBnItems, configPage.tekuBnItems, nil)
		configPage.layout.addFormItemsWithCommonParams(configPage.vcCommonItems, configPage.tekuVcItems, nil)
	}

	configPage.layout.refresh()
}

// Handle all of the form changes when the BN has changed (external mode)
func (configPage *BeaconConfigPage) handleExternalBnChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.clientModeDropdown.item)
	configPage.layout.form.AddFormItem(configPage.externalBnDropdown.item)
	selectedBn := configPage.masterConfig.ExternalBeaconClient.BeaconNode.Value

	// Split into Prysm and non-Prysm
	commonSettings := []*parameterizedFormItem{}
	prysmSettings := []*parameterizedFormItem{}
	for _, item := range configPage.externalBnItems {
		if item.parameter.GetCommon().ID == ids.PrysmRpcUrlID {
			prysmSettings = append(prysmSettings, item)
		} else {
			commonSettings = append(commonSettings, item)
		}
	}
	configPage.layout.addFormItems(commonSettings)

	switch selectedBn {
	case config.BeaconNode_Lighthouse:
		configPage.layout.addFormItemsWithCommonParams(configPage.vcCommonItems, configPage.lighthouseVcItems, nil)
	case config.BeaconNode_Lodestar:
		configPage.layout.addFormItemsWithCommonParams(configPage.vcCommonItems, configPage.lodestarVcItems, nil)
	case config.BeaconNode_Nimbus:
		configPage.layout.addFormItemsWithCommonParams(configPage.vcCommonItems, configPage.nimbusVcItems, nil)
	case config.BeaconNode_Prysm:
		configPage.layout.addFormItems(prysmSettings)
		configPage.layout.addFormItemsWithCommonParams(configPage.vcCommonItems, configPage.prysmVcItems, nil)
	case config.BeaconNode_Teku:
		configPage.layout.addFormItemsWithCommonParams(configPage.vcCommonItems, configPage.tekuVcItems, nil)
	}

	configPage.layout.refresh()
}

// Handle a bulk redraw request
func (configPage *BeaconConfigPage) handleLayoutChanged() {
	configPage.handleClientModeChanged()
}
