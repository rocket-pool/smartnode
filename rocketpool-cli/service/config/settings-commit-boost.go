package config

import (
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// The page wrapper for the Commit-boost config
type CommitBoostConfigPage struct {
	home                  *settingsHome
	page                  *page
	layout                *standardLayout
	masterConfig          *config.RocketPoolConfig
	enableBox             *parameterizedFormItem
	modeBox               *parameterizedFormItem
	selectionModeBox      *parameterizedFormItem
	localItems            []*parameterizedFormItem
	externalItems         []*parameterizedFormItem
	regulatedAllMevBox    *parameterizedFormItem
	unregulatedAllMevBox  *parameterizedFormItem
	flashbotsBox          *parameterizedFormItem
	bloxrouteMaxProfitBox *parameterizedFormItem
	bloxrouteRegulatedBox *parameterizedFormItem
	ultrasoundBox         *parameterizedFormItem
	ultrasoundFilteredBox *parameterizedFormItem
	aestusBox             *parameterizedFormItem
	titanGlobalBox        *parameterizedFormItem
	titanRegionalBox      *parameterizedFormItem
	btcsOfacBox           *parameterizedFormItem
}

// Creates a new page for the Commit-Boost settings
func NewCommitBoostConfigPage(home *settingsHome) *CommitBoostConfigPage {

	configPage := &CommitBoostConfigPage{
		home:         home,
		masterConfig: home.md.Config,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-commit-boost",
		"Commit-Boost",
		"Select this to configure the settings for the Commit-Boost client, the source of blocks with MEV rewards for your minipools.\n\n",
		configPage.layout.grid,
	)

	return configPage

}

// Get the underlying page
func (configPage *CommitBoostConfigPage) getPage() *page {
	return configPage.page
}

// Creates the content for the Commit-Boost settings page
func (configPage *CommitBoostConfigPage) createContent() {

	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Smartnode.Network, "Commit-Boost Settings")
	configPage.layout.setupEscapeReturnHomeHandler(configPage.home.md, configPage.home.homePage)

	// Set up the form items
	configPage.enableBox = createParameterizedCheckbox(&configPage.masterConfig.EnableCommitBoost)
	configPage.modeBox = createParameterizedDropDown(&configPage.masterConfig.CommitBoost.Mode, configPage.layout.descriptionBox)

	localParams := []*cfgtypes.Parameter{
		&configPage.masterConfig.CommitBoost.Port,
		&configPage.masterConfig.CommitBoost.OpenRpcPort,
		&configPage.masterConfig.CommitBoost.ContainerTag,
		&configPage.masterConfig.CommitBoost.AdditionalFlags,
	}
	externalParams := []*cfgtypes.Parameter{&configPage.masterConfig.CommitBoost.ExternalUrl}

	configPage.localItems = createParameterizedFormItems(localParams, configPage.layout.descriptionBox)
	configPage.externalItems = createParameterizedFormItems(externalParams, configPage.layout.descriptionBox)

	configPage.flashbotsBox = createParameterizedCheckbox(&configPage.masterConfig.MevBoost.FlashbotsRelay)
	configPage.bloxrouteMaxProfitBox = createParameterizedCheckbox(&configPage.masterConfig.MevBoost.BloxRouteMaxProfitRelay)
	configPage.bloxrouteRegulatedBox = createParameterizedCheckbox(&configPage.masterConfig.MevBoost.BloxRouteRegulatedRelay)
	configPage.ultrasoundBox = createParameterizedCheckbox(&configPage.masterConfig.MevBoost.UltrasoundRelay)
	configPage.ultrasoundFilteredBox = createParameterizedCheckbox(&configPage.masterConfig.MevBoost.UltrasoundFilteredRelay)
	configPage.aestusBox = createParameterizedCheckbox(&configPage.masterConfig.MevBoost.AestusRelay)
	configPage.titanGlobalBox = createParameterizedCheckbox(&configPage.masterConfig.MevBoost.TitanGlobalRelay)
	configPage.titanRegionalBox = createParameterizedCheckbox(&configPage.masterConfig.MevBoost.TitanRegionalRelay)
	configPage.btcsOfacBox = createParameterizedCheckbox(&configPage.masterConfig.MevBoost.BtcsOfacRelay)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.enableBox, configPage.modeBox, configPage.selectionModeBox)
	configPage.layout.mapParameterizedFormItems(configPage.flashbotsBox, configPage.bloxrouteMaxProfitBox, configPage.bloxrouteRegulatedBox, configPage.ultrasoundBox, configPage.ultrasoundFilteredBox, configPage.aestusBox, configPage.titanGlobalBox, configPage.titanRegionalBox, configPage.btcsOfacBox)
	configPage.layout.mapParameterizedFormItems(configPage.localItems...)
	configPage.layout.mapParameterizedFormItems(configPage.externalItems...)

	// Set up the setting callbacks
	configPage.enableBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if configPage.masterConfig.EnableCommitBoost.Value == checked {
			return
		}
		configPage.masterConfig.EnableCommitBoost.Value = checked
		configPage.handleLayoutChanged()
	})
	configPage.modeBox.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.masterConfig.CommitBoost.Mode.Value == configPage.masterConfig.CommitBoost.Mode.Options[index].Value {
			return
		}
		configPage.masterConfig.CommitBoost.Mode.Value = configPage.masterConfig.CommitBoost.Mode.Options[index].Value
		configPage.handleModeChanged()
	})

	// Do the initial draw
	configPage.handleLayoutChanged()
}

// Handle all of the form changes when the MEV-Boost mode has changed
func (configPage *CommitBoostConfigPage) handleModeChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.enableBox.item)
	if configPage.masterConfig.EnableCommitBoost.Value == true {
		configPage.layout.form.AddFormItem(configPage.modeBox.item)

		selectedMode := configPage.masterConfig.CommitBoost.Mode.Value.(cfgtypes.Mode)
		switch selectedMode {
		case cfgtypes.Mode_Local:
			configPage.handleSelectionModeChanged()
		case cfgtypes.Mode_External:
			if configPage.masterConfig.ExecutionClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
				// Only show these to Docker users, not Hybrid users
				configPage.layout.addFormItems(configPage.externalItems)
			}
		}
	}

	configPage.layout.refresh()
}

// Handle all of the form changes when the relay selection mode has changed
func (configPage *CommitBoostConfigPage) handleSelectionModeChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.enableBox.item)
	configPage.layout.form.AddFormItem(configPage.modeBox.item)

	configPage.layout.form.AddFormItem(configPage.selectionModeBox.item)

	configPage.layout.addFormItems(configPage.localItems)
}

// Handle a bulk redraw request
func (configPage *CommitBoostConfigPage) handleLayoutChanged() {

	// Rebuild the parameter maps based on the selected network
	configPage.handleModeChanged()
}
