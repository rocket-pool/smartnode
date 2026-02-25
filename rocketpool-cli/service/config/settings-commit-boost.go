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
	flashbotsBox            *parameterizedFormItem
	bloxrouteMaxProfitBox   *parameterizedFormItem
	bloxrouteRegulatedBox   *parameterizedFormItem
	titanRegionalBox        *parameterizedFormItem
	ultrasoundFilteredBox   *parameterizedFormItem
	btcsOfacBox             *parameterizedFormItem
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
		"Select this to configure the settings for the Commit-Boost PBS client, the source of blocks with MEV rewards for your validators.\n\nCommit-Boost is a powerful PBS client that offers detailed reporting and transparency around block building, very low latency, and a responsive development team.\n\n",
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
	configPage.selectionModeBox = createParameterizedDropDown(&configPage.masterConfig.CommitBoost.RelaySelectionMode, configPage.layout.descriptionBox)

	localParams := []*cfgtypes.Parameter{
		&configPage.masterConfig.CommitBoost.CustomRelays,
		&configPage.masterConfig.CommitBoost.Port,
		&configPage.masterConfig.CommitBoost.OpenRpcPort,
		&configPage.masterConfig.CommitBoost.ContainerTag,
	}
	externalParams := []*cfgtypes.Parameter{&configPage.masterConfig.CommitBoost.ExternalUrl}

	configPage.localItems = createParameterizedFormItems(localParams, configPage.layout.descriptionBox)
	configPage.externalItems = createParameterizedFormItems(externalParams, configPage.layout.descriptionBox)

	// Relay checkboxes - using CommitBoost's own relay parameters
	configPage.flashbotsBox = createParameterizedCheckbox(&configPage.masterConfig.CommitBoost.FlashbotsRelay)
	configPage.bloxrouteMaxProfitBox = createParameterizedCheckbox(&configPage.masterConfig.CommitBoost.BloxRouteMaxProfitRelay)
	configPage.bloxrouteRegulatedBox = createParameterizedCheckbox(&configPage.masterConfig.CommitBoost.BloxRouteRegulatedRelay)
	configPage.titanRegionalBox = createParameterizedCheckbox(&configPage.masterConfig.CommitBoost.TitanRegionalRelay)
	configPage.ultrasoundFilteredBox = createParameterizedCheckbox(&configPage.masterConfig.CommitBoost.UltrasoundFilteredRelay)
	configPage.btcsOfacBox = createParameterizedCheckbox(&configPage.masterConfig.CommitBoost.BtcsOfacRelay)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.enableBox, configPage.modeBox, configPage.selectionModeBox)
	configPage.layout.mapParameterizedFormItems(configPage.flashbotsBox, configPage.bloxrouteMaxProfitBox, configPage.bloxrouteRegulatedBox, configPage.titanRegionalBox, configPage.ultrasoundFilteredBox, configPage.btcsOfacBox)
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
	configPage.selectionModeBox.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.masterConfig.CommitBoost.RelaySelectionMode.Value == configPage.masterConfig.CommitBoost.RelaySelectionMode.Options[index].Value {
			return
		}
		configPage.masterConfig.CommitBoost.RelaySelectionMode.Value = configPage.masterConfig.CommitBoost.RelaySelectionMode.Options[index].Value
		configPage.handleSelectionModeChanged()
	})

	// Do the initial draw
	configPage.handleLayoutChanged()
}

// Handle all of the form changes when the Commit-Boost mode has changed
func (configPage *CommitBoostConfigPage) handleModeChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.enableBox.item)
	if configPage.masterConfig.EnableCommitBoost.Value == true {
		configPage.layout.form.AddFormItem(configPage.modeBox.item)

		var selectedMode cfgtypes.Mode
		if configPage.masterConfig.CommitBoost.Mode.Value != nil {
			selectedMode = configPage.masterConfig.CommitBoost.Mode.Value.(cfgtypes.Mode)
		} else {
			selectedMode = cfgtypes.Mode_Local
		}
		switch selectedMode {
		case cfgtypes.Mode_Local:
			configPage.handleSelectionModeChanged()
		case cfgtypes.Mode_External:
			var execMode cfgtypes.Mode
			if configPage.masterConfig.ExecutionClientMode.Value != nil {
				execMode = configPage.masterConfig.ExecutionClientMode.Value.(cfgtypes.Mode)
			} else {
				execMode = cfgtypes.Mode_Local
			}
			if execMode == cfgtypes.Mode_Local {
				// Only show these to Docker users, not Hybrid users
				configPage.layout.addFormItems(configPage.externalItems)
			}
		}
	}

	configPage.layout.refresh()

	// Show the external mode warning after refresh() so it doesn't get overwritten
	// by the form's ChangedFunc callback triggered during SetFocus(0) inside refresh()
	if configPage.masterConfig.EnableCommitBoost.Value == true {
		if mode, ok := configPage.masterConfig.CommitBoost.Mode.Value.(cfgtypes.Mode); ok && mode == cfgtypes.Mode_External {
			configPage.layout.descriptionBox.SetText("[orange]NOTE: You have externally-managed client mode selected and Commit-Boost enabled. You must have Commit-Boost enabled in your externally-managed Beacon Node's configuration for this to function properly - otherwise you may not be able to publish blocks and will miss significant rewards!")
		}
	}
}

// Handle all of the form changes when the relay selection mode has changed
func (configPage *CommitBoostConfigPage) handleSelectionModeChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.enableBox.item)
	configPage.layout.form.AddFormItem(configPage.modeBox.item)

	configPage.layout.form.AddFormItem(configPage.selectionModeBox.item)
	selectedMode := configPage.masterConfig.CommitBoost.RelaySelectionMode.Value.(config.PbsRelaySelectionMode)
	switch selectedMode {
	case config.PbsRelaySelectionMode_All:
		// "Use All Relays" mode - no individual checkboxes needed

	case config.PbsRelaySelectionMode_Manual:
		// Show available relay checkboxes for the current network
		availableRelays := configPage.masterConfig.CommitBoost.GetAvailableRelays()
		for _, relay := range availableRelays {
			switch relay.ID {
			case cfgtypes.MevRelayID_Flashbots:
				configPage.layout.form.AddFormItem(configPage.flashbotsBox.item)
			case cfgtypes.MevRelayID_BloxrouteMaxProfit:
				configPage.layout.form.AddFormItem(configPage.bloxrouteMaxProfitBox.item)
			case cfgtypes.MevRelayID_BloxrouteRegulated:
				configPage.layout.form.AddFormItem(configPage.bloxrouteRegulatedBox.item)
			case cfgtypes.MevRelayID_TitanRegional:
				configPage.layout.form.AddFormItem(configPage.titanRegionalBox.item)
			case cfgtypes.MevRelayID_UltrasoundFiltered:
				configPage.layout.form.AddFormItem(configPage.ultrasoundFilteredBox.item)
			case cfgtypes.MevRelayID_BTCSOfac:
				configPage.layout.form.AddFormItem(configPage.btcsOfacBox.item)
			}
		}
	}

	// Show local settings (custom relays, port, container tag, etc.)
	configPage.layout.addFormItems(configPage.localItems)

	configPage.layout.refresh()
}

// Handle a bulk redraw request
func (configPage *CommitBoostConfigPage) handleLayoutChanged() {

	// Rebuild the parameter maps based on the selected network
	configPage.handleModeChanged()
}
