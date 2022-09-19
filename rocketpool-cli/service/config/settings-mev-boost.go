package config

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/services/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// The page wrapper for the MEV-boost config
type MevBoostConfigPage struct {
	home              *settingsHome
	page              *page
	layout            *standardLayout
	masterConfig      *config.RocketPoolConfig
	enableBox         *parameterizedFormItem
	modeBox           *parameterizedFormItem
	localItemsMainnet []*parameterizedFormItem
	localItemsTestnet []*parameterizedFormItem
	externalItems     []*parameterizedFormItem
}

// Creates a new page for the MEV-Boost settings
func NewMevBoostConfigPage(home *settingsHome) *MevBoostConfigPage {

	configPage := &MevBoostConfigPage{
		home:         home,
		masterConfig: home.md.Config,
	}
	configPage.createContent()

	configPage.page = newPage(
		home.homePage,
		"settings-mev-boost",
		"MEV-Boost",
		"Select this to configure the settings for the Flashbots MEV-Boost client, the source of blocks with MEV rewards for your minipools.\n\nFor more information on Flashbots, MEV, and MEV-Boost, please see https://writings.flashbots.net/writings/why-run-mevboost/",
		configPage.layout.grid,
	)

	return configPage

}

// Get the underlying page
func (configPage *MevBoostConfigPage) getPage() *page {
	return configPage.page
}

// Creates the content for the MEV-Boost settings page
func (configPage *MevBoostConfigPage) createContent() {

	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Smartnode.Network, "MEV-Boost Settings")

	// Return to the home page after pressing Escape
	configPage.layout.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Return to the home page
		if event.Key() == tcell.KeyEsc {
			// Close all dropdowns and break if one was open
			for _, param := range configPage.layout.parameters {
				dropDown, ok := param.item.(*DropDown)
				if ok && dropDown.open {
					dropDown.CloseList(configPage.home.md.app)
					return nil
				}
			}

			configPage.home.md.setPage(configPage.home.homePage)
			return nil
		}
		return event
	})

	// Set up the form items
	configPage.enableBox = createParameterizedCheckbox(&configPage.masterConfig.EnableMevBoost)
	configPage.modeBox = createParameterizedDropDown(&configPage.masterConfig.MevBoost.Mode, configPage.layout.descriptionBox)

	localParamsMainnet := []*cfgtypes.Parameter{
		&configPage.masterConfig.MevBoost.FlashbotsRelay,
		&configPage.masterConfig.MevBoost.BloxRouteEthicalRelay,
		&configPage.masterConfig.MevBoost.BloxRouteMaxProfitRelay,
		&configPage.masterConfig.MevBoost.BloxRouteRegulatedRelay,
		&configPage.masterConfig.MevBoost.Port,
		&configPage.masterConfig.MevBoost.OpenRpcPort,
		&configPage.masterConfig.MevBoost.ContainerTag,
		&configPage.masterConfig.MevBoost.AdditionalFlags,
	}
	localParamsTestnet := []*cfgtypes.Parameter{
		&configPage.masterConfig.MevBoost.FlashbotsRelay,
		&configPage.masterConfig.MevBoost.BloxRouteMaxProfitRelay,
		&configPage.masterConfig.MevBoost.Port,
		&configPage.masterConfig.MevBoost.OpenRpcPort,
		&configPage.masterConfig.MevBoost.ContainerTag,
		&configPage.masterConfig.MevBoost.AdditionalFlags,
	}
	externalParams := []*cfgtypes.Parameter{&configPage.masterConfig.MevBoost.ExternalUrl}

	configPage.localItemsMainnet = createParameterizedFormItems(localParamsMainnet, configPage.layout.descriptionBox)
	configPage.localItemsTestnet = createParameterizedFormItems(localParamsTestnet, configPage.layout.descriptionBox)
	configPage.externalItems = createParameterizedFormItems(externalParams, configPage.layout.descriptionBox)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.enableBox, configPage.modeBox)
	configPage.layout.mapParameterizedFormItems(configPage.localItemsMainnet...)
	configPage.layout.mapParameterizedFormItems(configPage.localItemsTestnet...)
	configPage.layout.mapParameterizedFormItems(configPage.externalItems...)

	// Set up the setting callbacks
	configPage.enableBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if configPage.masterConfig.EnableMevBoost.Value == checked {
			return
		}
		configPage.masterConfig.EnableMevBoost.Value = checked
		configPage.handleLayoutChanged()
	})
	configPage.modeBox.item.(*DropDown).SetSelectedFunc(func(text string, index int) {
		if configPage.masterConfig.MevBoost.Mode.Value == configPage.masterConfig.MevBoost.Mode.Options[index].Value {
			return
		}
		configPage.masterConfig.MevBoost.Mode.Value = configPage.masterConfig.MevBoost.Mode.Options[index].Value
		configPage.handleModeChanged()
	})

	// Do the initial draw
	configPage.handleLayoutChanged()
}

// Handle all of the form changes when the MEV-Boost mode has changed
func (configPage *MevBoostConfigPage) handleModeChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.enableBox.item)
	if configPage.masterConfig.EnableMevBoost.Value == true {
		configPage.layout.form.AddFormItem(configPage.modeBox.item)

		selectedMode := configPage.masterConfig.MevBoost.Mode.Value.(cfgtypes.Mode)
		switch selectedMode {
		case cfgtypes.Mode_Local:
			if configPage.masterConfig.Smartnode.Network.Value.(cfgtypes.Network) == cfgtypes.Network_Mainnet {
				configPage.layout.addFormItems(configPage.localItemsMainnet)
			} else {
				configPage.layout.addFormItems(configPage.localItemsTestnet)
			}
		case cfgtypes.Mode_External:
			if configPage.masterConfig.ExecutionClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
				// Only show these to Docker users, not Hybrid users
				configPage.layout.addFormItems(configPage.externalItems)
			}
		}
	}

	configPage.layout.refresh()
}

// Handle a bulk redraw request
func (configPage *MevBoostConfigPage) handleLayoutChanged() {
	configPage.handleModeChanged()
}
