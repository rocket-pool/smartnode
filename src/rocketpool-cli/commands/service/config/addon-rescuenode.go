package config

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rocket-pool/smartnode/shared/config"
	"github.com/rocket-pool/smartnode/shared/types/addons"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
)

// The page wrapper for the Rescue Node addon config
type AddonRescueNodePage struct {
	addonsPage   *AddonsPage
	page         *page
	layout       *standardLayout
	masterConfig *config.RocketPoolConfig
	addon        addons.SmartnodeAddon
	enabledBox   *parameterizedFormItem
	otherParams  []*parameterizedFormItem
}

// Creates a new page for the Graffiti Wall Writer addon settings
func NewAddonRescueNodePage(addonsPage *AddonsPage, addon addons.SmartnodeAddon) *AddonRescueNodePage {

	configPage := &AddonRescueNodePage{
		addonsPage:   addonsPage,
		masterConfig: addonsPage.home.md.Config,
		addon:        addon,
	}
	configPage.createContent()

	configPage.page = newPage(
		addonsPage.page,
		"settings-addon-rescue-node",
		addon.GetName(),
		addon.GetDescription(),
		configPage.layout.grid,
	)

	return configPage
}

// Get the underlying page
func (configPage *AddonRescueNodePage) getPage() *page {
	return configPage.page
}

// Creates the content for the Rescue Node settings page
func (configPage *AddonRescueNodePage) createContent() {

	// Create the layout
	configPage.layout = newStandardLayout()
	configPage.layout.createForm(&configPage.masterConfig.Smartnode.Network, fmt.Sprintf("%s Settings", configPage.addon.GetName()))

	// Return to the home page after pressing Escape
	configPage.layout.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			// Close all dropdowns and break if one was open
			for _, param := range configPage.layout.parameters {
				dropDown, ok := param.item.(*DropDown)
				if ok && dropDown.open {
					dropDown.CloseList(configPage.addonsPage.home.md.app)
					return nil
				}
			}

			// Return to the home page
			configPage.addonsPage.home.md.setPage(configPage.addonsPage.page)
			return nil
		}
		return event
	})

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

// Handle all of the form changes when the Enabled box has changed
func (configPage *AddonRescueNodePage) handleEnableChanged() {
	configPage.layout.form.Clear(true)
	configPage.layout.form.AddFormItem(configPage.enabledBox.item)

	// Only add the supporting stuff if the rescue node is enabled
	if configPage.addon.GetEnabledParameter().Value == false {
		return
	}
	configPage.layout.addFormItems(configPage.otherParams)
	configPage.layout.refresh()
}

// Handle a bulk redraw request
func (configPage *AddonRescueNodePage) handleLayoutChanged() {
	configPage.handleEnableChanged()
}
