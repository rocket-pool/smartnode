package config

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/types/addons"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// The page wrapper for the Graffiti Wall Writer addon config
type AddonGwwPage struct {
	addonsPage   *AddonsPage
	page         *page
	layout       *standardLayout
	masterConfig *config.RocketPoolConfig
	addon        addons.SmartnodeAddon
	enabledBox   *parameterizedFormItem
	otherParams  []*parameterizedFormItem
}

// Creates a new page for the Graffiti Wall Writer addon settings
func NewAddonGwwPage(addonsPage *AddonsPage, addon addons.SmartnodeAddon) *AddonGwwPage {

	configPage := &AddonGwwPage{
		addonsPage:   addonsPage,
		masterConfig: addonsPage.home.md.Config,
		addon:        addon,
	}
	configPage.createContent()

	configPage.page = newPage(
		addonsPage.page,
		"settings-addon-gww",
		addon.GetName(),
		addon.GetDescription(),
		configPage.layout.grid,
	)

	return configPage

}

// Get the underlying page
func (configPage *AddonGwwPage) getPage() *page {
	return configPage.page
}

// Creates the content for the GWW settings page
func (configPage *AddonGwwPage) createContent() {

	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Smartnode.Network, fmt.Sprintf("%s Settings", configPage.addon.GetName()))
	configPage.layout.setupEscapeReturnHomeHandler(configPage.addonsPage.home.md, configPage.addonsPage.page)

	// Get the parameters
	enabledParam := configPage.addon.GetEnabledParameter()
	otherParams := []*cfgtypes.Parameter{}

	for _, param := range configPage.addon.GetConfig().GetParameters() {
		if param.ID != enabledParam.ID {
			otherParams = append(otherParams, param)
		}
	}

	// Set up the form items
	configPage.enabledBox = createParameterizedCheckbox(enabledParam)
	configPage.otherParams = createParameterizedFormItems(otherParams, configPage.layout.descriptionBox)

	// Map the parameters to the form items in the layout
	configPage.layout.mapParameterizedFormItems(configPage.enabledBox)
	configPage.layout.mapParameterizedFormItems(configPage.otherParams...)

	// Set up the setting callbacks
	configPage.enabledBox.item.(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		if enabledParam.Value == checked {
			return
		}
		enabledParam.Value = checked
		configPage.handleEnableChanged()
	})

	// Do the initial draw
	configPage.handleEnableChanged()

}

// Handle all of the form changes when the Use Fallback EC box has changed
func (configPage *AddonGwwPage) handleEnableChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.enabledBox.item)

	// Only add the supporting stuff if external clients are enabled
	if configPage.addon.GetEnabledParameter().Value == false {
		return
	}
	configPage.layout.addFormItems(configPage.otherParams)
	configPage.layout.refresh()
}

// Handle a bulk redraw request
func (configPage *AddonGwwPage) handleLayoutChanged() {
	configPage.handleEnableChanged()
}
